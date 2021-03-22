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
	"io/ioutil"
	"path"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/etc3/util"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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
	// HandlerTypeRollback is the type of a rollback handler
	HandlerTypeRollback HandlerType = "Rollback"
	// HandlerTypeFailure is the type of a failure handler
	HandlerTypeFailure HandlerType = "Failure"
	// HandlerTypeLoop is the type of a loop handler
	HandlerTypeLoop HandlerType = "Loop"

	// LabelExperimentName is key of label to be added to handler jobs for experiment name
	LabelExperimentName = "iter8/experimentName"
	// LabelExperimentNamespace is key of label to be added to handler jobs for experiment namespace
	LabelExperimentNamespace = "iter8/experimentNamespace"
)

var allHandlerTypes []HandlerType = []HandlerType{
	HandlerTypeStart,
	HandlerTypeFinish,
	HandlerTypeRollback,
	HandlerTypeFailure,
	HandlerTypeLoop,
}

var terminalHandlerTypes []HandlerType = []HandlerType{
	HandlerTypeFinish,
	HandlerTypeRollback,
	HandlerTypeFailure,
}

// GetHandler returns handler of a given type
func (r *ExperimentReconciler) GetHandler(instance *v2alpha2.Experiment, t HandlerType) *string {
	switch t {
	case HandlerTypeStart:
		return instance.Spec.GetStartHandler(r.Iter8Config)
	case HandlerTypeFinish:
		return instance.Spec.GetFinishHandler(r.Iter8Config)
	case HandlerTypeRollback:
		return instance.Spec.GetRollbackHandler(r.Iter8Config)
	case HandlerTypeFailure:
		return instance.Spec.GetFailureHandler(r.Iter8Config)
	default: // case HandlerTypeLoop:
		return instance.Spec.GetLoopHandler(r.Iter8Config)
	}
}

// IsHandlerLaunched returns the handler (job) if one has been launched
// Otherwise it returns nil
func (r *ExperimentReconciler) IsHandlerLaunched(ctx context.Context, instance *v2alpha2.Experiment, handler string, handlerInstance *int) (*batchv1.Job, error) {
	log := util.Logger(ctx)
	log.Info("IsHandlerLaunched called", "handler", handler)

	job := &batchv1.Job{}
	ref := types.NamespacedName{Namespace: r.Iter8Config.Namespace, Name: jobName(instance, handler, handlerInstance)}
	err := r.Get(ctx, ref, job)
	if err != nil {
		log.Info("IsHandlerLaunched returning", "handler", handler, "launched", false)
		return nil, err
	}
	log.Info("IsHandlerLaunched returning", "handler", handler, "launched", true)
	return job, nil
}

// LaunchHandler lauches the job that implements a particular handler
func (r *ExperimentReconciler) LaunchHandler(ctx context.Context, instance *v2alpha2.Experiment, handler string, handlerInstance *int) error {
	log := util.Logger(ctx)
	log.Info("LaunchHandler called", "handler", handler)
	defer log.Info("LaunchHandler completed", "handler", handler)

	handlerJobYaml := path.Join(r.Iter8Config.HandlersDir, fmt.Sprintf("%s.yaml", handler))
	log.Info("launchHandler", "jobYaml", handlerJobYaml)
	job := batchv1.Job{}
	if err := readJobSpec(handlerJobYaml, &job); err != nil {
		return err
	}

	// update job spec:
	//   - assign a name unique for this experiment, handler type
	//   - assign namespace same as namespace of iter8
	//   - define labels LabelExperimentName and LabelExperimentNamespace used for event filtering
	//   - set serviceAccountName to iter8-handlers
	//   - set environment variables: EXPERIMENT_NAME, EXPERIMENT_NAMESPACE
	job.Name = jobName(instance, handler, handlerInstance)
	job.Namespace = r.Iter8Config.Namespace
	if job.Spec.Template.ObjectMeta.Labels == nil {
		job.Spec.Template.ObjectMeta.SetLabels(map[string]string{})
	}
	job.Spec.Template.ObjectMeta.Labels[LabelExperimentName] = instance.Name
	job.Spec.Template.ObjectMeta.Labels[LabelExperimentNamespace] = instance.Namespace
	job.Spec.Template.Spec.ServiceAccountName = ServiceAccountForHandlers
	job.Spec.Template.Spec.Containers[0].Env = setEnvVariable(job.Spec.Template.Spec.Containers[0].Env, "EXPERIMENT_NAME", instance.Name)
	job.Spec.Template.Spec.Containers[0].Env = setEnvVariable(job.Spec.Template.Spec.Containers[0].Env, "EXPERIMENT_NAMESPACE", instance.Namespace)
	job.Spec.Template.Spec.Containers[0].Env = setEnvVariable(job.Spec.Template.Spec.Containers[0].Env, "ACTION", handler)

	// job := defineJob(jobHandlerConfig{
	// 	JobName:               jobName(instance, handler),
	// 	JobNamespace:          instance.Namespace,
	// 	JobServiceAccountName: "default",
	// 	Image:                 "iter8/iter8-kfserving:latest",
	// 	Commands:              []string{"handlers/scripts/start.sh"},
	// 	ExperimentName:        instance.Name,
	// 	ExperimentNamespace:   instance.Namespace,
	// })

	// jobs are in iter8-system namespace; not experiment namespace
	// so experiments can't be owners.
	// Perhaps no owner is necessary. Or perhaps the iter8-controller Deployment
	// // assign owner to job (so job is automatically deleted when experiment is deleted)
	// controllerutil.SetControllerReference(instance, &job, r.Scheme)
	log.Info("LaunchHandler job", "job", job)

	// launch job
	if err := r.Create(ctx, &job); err != nil {
		// if job already exists ignore the error
		if !errors.IsAlreadyExists(err) {
			log.Error(err, "create job failed")
			return err
		}
	}

	return nil
}

// readJobSpec reads job from yaml file to batchv1.Job object
// Found that the whole object was not getting unmarshalled
// Converting to JSON first seems to work better
// Could do this directly (cf. https://stackoverflow.com/questions/40737122/convert-yaml-to-json-without-struct)
// or using https://github.com/ghodss/yaml
// We use the latter
func readJobSpec(templateFile string, job *batchv1.Job) error {
	yamlFile, err := ioutil.ReadFile(templateFile)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(yamlFile, job); err == nil {
		return err
	}

	return nil
}

func setEnvVariable(env []v1.EnvVar, name string, value string) []v1.EnvVar {
	for i, e := range env {
		if e.Name == name {
			env[i].Value = value
			e.Value = value
			return env
		}
	}
	return append(env, v1.EnvVar{
		Name:  name,
		Value: value,
	})
}

// This is an alternate way to define a batchv2.Job via a hardcoded pattern
// For now at least, we use a domain package provided job spec on the assumption
// that the domain author needs to create one to test the jobs anyway.

// type jobHandlerConfig struct {
// 	JobName               string
// 	JobNamespace          string
// 	JobServiceAccountName string
// 	Image                 string
// 	Commands              []string
// 	BackoffLimit          *int32
// 	ExperimentName        string
// 	ExperimentNamespace   string
// }

// const (
// 	defaultServiceAccountName  = "default"
// 	defaultBackoffLimit        = int32(4)
// 	defaultJobNamespace        = "iter8"
// 	defaultExperimentNamespace = "default"
// )

// func defineJob(jobCfg jobHandlerConfig) *batchv1.Job {
// 	if jobCfg.BackoffLimit == nil {
// 		limit := defaultBackoffLimit
// 		jobCfg.BackoffLimit = &limit
// 	}
// 	if jobCfg.JobServiceAccountName == "" {
// 		jobCfg.JobServiceAccountName = defaultServiceAccountName
// 	}
// 	if jobCfg.ExperimentNamespace == "" {
// 		jobCfg.ExperimentNamespace = defaultExperimentNamespace
// 	}

// 	return &batchv1.Job{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      jobCfg.JobName,
// 			Namespace: jobCfg.JobNamespace,
// 		},
// 		Spec: batchv1.JobSpec{
// 			BackoffLimit: jobCfg.BackoffLimit,
// 			Template: corev1.PodTemplateSpec{
// 				Spec: corev1.PodSpec{
// 					ServiceAccountName: jobCfg.JobServiceAccountName,
// 					RestartPolicy:      "Never",
// 					Containers: []corev1.Container{{
// 						Name:    "handler",
// 						Image:   jobCfg.Image,
// 						Command: jobCfg.Commands,
// 						Env: []corev1.EnvVar{{
// 							Name:  "EXPERIMENT_NAME",
// 							Value: jobCfg.ExperimentName,
// 						}, {
// 							Name:  "EXPERIMENT_NAMESPACE",
// 							Value: jobCfg.ExperimentNamespace,
// 						}},
// 					}},
// 				},
// 			},
// 		},
// 	}
// }

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
func jobName(instance *v2alpha2.Experiment, handler string, handlerInstance *int) string {
	uid := string(instance.UID)
	name := fmt.Sprintf("%s-handler-%s-%s", handler, instance.Name, uid[strings.LastIndex(uid, "-")+1:])
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
func (r *ExperimentReconciler) GetHandlerStatus(ctx context.Context, instance *v2alpha2.Experiment, handler *string, handlerInstance *int) HandlerStatusType {
	log := util.Logger(ctx)
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
