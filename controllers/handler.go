/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// handlers.go implements code to start jobs

package controllers

import (
	"context"
	"fmt"

	"github.com/iter8-tools/etc3/api/v2beta1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// HandlerType types of handlers
type HandlerType string

const (
	// ServiceAccountForHandlers is the service account name to use for jobs
	ServiceAccountForHandlers string = "iter8-handlers"
	// HandlerTypeStart is the type of a start handler
	HandlerTypeStart HandlerType = "Start"
	// HandlerTypeFinish is the type of a finish handler
	HandlerTypeFinish HandlerType = "Finish"
	// HandlerTypeLoop is the type of a loop handler
	HandlerTypeLoop HandlerType = "Loop"

	// HandlerYaml is the name of the job spec used for handlers
	HandlerYaml = "handler.yaml"

	// LabelExperimentName is key of label to be added to handler jobs for experiment name
	LabelExperimentName = "iter8/experimentName"
	// LabelExperimentNamespace is key of label to be added to handler jobs for experiment namespace
	LabelExperimentNamespace = "iter8/experimentNamespace"
)

var allHandlerTypes []HandlerType = []HandlerType{
	HandlerTypeStart,
	HandlerTypeFinish,
	HandlerTypeLoop,
}

// GetHandler returns handler of a given type
func (r *ExperimentReconciler) GetHandler(instance *v2beta1.Experiment, t HandlerType) *string {
	var hdlr *string
	switch t {
	case HandlerTypeStart:
		hdlr = instance.Spec.GetStartHandler()
	case HandlerTypeFinish:
		hdlr = instance.Spec.GetFinishHandler()
	default: // case HandlerTypeLoop:
		hdlr = instance.Spec.GetLoopHandler()
	}

	// Before returning, check if there are actually actions to execute.
	// If not, return nil (no handler). Otherwise, return the handler.
	// This approach is an optimization (we won't start jobs that do basically nothing).
	// It also helps writing test cases because we don't fail immediately after the start handler.
	if _, ok := instance.Spec.Actions[*hdlr]; !ok {
		// no actions for this handler, return nil
		return nil
	}
	return hdlr
}

// JobManager enables mocking of handler jobs during tests
type JobManager interface {
	Get(ctx context.Context, ref types.NamespacedName, job *batchv1.Job) error
}

// IsHandlerLaunched returns the handler (job) if one has been launched
// Otherwise it returns nil
func (r *ExperimentReconciler) IsHandlerLaunched(ctx context.Context, instance *v2beta1.Experiment, handler string, handlerInstance *int) (*batchv1.Job, error) {
	log := Logger(ctx)
	log.Info("IsHandlerLaunched called", "handler", handler)

	job := &batchv1.Job{}
	ref := types.NamespacedName{Namespace: r.Iter8Config.Namespace, Name: jobName(instance, handler, handlerInstance)}
	// err := r.Get(ctx, ref, job)
	err := r.JobManager.Get(ctx, ref, job)
	if err != nil {
		log.Info("IsHandlerLaunched returning", "handler", handler, "launched", false)
		return nil, err
	}
	log.Info("IsHandlerLaunched returning", "handler", handler, "launched", true, "job", *job)
	return job, nil
}

// LaunchHandler lauches the job that implements a particular handler
func (r *ExperimentReconciler) LaunchHandler(ctx context.Context, instance *v2beta1.Experiment, handler string, handlerInstance *int) error {
	log := Logger(ctx)
	log.Info("LaunchHandler called", "handler", handler)
	defer log.Info("LaunchHandler completed", "handler", handler)

	job := defineJob(jobHandlerConfig{
		JobName:             jobName(instance, handler, handlerInstance),
		JobNamespace:        r.Iter8Config.Namespace,
		Image:               r.Iter8Config.TaskRunner,
		Action:              handler,
		ExperimentName:      instance.Name,
		ExperimentNamespace: instance.Namespace,
		LogLevel:            "trace",
	})

	// jobs are in iter8-system namespace; not experiment namespace
	// so experiments can't be owners.
	// Perhaps no owner is necessary. Or perhaps the iter8-controller Deployment
	// // assign owner to job (so job is automatically deleted when experiment is deleted)
	// controllerutil.SetControllerReference(instance, &job, r.Scheme)
	log.Info("LaunchHandler job", "job", job)

	// launch job
	if err := r.Create(ctx, job); err != nil {
		// if job already exists ignore the error
		if !errors.IsAlreadyExists(err) {
			log.Error(err, "create job failed")
			return err
		}
	}

	return nil
}

// This is an alternate way to define a batchv2.Job via a hardcoded pattern
// For now at least, we use a domain package provided job spec on the assumption
// that the domain author needs to create one to test the jobs anyway.

type jobHandlerConfig struct {
	JobName             string
	JobNamespace        string
	Image               string
	Action              string
	ExperimentName      string
	ExperimentNamespace string
	LogLevel            string
}

const (
	defaultBackoffLimit        = int32(1)
	defaultActiveDeadline      = int64(300)
	defaultExperimentNamespace = "iter8-system"
)

func defineJob(jobCfg jobHandlerConfig) *batchv1.Job {
	backoffLimit := defaultBackoffLimit
	activeDeadline := defaultActiveDeadline

	if jobCfg.ExperimentNamespace == "" {
		jobCfg.ExperimentNamespace = defaultExperimentNamespace
	}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobCfg.JobName,
			Namespace: jobCfg.JobNamespace,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:          &backoffLimit,
			ActiveDeadlineSeconds: &activeDeadline,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						LabelExperimentName:      jobCfg.ExperimentName,
						LabelExperimentNamespace: jobCfg.ExperimentNamespace,
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "iter8-handlers",
					RestartPolicy:      "Never",
					Containers: []corev1.Container{{
						Name:    "iter8-handler",
						Image:   jobCfg.Image,
						Command: []string{"handler"},
						Args:    []string{"run", "-a", jobCfg.Action},
						Env: []corev1.EnvVar{{
							Name:  "EXPERIMENT_NAME",
							Value: jobCfg.ExperimentName,
						}, {
							Name:  "EXPERIMENT_NAMESPACE",
							Value: jobCfg.ExperimentNamespace,
						}, {
							Name:  "ACTION",
							Value: jobCfg.Action,
						}, {
							Name:  "LOG_LEVEL",
							Value: jobCfg.LogLevel,
						}},
					}},
				},
			},
		},
	}
}

// HandlerJobCompleted returns true if the job is completed (has the JobComplete condition set to true)
func HandlerJobCompleted(handlerJob *batchv1.Job) bool {
	c := GetJobCondition(handlerJob, batchv1.JobComplete)
	return c != nil && c.Status == corev1.ConditionTrue
}

// HandlerJobFailed returns  true if the job has failed (has the JobFailed condition set to true)
func HandlerJobFailed(handlerJob *batchv1.Job) bool {
	c := GetJobCondition(handlerJob, batchv1.JobFailed)
	return c != nil && c.Status == corev1.ConditionTrue
}

// generate job name
func jobName(instance *v2beta1.Experiment, handler string, handlerInstance *int) string {
	name := fmt.Sprintf("%s-%s", instance.Name, handler)
	if handlerInstance != nil {
		name = fmt.Sprintf("%s-%d", name, *handlerInstance)
	}

	return name
}

// GetJobCondition is a utility to retrieve a condition from a Job resource
// returns nil if it is not present
func GetJobCondition(job *batchv1.Job, condition batchv1.JobConditionType) *batchv1.JobCondition {
	for _, c := range job.Status.Conditions {
		if c.Type == condition {
			return &c
		}
	}
	return nil
}

// HandlerStatusType is the type of a handler status
type HandlerStatusType string

const (
	// HandlerStatusNoHandler indicates that there is no handler
	HandlerStatusNoHandler HandlerStatusType = "NoHandler"
	// HandlerStatusNotLaunched indicates that the handler has not been lauched
	HandlerStatusNotLaunched HandlerStatusType = "NotLaunched"
	// HandlerStatusRunning indicates that the handler is executing
	HandlerStatusRunning HandlerStatusType = "Running"
	// HandlerStatusFailed indicates that the handler failed during execution
	HandlerStatusFailed HandlerStatusType = "Failed"
	// HandlerStatusComplete indicates that the handler has successfully executed to completion
	HandlerStatusComplete HandlerStatusType = "Complete"
)

// GetHandlerStatus determines a handlers status
func (r *ExperimentReconciler) GetHandlerStatus(ctx context.Context, instance *v2beta1.Experiment, handler *string, handlerInstance *int) HandlerStatusType {
	log := Logger(ctx)
	log.Info("GetHandlerStatus called", "handler", handler)

	if nil == handler {
		log.Info("GetHandlerStatus returning", "handler", handler, "status", HandlerStatusNoHandler)
		return HandlerStatusNoHandler
	}

	// has a handler specified
	handlerJob, err := r.IsHandlerLaunched(ctx, instance, *handler, handlerInstance)
	if err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, "Error trying to find handler job.")
			log.Info("GetHandlerStatus returning", "handler", handler, "status", HandlerStatusFailed)
			return HandlerStatusFailed
		}
	}

	if handlerJob == nil {
		// handler job not lauched
		log.Info("GetHandlerStatus returning", "handler", handler, "status", HandlerStatusNotLaunched)
		return HandlerStatusNotLaunched
	}

	// handler job has already been launched

	if HandlerJobCompleted(handlerJob) {
		log.Info("GetHandlerStatus returning", "handler", handler, "status", HandlerStatusComplete)
		return HandlerStatusComplete
	}
	if HandlerJobFailed(handlerJob) {
		log.Info("GetHandlerStatus returning", "handler", handler, "status", HandlerStatusFailed)
		return HandlerStatusFailed
	}

	// handler job exists and is done
	log.Info("GetHandlerStatus returning", "handler", handler, "status", HandlerStatusRunning)
	return HandlerStatusRunning
}
