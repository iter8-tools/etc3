package core

import (
	"context"

	"github.com/antonmedv/expr"
	iter8 "github.com/iter8-tools/etc3/api/v2beta1"
)

func init() {
	log = GetLogger()
}

// Task defines common method signatures for every task.
type Task interface {
	Run(ctx context.Context) error
	GetIf() *string
}

// IsARun determines if the given task spec is in fact a run spec.
func IsARun(t *iter8.TaskSpec) bool {
	return t.Run != nil && len(*t.Run) > 0
}

// IsARun determines if the given task spec is in fact a task spec.
func IsATask(t *iter8.TaskSpec) bool {
	return t.Task != nil && len(*t.Task) > 0
}

// Action is a slice of Tasks.
type Action []Task

// TaskMeta is common to all Tasks
type TaskMeta struct {
	Task *string `json:"task,omitempty" yaml:"task,omitempty"`
	Run  *string `json:"run,omitempty" yaml:"run,omitempty"`
	If   *string `json:"if,omitempty" yaml:"if,omitempty"`
}

// GetIf returns any 'if' from TaskMeta
func (tm TaskMeta) GetIf() *string {
	return tm.If
}

// NamedValue name/value to be used in constructing a REST query to backend metrics server
type NamedValue struct {
	// Name of parameter
	Name string `json:"name" yaml:"name"`

	// Value of parameter
	Value string `json:"value" yaml:"value"`
}
type VersionInfo struct {
	Variables []NamedValue
}

// Run the given action.
func (a *Action) Run(ctx context.Context) error {
	for i := 0; i < len(*a); i++ {
		log.Info("------ task starting")
		shouldRun := true
		exp, err := GetExperimentFromContext(ctx)
		if err != nil {
			return err
		}
		// if task has a condition
		if cond := (*a)[i].GetIf(); cond != nil {
			// condition evaluates to false ... then shouldRun is false
			program, err := expr.Compile(*cond, expr.Env(exp), expr.AsBool())
			if err != nil {
				return err
			}

			output, err := expr.Run(program, exp)
			if err != nil {
				return err
			}

			shouldRun = output.(bool)
		}
		if shouldRun {
			err := (*a)[i].Run(ctx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// GetDefaultTags creates interpolation.Tags from experiment referenced by context
func GetDefaultTags(ctx context.Context) *Tags {
	tags := NewTags()
	exp, err := GetExperimentFromContext(ctx)
	if err == nil {
		obj, err := exp.ToMap()
		if err == nil {
			tags = tags.
				With("this", obj)
		}
	} else {
		log.Warn("No experiment found in context")
	}

	return &tags
}
