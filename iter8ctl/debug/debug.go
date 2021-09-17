// Package debug implements the `iter8ctl debug` subcommand.
package debug

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	expr "github.com/iter8-tools/etc3/iter8ctl/experiment"
	"github.com/iter8-tools/etc3/iter8ctl/utils"
)

const (
	iter8NameSpace       string = "iter8-system"
	iter8ExpNameKey      string = "iter8/experimentName"
	iter8ExpNamespaceKey string = "iter8/experimentNamespace"
	taskRunnerSource     string = "task-runner"
)

type Iter8Log struct {
	IsIter8Log          bool   `json:"isIter8Log" yaml:"isIter8Log"`
	ExperimentName      string `json:"experimentName" yaml:"experimentName"`
	ExperimentNamespace string `json:"experimentNamespace" yaml:"experimentNamespace"`
	Source              string `json:"source" yaml:"source"`
	// ActionIndex = 0 for start action...
	// ActionIndex = number of completed loops + 1, for loop action, finish action, controller & analytcs
	ActionIndex int    `json:"actionIndex" yaml:"actionIndex"`
	Message     string `json:"message" yaml:"message"`
	Priority    uint8  `json:"priority" yaml:"priority"`
}

// byPrecedence implements sort.Interface based on the precedence of Iter8Log
type byPrecedence []Iter8Log

// Len returns length of the log slice
func (a byPrecedence) Len() int {
	return len(a)
}

// Less is true if i^th log should precede the j^th log and false otherwise
func (a byPrecedence) Less(i, j int) bool {
	if a[i].Source == a[j].Source && a[i].Source == taskRunnerSource {
		if a[i].ActionIndex < a[j].ActionIndex {
			return true
		}
		if a[i].ActionIndex == a[j].ActionIndex {
			return i < j
		}
		return false
	} else {
		panic(fmt.Sprintf("only supported source at the moment is %s", taskRunnerSource))
	}
}

// Swap two entries in the log slice
func (a byPrecedence) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// getTaskRunnerLogs gets the logs for the task runner jobs for the given experiment
func getTaskRunnerLogs(exp *expr.Experiment) ([]byte, error) {
	selector := fmt.Sprintf("%s=%s,%s=%s", iter8ExpNameKey, exp.Name, iter8ExpNamespaceKey, exp.Namespace)

	cmd := exec.Command("kubectl", "logs", "-l", selector, "-n", iter8NameSpace, "--tail=-1")
	stdout, err := cmd.CombinedOutput()

	if err != nil {
		return nil, err
	}

	return stdout, nil
}

// Debug prints iter8-logs for the given experiment
func Debug(exp *expr.Experiment, priority uint8) ([]Iter8Log, error) {
	// check priority
	if priority < 1 || priority > 3 {
		return nil, errors.New("priority can only be set to 1, 2, or 3")
	}

	// fetch task runner job logs
	tr, err := getTaskRunnerLogs(exp)
	if err != nil {
		return nil, err
	}

	// fetch controller logs
	// fetch analytics logs

	// initialize Iter8logs
	ils := []Iter8Log{}

	scanner := bufio.NewScanner(strings.NewReader(string(tr)))
	for scanner.Scan() {
		line := scanner.Text()
		if utils.IsJSONObject(line) {
			il := Iter8Log{}
			if json.Unmarshal([]byte(line), &il) == nil {
				// filter Iter8logs for this experiment
				if il.IsIter8Log &&
					il.ExperimentName == exp.Name &&
					il.ExperimentNamespace == exp.Namespace &&
					il.Priority <= priority {
					ils = append(ils, il)
				}
			}
		}

		// sort logs by precedence
		sort.Sort(byPrecedence(ils))
	}

	// return iter8-logs sorted by precedence
	return ils, nil

}
