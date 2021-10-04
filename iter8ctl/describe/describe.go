// Package describe implements the `iter8ctl describe` subcommand.
package describe

import (
	"fmt"
	"os"
	"strings"

	iter8 "github.com/iter8-tools/etc3/api/v2beta1"
	expr "github.com/iter8-tools/etc3/iter8ctl/experiment"
	"github.com/iter8-tools/etc3/taskrunner/core"
	"github.com/olekukonko/tablewriter"
)

// Result struct contains fields that store intermediate results associated with an invocation of 'iter8ctl describe' subcommand.
type Result struct {
	experiment  *expr.Experiment
	description strings.Builder
	err         error
}

// Builder returns an initialized Cmd struct pointer.
// Builder enables the builder design pattern along with method chaining.
func Builder() *Result {
	var d = &Result{
		experiment:  nil,
		description: strings.Builder{},
		err:         nil,
	}
	return d
}

// Error returns any error generated during the invocation of Cmd methods, or nil if there are no errors.
func (d *Result) Error() error {
	return d.err
}

// WithExperiment populates the Result struct with an experiment.
func (d *Result) WithExperiment(exp *expr.Experiment) *Result {
	if d.err != nil {
		return d
	}
	d.experiment = exp
	return d
}

// FromFile populates the Result struct with an experiment from file.
func (d *Result) FromFile(path string) *Result {
	if d.err != nil {
		return d
	}
	exp, err := (&core.Builder{}).FromFile(path).Build()
	if err != nil {
		d.err = err
		return d
	}
	d.experiment = &expr.Experiment{Experiment: exp.Experiment}
	return d
}

// printProgress prints name, namespace, and target of the experiment and the number of completed iterations into d's description buffer.
func (d *Result) printProgress() *Result {
	if d.err != nil {
		return d
	}
	d.description.WriteString("\n****** Overview ******\n")
	d.description.WriteString("Experiment name: " + d.experiment.Name + "\n")
	d.description.WriteString("Experiment namespace: " + d.experiment.Namespace + "\n")
	d.description.WriteString(fmt.Sprintf("Testing pattern: %v\n", d.experiment.Status.TestingPattern))

	d.description.WriteString("\n****** Progress Summary ******\n")
	sta := d.experiment.Status
	if sta.Stage != nil {
		d.description.WriteString(fmt.Sprintf("Experiment stage: %s\n", *sta.Stage))
	}
	if sta.CompletedLoops == nil || *sta.CompletedLoops == 0 {
		d.description.WriteString("Number of completed loops: 0\n")
	} else {
		d.description.WriteString(fmt.Sprintf("Number of completed loops: %v\n", *sta.CompletedLoops))
	}
	return d
}

// printWinnerAssessment prints the winning version in the experiment into d's description buffer.
// If winner assessment is unavailable for the underlying experiment, this method will indicate likewise.
func (d *Result) printWinnerAssessment() *Result {
	if d.err != nil {
		return d
	}
	if a := d.experiment.Status.Analysis; a != nil {
		if w := a.Winner; w != nil {
			d.description.WriteString("\n****** Winner Assessment ******\n")
			var explanation string = ""
			if d.experiment.Status.TestingPattern != nil {
				switch *d.experiment.Status.TestingPattern {
				case iter8.TestingPatternSLOValidation:
					explanation = "> If just one version satisfies the obectives, it is the winner. If both versions satisify the objectives, the candidate is the winner. Otherwise, there is no winner."
				case iter8.TestingPatternAB:
				case iter8.TestingPatternABN:
					explanation = "> The version that yields the greatest reward is the winner.\n"
				case iter8.TestingPatternHybridAB:
				case iter8.TestingPatternHybridABN:
					explanation = "> The version that yields the greatest reward and satisfies the experiment objectives is the winner. If no version satisfies the objectives, there is no winner.\n"
				default:
					explanation = ""
				}
			}
			d.description.WriteString(explanation)
			if len(d.experiment.Spec.VersionInfo) != 0 {
				d.description.WriteString(fmt.Sprintf("App versions in this experiment: %s\n", d.experiment.Spec.VersionInfo))
			}
			if w.WinnerFound {
				d.description.WriteString(fmt.Sprintf("Winning version: %s\n", *w.Winner))
			} else {
				d.description.WriteString("Winning version: not found\n")
			}
		}
	}
	return d
}

// printRewardAssessment prints a matrix of values for each reward-version pair.
// Rows correspond to experiment rewards. Columns correspond to versions.
// The current "best" version for each reward is denoted with a "*".
func (d *Result) printRewardAssessment() *Result {
	if d.err != nil ||
		d.experiment.Status.Analysis == nil ||
		d.experiment.Status.Analysis.Objectives == nil ||
		d.experiment.Spec.Criteria == nil ||
		len(d.experiment.Spec.Criteria.Rewards) == 0 {
		return d
	}

	d.description.WriteString("\n****** Reward Assessment ******\n")
	d.description.WriteString("> Identifies values of reward metrics for each version. The best version is marked with a '*'.\n")
	table := tablewriter.NewWriter(&d.description)
	table.SetRowLine(true)
	versions := d.experiment.GetVersions()
	table.SetHeader(append([]string{"Reward"}, versions...))
	for _, reward := range d.experiment.Spec.Criteria.Rewards {
		row := []string{expr.StringifyReward(reward)}
		table.Append(append(row, d.experiment.GetAnnotatedMetricStrs(reward)...))
	}
	table.Render()

	return d
}

// printObjectiveAssessment prints a matrix of boolean values into d's description buffer.
// Rows correspond to experiment objectives, columns correspond to versions, and entry [i, j] indicates if objective i is satisfied by version j.
// Objective assessments are printed in the same sequence as in the experiment's spec.criteria.objectives section.
// If objective assessments are unavailable for the underlying experiment, this method will indicate likewise.
func (d *Result) printObjectiveAssessment() *Result {
	if d.err != nil {
		return d
	}
	if a := d.experiment.Status.Analysis; a != nil {
		if v := a.Objectives; v != nil {
			d.description.WriteString("\n****** Objective Assessment ******\n")
			d.description.WriteString("> Identifies whether or not the experiment objectives are satisfied by the most recently observed metrics values for each version.\n")
			table := tablewriter.NewWriter(&d.description)
			table.SetRowLine(true)
			versions := d.experiment.GetVersions()
			table.SetHeader(append([]string{"Objective"}, versions...))
			for i, objective := range d.experiment.Spec.Criteria.Objectives {
				row := []string{expr.StringifyObjective(objective)}
				table.Append(append(row, d.experiment.GetSatisfyStrs(i)...))
			}
			table.Render()
		}
	}
	return d
}

// printVersionAssessment prints how each version is performing with respect to experiment criteria into d's description buffer. This method invokes printObjectiveAssessment under the covers.
func (d *Result) printVersionAssessment() *Result {
	if d.err != nil {
		return d
	}
	if c := d.experiment.Spec.Criteria; c != nil && len(c.Objectives) > 0 {
		d.printObjectiveAssessment()
	}
	return d
}

// printMetrics prints a matrix of (decimal) metric values into d's description buffer.
// Rows correspond to experiment metrics, columns correspond to versions, and entry [i, j] indicates the value of metric i for version j.
// Metrics are printed in the same sequence as in the experiment's status.metrics section.
// If metrics are unavailable for the underlying experiment, this method will indicate likewise.
func (d *Result) printMetrics() *Result {
	if d.err != nil {
		return d
	}
	if a := d.experiment.Status.Analysis; a != nil {
		if v := a.Metrics; v != nil {
			d.description.WriteString("\n****** Metrics Assessment ******\n")
			d.description.WriteString("> Most recently read values of experiment metrics for each version.\n")
			table := tablewriter.NewWriter(&d.description)
			table.SetRowLine(true)
			versions := d.experiment.GetVersions()
			table.SetHeader(append([]string{"Metric"}, versions...))
			for _, backend := range d.experiment.Spec.Backends {
				for _, metricInfo := range backend.Metrics {
					row := []string{expr.GetMetricNameAndUnits(metricInfo)}
					table.Append(append(row, d.experiment.GetMetricStrs(metricInfo.Name)...))
				}
			}
			table.Render()
		}
	}
	return d
}

// PrintAnalysis prints the progress of the iter8 experiment, winner assessment, version assessment, and metrics.
func (d *Result) PrintAnalysis() *Result {
	if d.err != nil {
		return d
	}
	d.printProgress()
	if d.experiment.Started() {
		d.printWinnerAssessment().
			printRewardAssessment().
			printVersionAssessment().
			printMetrics()
	}
	if d.err == nil {
		fmt.Fprintln(os.Stdout, d.description.String())
	}
	return d
}
