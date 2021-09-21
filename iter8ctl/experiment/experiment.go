// Package experiment enables extraction of useful information from experiment objects and their formatting.
package experiment

import (
	"context"
	"errors"
	"fmt"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	iter8 "github.com/iter8-tools/etc3/api/v2beta1"
	tasks "github.com/iter8-tools/etc3/taskrunner/core"
	"github.com/sirupsen/logrus"
	"gopkg.in/inf.v0"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var log *logrus.Logger

func init() {
	log = tasks.GetLogger()
}

// Experiment is an enhancement of iter8.Experiment struct, and supports various methods used in describing an experiment.
type Experiment struct {
	iter8.Experiment
}

// ConditionType is a type for conditions that can be asserted
type ConditionType string

const (
	// Completed implies experiment is complete
	Completed ConditionType = "completed"
	// Successful     ConditionType = "successful"
	// Failure        ConditionType = "failure"
	// HandlerFailure ConditionType = "handlerFailure"

	// WinnerFound implies experiment has found a winner
	WinnerFound ConditionType = "winnerFound"
	// CandidateWon   ConditionType = "candidateWon"
	// BaselineWon    ConditionType = "baselineWon"
	// NoWinner       ConditionType = "noWinner"
)

// for mocking in tests
var k8sClient client.Client

// GetConfig variable is useful for test mocks.
var GetConfig = func() (*rest.Config, error) {
	return config.GetConfig()
}

// GetClient constructs and returns a K8s client.
// The returned client has experiment types registered.
var GetClient = func() (rc client.Client, err error) {
	var restConf *rest.Config
	restConf, err = GetConfig()
	if err != nil {
		return nil, err
	}

	var addKnownTypes = func(scheme *runtime.Scheme) error {
		// register iter8.GroupVersion and type
		metav1.AddToGroupVersion(scheme, iter8.GroupVersion)
		scheme.AddKnownTypes(iter8.GroupVersion, &iter8.Experiment{})
		scheme.AddKnownTypes(iter8.GroupVersion, &iter8.ExperimentList{})
		return nil
	}

	var schemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	scheme := runtime.NewScheme()
	err = schemeBuilder.AddToScheme(scheme)

	if err == nil {
		rc, err = client.New(restConf, client.Options{
			Scheme: scheme,
		})
		if err == nil {
			return rc, nil
		}
	}
	return nil, errors.New("cannot get client using rest config")
}

// GetExperiment gets the experiment from cluster
func GetExperiment(latest bool, name string, namespace string) (*Experiment, error) {
	results := iter8.ExperimentList{}
	var exp *iter8.Experiment
	var err error

	// get all experiments
	var rc client.Client
	if rc, err = GetClient(); err == nil {
		err = rc.List(context.Background(), &results, &client.ListOptions{})
	}

	// get latest experiment
	if latest && err == nil {
		if len(results.Items) > 0 {
			exp = &results.Items[len(results.Items)-1]
		} else {
			err = errors.New("no experiments found in cluster")
		}
	}

	// get named experiment
	if !latest && err == nil {
		for i := range results.Items {
			if results.Items[i].Name == name && results.Items[i].Namespace == namespace {
				exp = &results.Items[i]
				break
			}
		}
		if exp == nil {
			err = errors.New("Experiment " + name + " not found in namespace " + namespace)
		}
	}

	// return error
	if err != nil {
		return nil, err
	}

	// Return experiment
	return &Experiment{
		*exp,
	}, nil
}

// Started indicates if at least one iteration of the experiment has completed.
func (e *Experiment) Started() bool {
	if e == nil {
		return false
	}
	c := e.Status.CompletedLoops
	return c != nil && *c > 0
}

// Completed indicates if the experiment has completed.
func (e *Experiment) Completed() bool {
	if e == nil {
		return false
	}
	c := e.Status.GetCondition(iter8.ExperimentConditionExperimentCompleted)
	return c != nil && c.IsTrue()
}

// WinnerFound indicates if the experiment has found a winning version (winner).
func (e *Experiment) WinnerFound() bool {
	if e == nil {
		return false
	}
	if a := e.Status.Analysis; a != nil {
		if w := a.Winner; w != nil {
			return w.WinnerFound
		}
	}
	return false
}

// GetVersions returns the slice of version name strings. If the VersionInfo section is not present in the experiment's spec, then this slice is empty.
func (e *Experiment) GetVersions() []string {
	return e.Spec.VersionInfo
}

// GetMetricStr returns the metric value as a string for a given metric and a given version.
func (e *Experiment) GetMetricStr(metric string, versionIdx int) string {
	v := e.GetMetricDec(metric, versionIdx)
	if v == nil {
		return "unavailable"
	}
	return v.String()
}

// GetMetricStrs returns the given metric's value as a slice of strings, whose elements correspond to versions.
func (e *Experiment) GetMetricStrs(metric string) []string {
	versions := e.GetVersions()
	reqs := make([]string, len(versions))
	for i := range versions {
		reqs[i] = e.GetMetricStr(metric, i)
	}
	return reqs
}

// GetMetricNameAndUnits extracts the name, and if specified, units for the given metricInfo object and combines them into a string.
func GetMetricNameAndUnits(metricInfo iter8.Metric) string {
	r := metricInfo.Name
	if metricInfo.Units != nil {
		r += fmt.Sprintf(" (" + *metricInfo.Units + ")")
	}
	return r
}

// StringifyObjective returns a string representation of the given objective.
func StringifyObjective(objective iter8.Objective) string {
	r := ""
	if objective.LowerLimit != nil {
		z := new(inf.Dec).Round(objective.LowerLimit.AsDec(), 3, inf.RoundCeil)
		r += z.String() + " <= "
	}
	r += objective.Metric
	if objective.UpperLimit != nil {
		z := new(inf.Dec).Round(objective.UpperLimit.AsDec(), 3, inf.RoundCeil)
		r += " <= " + z.String()
	}
	return r
}

// GetSatisfyStr returns a true/false/unavailable valued string denotating if a version satisfies the objective.
func (e *Experiment) GetSatisfyStr(objectiveIndex int, versionIndex int) string {
	ana := e.Status.Analysis
	if ana == nil {
		return "unavailable"
	}
	objectivesByVersion := ana.Objectives
	if versionIndex > len(objectivesByVersion) {
		return "unavailable"
	}

	assessmentsForVersion := objectivesByVersion[versionIndex]
	if len(assessmentsForVersion) > objectiveIndex {
		return fmt.Sprintf("%v", assessmentsForVersion[objectiveIndex])
	}

	return "unavailable"
}

// GetSatisfyStrs returns a slice of true/false/unavailable valued strings for an objective denoting if it is satisfied by versions.
func (e *Experiment) GetSatisfyStrs(objectiveIndex int) []string {
	versions := e.GetVersions()
	sat := make([]string, len(versions))
	for versionIndex := range versions {
		sat[versionIndex] = e.GetSatisfyStr(objectiveIndex, versionIndex)
	}
	return sat
}

// StringifyReward returns a string representation of the given reward.
func StringifyReward(reward iter8.Reward) string {
	r := ""
	r += reward.Metric
	if reward.PreferredDirection == iter8.PreferredDirectionHigher {
		r += " (higher better)"
	} else {
		r += " (lower better)"
	}
	return r
}

// GetMetricDec returns the metric value as a string for a given metric and a given version.
func (e *Experiment) GetMetricDec(metric string, versionIndex int) *inf.Dec {
	if e.Status.Analysis == nil || versionIndex > len(e.Status.Analysis.Metrics) {
		return nil
	}
	if values, ok := e.Status.Analysis.Metrics[versionIndex][metric]; ok {
		val := values[len(values)-1]
		z := new(inf.Dec).Round(val.AsDec(), 3, inf.RoundCeil)
		return z
	}

	return nil
}

// GetAnnotatedMetricStrs returns a slice of values for a reward
func (e *Experiment) GetAnnotatedMetricStrs(reward iter8.Reward) []string {
	versions := e.GetVersions()
	row := make([]string, len(versions))
	var currentBestIndex *int
	var currentBestValue *inf.Dec
	for versionIndex := range versions {
		val := e.GetMetricDec(reward.Metric, versionIndex)

		if val == nil {
			row[versionIndex] = "unavailable"
			continue
		}

		row[versionIndex] = val.String()

		// set currentBest if not already set
		if currentBestIndex == nil {
			currentBestIndex, currentBestValue = &versionIndex, val
			continue
		}

		// update currentBest

		if reward.PreferredDirection == iter8.PreferredDirectionHigher {
			if -1 == currentBestValue.Cmp(val) {
				currentBestIndex, currentBestValue = &versionIndex, val
			}
			continue
		}

		// reward.PreferredDirection == iter8.PreferredDirectionLower
		if currentBestValue.Cmp(val) == 1 {
			currentBestIndex, currentBestValue = &versionIndex, val
		}
	}

	// mark current best with '*'
	if currentBestIndex != nil {
		row[*currentBestIndex] = row[*currentBestIndex] + " *"
	}
	return row
}

// Assert verifies a given set of conditions for the experiment.
func (e *Experiment) Assert(conditions []ConditionType) error {
	for _, cond := range conditions {
		switch cond {
		case Completed:
			if !e.Completed() {
				return errors.New("experiment has not completed")
			}
		case WinnerFound:
			if !e.WinnerFound() {
				return errors.New("no winner found in experiment")
			}
		default:
			return errors.New("unsupported condition found in assertion")
		}
	}
	return nil
}
