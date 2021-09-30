package experiment

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/etc3/iter8ctl/utils"
	tasks "github.com/iter8-tools/etc3/taskrunner/core"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

// getExp is a helper function for extracting an experiment object from experiment filenamePrefix
// filePath is relative to testdata folder
func getExp(filenamePrefix string) (*Experiment, error) {
	experimentFilepath := utils.CompletePath("../", fmt.Sprintf("testdata/%s.yaml", filenamePrefix))
	expBytes, err := ioutil.ReadFile(experimentFilepath)
	if err != nil {
		return nil, err
	}

	exp := &Experiment{}
	err = yaml.Unmarshal(expBytes, exp)
	if err != nil {
		return nil, err
	}
	return exp, nil
}

type test struct {
	name                   string // name of this test
	started                bool
	exp                    *Experiment
	errorRates, fakeMetric []string
	satisfyStrs, fakeObj   []string
}

var fakeValStrs = []string{"unavailable", "unavailable"}

var satisfyStrs = []string{"true", "true"}

var errorRateStrs = []string{"0.000", "0.000"}

// table driven tests
var tests = []test{
	{name: "experiment1", started: false, errorRates: []string{}, fakeMetric: []string{}, satisfyStrs: []string{}, fakeObj: []string{}},
	{name: "experiment2", started: false, errorRates: []string{}, fakeMetric: []string{}, satisfyStrs: []string{}, fakeObj: []string{}},
	{name: "experiment3", started: true, errorRates: errorRateStrs, fakeMetric: fakeValStrs, satisfyStrs: satisfyStrs, fakeObj: fakeValStrs},
	{name: "experiment4", started: true, errorRates: errorRateStrs, fakeMetric: fakeValStrs, satisfyStrs: satisfyStrs, fakeObj: fakeValStrs},
	{name: "experiment5", started: true, errorRates: errorRateStrs, fakeMetric: fakeValStrs, satisfyStrs: satisfyStrs, fakeObj: fakeValStrs},
	{name: "experiment6", started: true, errorRates: errorRateStrs, fakeMetric: fakeValStrs, satisfyStrs: satisfyStrs, fakeObj: fakeValStrs},
	{name: "experiment7", started: true, errorRates: errorRateStrs, fakeMetric: fakeValStrs, satisfyStrs: satisfyStrs, fakeObj: fakeValStrs},
	{name: "experiment8", started: true, errorRates: errorRateStrs, fakeMetric: fakeValStrs, satisfyStrs: satisfyStrs, fakeObj: fakeValStrs},
	{name: "experiment9", started: true, errorRates: errorRateStrs, fakeMetric: fakeValStrs, satisfyStrs: satisfyStrs, fakeObj: fakeValStrs},
}

func init() {
	for i := 0; i < len(tests); i++ {
		e, err := getExp(tests[i].name)
		if err == nil {
			tests[i].exp = e
		} else {
			fmt.Println("Unable to extract experiment objects from files")
			os.Exit(1)
		}
	}
}

/* Tests */

func TestExperiment(t *testing.T) {
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// test Started()
			assert.Equal(t, tc.started, tc.exp.Started())
			// test GetVersions()
			if tc.exp.Started() {
				assert.Equal(t, []string{"default", "canary"}, tc.exp.GetVersions())
			} else {
				assert.Equal(t, []string([]string(nil)), tc.exp.GetVersions())
			}
			// test GetMetricStrs(...)
			assert.Equal(t, tc.errorRates, tc.exp.GetMetricStrs("error-rate"))
			assert.Equal(t, tc.fakeMetric, tc.exp.GetMetricStrs("fake-metric"))
			// test GetSatisfyStrs()
			assert.Equal(t, tc.satisfyStrs, tc.exp.GetSatisfyStrs(0))
			assert.Equal(t, tc.fakeObj, tc.exp.GetSatisfyStrs(10))
		})
	}
}

func TestGetMetricNameAndUnits(t *testing.T) {
	metricNameAndUnits := [4]string{"95th-percentile-tail-latency (milliseconds)", "mean-latency (milliseconds)", "error-rate", "request-count"}
	mnu := [4]string{}
	for i := 0; i < 4; i++ {
		mnu[i] = GetMetricNameAndUnits(tests[2].exp.Status.Metrics[i])
	}
	assert.Equal(t, metricNameAndUnits, mnu)
}

func TestConditionFromObjective(t *testing.T) {
	objectives := [2]string{"<= 1000.000", "<= 0.010"}
	objs := [2]string{}
	for i := 0; i < 2; i++ {
		objs[i] = ConditionFromObjective(tests[2].exp.Spec.Criteria.Objectives[i])
	}
	assert.Equal(t, objectives, objs)
}

func TestStringifyReward(t *testing.T) {
	assert.Equal(t,
		"reward (lower better)",
		StringifyReward(v2alpha2.Reward{Metric: "reward", PreferredDirection: "Lower"}))
	assert.Equal(t,
		"reward (higher better)",
		StringifyReward(v2alpha2.Reward{Metric: "reward", PreferredDirection: "High"}))
}

func TestGetAnnotatedMetricStrs(t *testing.T) {
	e, err := getExp("experiment12")
	assert.NoError(t, err)
	assert.Equal(t,
		[]string{"5.030", "24.454 *"},
		e.GetAnnotatedMetricStrs(v2alpha2.Reward{Metric: "books-purchased", PreferredDirection: "High"}))
}

func TestAssertComplete(t *testing.T) {
	exp := v2alpha2.NewExperiment("test", "test").WithCondition(
		v2alpha2.ExperimentConditionExperimentCompleted,
		corev1.ConditionTrue,
		"experiment is over",
		"",
	).Build()

	err := (&Experiment{
		*exp,
	}).Assert([]ConditionType{Completed})

	assert.NoError(t, err)
}

func TestAssertInComplete(t *testing.T) {
	exp := v2alpha2.NewExperiment("test", "test").WithCondition(
		v2alpha2.ExperimentConditionExperimentCompleted,
		corev1.ConditionFalse,
		"experiment is not over",
		"",
	).Build()

	err := (&Experiment{
		*exp,
	}).Assert([]ConditionType{Completed})

	assert.Error(t, err)
}

func TestAssertWinnerFound(t *testing.T) {
	exp := v2alpha2.NewExperiment("test", "test").Build()
	exp.Status.Analysis = &v2alpha2.Analysis{}
	exp.Status.Analysis.WinnerAssessment = &v2alpha2.WinnerAssessmentAnalysis{
		Data: v2alpha2.WinnerAssessmentData{
			WinnerFound: true,
			Winner:      tasks.StringPointer("the best"),
		},
	}

	err := (&Experiment{
		*exp,
	}).Assert([]ConditionType{WinnerFound})

	assert.NoError(t, err)
}

func TestAssertNoWinnerFound(t *testing.T) {
	exp := v2alpha2.NewExperiment("test", "test").Build()
	exp.Status.Analysis = &v2alpha2.Analysis{}
	exp.Status.Analysis.WinnerAssessment = &v2alpha2.WinnerAssessmentAnalysis{
		AnalysisMetaData: v2alpha2.AnalysisMetaData{},
		Data: v2alpha2.WinnerAssessmentData{
			WinnerFound: false,
		},
	}

	err := (&Experiment{
		*exp,
	}).Assert([]ConditionType{WinnerFound})

	assert.Error(t, err)
}

func TestAssertNoWinnerFound2(t *testing.T) {
	exp := v2alpha2.NewExperiment("test", "test").Build()
	exp.Status.Analysis = &v2alpha2.Analysis{}
	exp.Status.Analysis.WinnerAssessment = nil

	err := (&Experiment{
		*exp,
	}).Assert([]ConditionType{WinnerFound})

	assert.Error(t, err)
}

func TestAssertNoWinnerFound3(t *testing.T) {
	exp := v2alpha2.NewExperiment("test", "test").Build()
	exp.Status.Analysis = nil

	err := (&Experiment{
		*exp,
	}).Assert([]ConditionType{WinnerFound})

	assert.Error(t, err)
}
