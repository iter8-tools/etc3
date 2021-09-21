package collect

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"sync"
	"time"

	iter8 "github.com/iter8-tools/etc3/api/v2beta1"
	"github.com/iter8-tools/etc3/taskrunner/core"
	"github.com/sirupsen/logrus"
)

const (
	// TaskName is the name of the task this file implements
	TaskName string = "metrics/collect"

	// DefaultQPS is the default value of QPS (queries per sec) in collect task inputs
	DefaultQPS float32 = 8

	// DefaultTime is the default value of time (duration of queries) in collect task inputs
	DefaultTime string = "5s"
)

var log *logrus.Logger

func init() {
	log = core.GetLogger()
}

// Version contains header and url information needed to send requests to each version.
type Version struct {
	// name of the version
	// version names must be unique and must match one of the version names in the
	// VersionInfo field of the experiment
	Name string `json:"name" yaml:"name"`
	// how many queries per second will be sent to this version; optional; default 8
	QPS *float32 `json:"qps,omitempty" yaml:"qps,omitempty"`
	// HTTP headers to use in the query for this version; optional
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	// URL to use for querying this version
	URL string `json:"url" yaml:"url"`
}

// CollectInputs contain the inputs to the metrics collection task to be executed.
type CollectInputs struct {
	// how long to run the metrics collector; optional; default 5s
	Time *string `json:"time,omitempty" yaml:"time,omitempty"`
	// list of versions
	Versions []Version `json:"versions" yaml:"versions"`
	// URL of the JSON file to send during the query; optional
	PayloadURL *string `json:"payloadURL,omitempty" yaml:"payloadURL,omitempty"`
	// if LoadOnly is set to true, this task will send requests without collecting metrics; optional
	LoadOnly *bool `json:"loadOnly,omitempty" yaml:"loadOnly,omitempty"`
}

// CollectTask enables collection of Iter8's built-in metrics.
type CollectTask struct {
	core.TaskMeta
	With CollectInputs `json:"with" yaml:"with"`
}

// Make constructs a CollectTask out of a collect task spec
func Make(t *iter8.TaskSpec) (core.Task, error) {
	if *t.Task != TaskName {
		return nil, errors.New("task need to be " + TaskName)
	}
	var err error
	var jsonBytes []byte
	var bt core.Task
	// convert t to jsonBytes
	jsonBytes, err = json.Marshal(t)
	// convert jsonString to CollectTask
	if err == nil {
		ct := &CollectTask{}
		err = json.Unmarshal(jsonBytes, &ct)
		if ct.With.Versions == nil {
			return nil, errors.New("collect task with nil versions")
		}
		bt = ct
	}
	return bt, err
}

// InitializeDefaults sets default values for time duration and QPS for Fortio run
// Default values are set only if the field is non-empty
func (t *CollectTask) InitializeDefaults() {
	if t.With.Time == nil {
		t.With.Time = core.StringPointer(DefaultTime)
	}
	for i := 0; i < len(t.With.Versions); i++ {
		if t.With.Versions[i].QPS == nil {
			t.With.Versions[i].QPS = core.Float32Pointer(DefaultQPS)
		}
	}
}

////
/////////////
////

// DurationSample is a Fortio duration sample
type DurationSample struct {
	Start float64
	End   float64
	Count int
}

// DurationHist is the Fortio duration histogram
type DurationHist struct {
	Count int
	Max   float64
	Sum   float64
	Data  []DurationSample
}

// Result is the result of a single Fortio run; it contains the result for a single version
type Result struct {
	DurationHistogram DurationHist
	RetCodes          map[string]int
}

// aggregate existing results, with a new result for a specific version
func aggregate(oldResults map[string]*Result, version string, newResult *Result) map[string]*Result {
	// there are no existing results...
	if oldResults == nil {
		m := make(map[string]*Result)
		m[version] = newResult
		return m
	}
	if updatedResult, ok := oldResults[version]; ok {
		// there are existing results for the input version
		// aggregate count, max and sum
		updatedResult.DurationHistogram.Count += newResult.DurationHistogram.Count
		updatedResult.DurationHistogram.Max = math.Max(oldResults[version].DurationHistogram.Max, newResult.DurationHistogram.Max)
		updatedResult.DurationHistogram.Sum = oldResults[version].DurationHistogram.Sum + newResult.DurationHistogram.Sum

		// aggregation duration histogram data
		updatedResult.DurationHistogram.Data = append(updatedResult.DurationHistogram.Data, newResult.DurationHistogram.Data...)

		// aggregate return code counts
		if updatedResult.RetCodes == nil {
			updatedResult.RetCodes = newResult.RetCodes
		} else {
			for key := range newResult.RetCodes {
				oldResults[version].RetCodes[key] += newResult.RetCodes[key]
			}
		}
	} else {
		// there are no existing results for the input version
		oldResults[version] = newResult
	}
	// this is efficient because oldResults is a map with pointer values
	// no deep copies of structs
	return oldResults
}

// getResultFromFile reads the contents from a Fortio output file and returns it as a Result
func getResultFromFile(fortioOutputFile string) (*Result, error) {
	// open JSON file
	jsonFile, err := os.Open(fortioOutputFile)
	// if os.Open returns an error, handle it
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// defer the closing of jsonFile so that we can parse it below
	defer jsonFile.Close()

	// read jsonFile as a byte array.
	bytes, err := ioutil.ReadAll(jsonFile)
	// if ioutil.ReadAll returns an error, handle it
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// unmarshal the result and return
	var res Result
	err = json.Unmarshal(bytes, &res)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return &res, nil
}

// payloadFile downloads JSON payload from a URL into a temp file, and returns its name
func payloadFile(url string) (string, error) {
	content, err := core.GetJSONBytes(url)
	if err != nil {
		log.Error("Error while getting JSON bytes: ", err)
		return "", err
	}

	tmpfile, err := ioutil.TempFile("/tmp", "payload.json")
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	if _, err := tmpfile.Write(content); err != nil {
		tmpfile.Close()
		log.Fatal(err)
		return "", err
	}
	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
		return "", err
	}

	return tmpfile.Name(), nil
}

// resultForVersion collects Fortio result for a given version
func (t *CollectTask) resultForVersion(entry *logrus.Entry, j int, pf string) (*Result, error) {
	// the main idea is to run Fortio shell command with proper args
	// collect Fortio output as a file
	// and extract the result from the file, and return the result

	var execOut bytes.Buffer
	// appending Fortio load subcommand
	args := []string{"load"}
	// append Fortio time flag
	args = append(args, "-t", *t.With.Time)
	// append Fortio qps flag
	args = append(args, "-qps", fmt.Sprintf("%f", *t.With.Versions[j].QPS))
	// append Fortio header flags
	for header, value := range t.With.Versions[j].Headers {
		args = append(args, "-H", fmt.Sprintf("%v: %v", header, value))
	}
	// download JSON payload if specified; and append Fortio ayload-file flag
	if t.With.PayloadURL != nil {
		args = append(args, "-payload-file", pf)
	}

	// create json output file; and Fortio append json flag
	jsonOutputFile, err := ioutil.TempFile("/tmp", "output.json.")
	if err != nil {
		entry.Fatal(err)
		return nil, err
	}
	args = append(args, "-json", jsonOutputFile.Name())
	jsonOutputFile.Close()

	// append URL to be queried by Fortio
	args = append(args, t.With.Versions[j].URL)

	// setup Fortio command
	cmd := exec.Command("fortio", args...)
	cmd.Stdout = &execOut
	cmd.Stderr = os.Stderr
	entry.Trace("Invoking: " + cmd.String())

	// execute Fortio command
	err = cmd.Run()
	if err != nil {
		entry.Fatal(err)
		return nil, err
	}

	// extract result from Fortio json output file
	ifr, err := getResultFromFile(jsonOutputFile.Name())
	if err != nil {
		entry.Fatal(err)
		return nil, err
	}

	return ifr, err
}

// Run executes the metrics/collect task
// Todo: error handling
func (t *CollectTask) Run(ctx context.Context) error {
	log.Trace("collect task run started...")
	t.InitializeDefaults()
	// we would like to query versions in parallel
	// sync waitgroup is a mechanism that enables this
	var wg sync.WaitGroup

	// if experiment already has fortio data, initialize them
	// this is possible if this task is run in loop actions repeatedly
	fortioData := make(map[string]*Result)
	// MK start
	// exp, err := core.GetExperimentFromContext(ctx)
	// if err != nil {
	// 	return err
	// }
	// // if this task is **not** loadOnly
	// if t.With.LoadOnly == nil || !*t.With.LoadOnly {
	// 	// bootstrap AggregatedBuiltinHists with data already present in experiment status
	// 	if exp.Status.Analysis != nil && exp.Status.Analysis.AggregatedBuiltinHists != nil {
	// 		jsonBytes, err := json.Marshal(exp.Status.Analysis.AggregatedBuiltinHists.Data)
	// 		// convert jsonBytes to fortioData
	// 		if err == nil {
	// 			err = json.Unmarshal(jsonBytes, &fortioData)
	// 		}
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// }
	// MK end

	// lock ensures thread safety while updating fortioData from go routines
	var lock sync.Mutex

	// if errors occur in one of the parallel go routines, errCh is used to communicate them
	errCh := make(chan error)
	defer close(errCh)

	// download JSON from URL if specified
	// this is intended to be used as a JSON payload file by Fortio
	tmpfileName := ""
	if t.With.PayloadURL != nil {
		var err error
		tmpfileName, err = payloadFile(*t.With.PayloadURL)
		if err != nil {
			return err
		}
	}
	defer os.Remove(tmpfileName) // clean up later

	// execute fortio queries to versions in parallel
	for j := range t.With.Versions {
		// Increment the WaitGroup counter.
		wg.Add(1)
		// get log entry
		entry := log.WithField("version", t.With.Versions[j].Name)
		// Launch a goroutine to fetch the Fortio data for this version.
		go func(entry *logrus.Entry, k int) {
			// Decrement the counter when the goroutine completes.
			defer wg.Done()
			// Get Fortio data for version
			data, err := t.resultForVersion(entry, k, tmpfileName)
			if err == nil {
				// if this task is **not** loadOnly
				if t.With.LoadOnly == nil || !*t.With.LoadOnly {
					// Update fortioData in a threadsafe manner
					lock.Lock()
					fortioData = aggregate(fortioData, t.With.Versions[k].Name, data)
					lock.Unlock()
				}
			} else {
				// if any error occured in this go routine, send it through the error channel
				// this helps metrics/collect task exit immediately upon error
				errCh <- err
			}
		}(entry, j)
		// never use loop variable directly within the inner go routine as it will get overwritten in loop iterations
		// go func is invoked with its arg k set to the value of j
		// eliminating 'k' and simply plugging 'j' in t.With.Versions[k].Name above will not work, and will result in the ultra helpful linter warning
	}

	// See https://stackoverflow.com/questions/32840687/timeout-for-waitgroup-wait
	// Compute timeout as duration of fortio requests + 30s
	dur, err := time.ParseDuration(*t.With.Time)
	if err != nil {
		return err
	}
	// wait for WaitGroup to be done... normal execution
	// timeout ... abnormal execution
	// error on errCh ... abnormal execution
	if err = core.WaitTimeoutOrError(&wg, dur+30*time.Second, errCh); err != nil {
		log.Error("Got error: ", err)
		return err
	}
	log.Trace("Wait group finished normally")

	// if this task is **not** loadOnly
	// MK start
	// if t.With.LoadOnly == nil || !*t.With.LoadOnly {
	// 	// Update experiment status with results
	// 	// update to experiment status will result in reconcile request to etc3
	// 	// unless the task runner job executing this action is completed, this request will not have have an immediate effect in the experiment reconcilation process

	// 	bytes1, err := json.Marshal(fortioData)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	exp.SetAggregatedBuiltinHists(v1.JSON{Raw: bytes1})

	// 	core.UpdateInClusterExperimentStatus(exp)

	// 	var prettyBody bytes.Buffer
	// 	bytes2, _ := json.Marshal(exp)

	// 	json.Indent(&prettyBody, bytes2, "", "  ")
	// 	log.Trace(prettyBody.String())
	// }
	// MK end

	return err
}
