package core

import (
	"errors"

	iter8 "github.com/iter8-tools/etc3/api/v2beta1"
)

// GetActionSpec gets a named action spec from an experiment.
func (e *Experiment) GetActionSpec(name string) (iter8.Action, error) {
	if e == nil {
		return nil, errors.New("GetActionSpec(...) called on nil experiment")
	}
	if e.Spec.Actions == nil {
		return nil, errors.New("nil actions")
	}
	if actionSpec, ok := e.Spec.Actions[name]; ok {
		return actionSpec, nil
	}
	return nil, errors.New("action with name " + name + " not found")
}

// MK
// // SetAggregatedBuiltinHists sets the experiment status field corresponding to aggregated built in hists
// func (e *Experiment) SetAggregatedBuiltinHists(fortioData v1.JSON) {
// 	if e.Status.Analysis == nil {
// 		e.Status.Analysis = &iter8.Analysis{}
// 	}
// 	abh := &iter8.AggregatedBuiltinHists{}
// 	e.Status.Analysis.AggregatedBuiltinHists = abh
// 	abh.Data = fortioData
// }
