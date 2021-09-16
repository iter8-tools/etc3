// Package debug implements the `iter8ctl debug` subcommand.
package debug

import (
	"fmt"
	"os/exec"

	expr "github.com/iter8-tools/etc3/iter8ctl/experiment"
	"github.com/iter8-tools/etc3/taskrunner/core"
)

const (
	iter8NameSpace       string = "iter8-system"
	iter8ExpNameKey      string = "iter8/experimentName"
	iter8ExpNamespaceKey string = "iter8/experimentNamespace"
)

// getTaskRunnerLogs gets the logs for the task runner jobs for the given experiment
func getTaskRunnerLogs(exp *expr.Experiment) (*string, error) {
	selector := fmt.Sprintf("%s=%s,%s=%s", iter8ExpNameKey, exp.Name, iter8ExpNamespaceKey, exp.Namespace)

	cmd := exec.Command("kubectl", "logs", "-l", selector, "-n", iter8NameSpace, "--tail=-1")
	stdout, err := cmd.CombinedOutput()

	if err != nil {
		return nil, err
	}

	return core.StringPointer(string(stdout)), nil
}

// Debug prints iter8-logs for the given experiment
func Debug(exp *expr.Experiment) (*string, error) {
	// fetch task runner job logs

	tr, err := getTaskRunnerLogs(exp)
	if err != nil {
		return nil, err
	}
	return tr, nil // this needs to be fixed to return properly formatted iter8-logs

	// fetch controller logs
	// fetch analytics logs

	// select log lines that are JSON
	// filter log lines that are valid iter8-log jsons.
	// compute & add precedence value to iter8-logs
	// return iter8-logs sorted by precedence

}
