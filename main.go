/*
Copyright 2021.

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

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	batchv1 "k8s.io/api/batch/v1"

	v2alpha2 "github.com/iter8-tools/etc3/api/v2alpha2"
	v2alpha3 "github.com/iter8-tools/etc3/api/v2alpha3"
	"github.com/iter8-tools/etc3/controllers"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

const (
	// Iter8Controller string constant used to label event recorder
	Iter8Controller = "iter8"
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(v2alpha2.AddToScheme(scheme))
	utilruntime.Must(v2alpha3.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

type iter8Http struct{}

func (m *iter8Http) Post(url, contentType string, payload []byte) (resp []byte, statusCode int, err error) {
	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, payload, "", "  ")
	setupLog.Info("post request", "URL", url)
	setupLog.Info(string(prettyJSON.Bytes()))
	raw, err := http.Post(url, contentType, bytes.NewBuffer(payload))
	// setupLog.Info("post result", "raw", raw)
	setupLog.Info("post error", "err", err)
	if err != nil {
		return nil, raw.StatusCode, err
	}

	defer raw.Body.Close()
	setupLog.Info("reading Body")
	b, err := ioutil.ReadAll(raw.Body)
	return b, raw.StatusCode, err
}

type iter8JobManager struct {
	Client client.Client
}

func (j iter8JobManager) Get(ctx context.Context, ref types.NamespacedName, job *batchv1.Job) error {
	return j.Client.Get(ctx, ref, job)
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	restCfg, err := config.GetConfig()
	if err != nil {
		setupLog.Error(err, "unable to configure manager")
	}

	cfg := controllers.Iter8Config{}
	if err := controllers.ReadConfig(&cfg); err != nil {
		setupLog.Error(err, "unable to configure manager")
		os.Exit(1)
	}
	setupLog.Info("read config", "cfg", cfg)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                     scheme,
		MetricsBindAddress:         metricsAddr,
		Port:                       9443,
		HealthProbeBindAddress:     probeAddr,
		LeaderElection:             enableLeaderElection,
		LeaderElectionResourceLock: "leases",
		LeaderElectionNamespace:    cfg.Namespace,
		LeaderElectionID:           "leader.iter8.tools",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.ExperimentReconciler{
		Client:        mgr.GetClient(),
		Log:           ctrl.Log.WithName("controllers").WithName("Experiment"),
		Scheme:        mgr.GetScheme(),
		RestConfig:    restCfg,
		EventRecorder: mgr.GetEventRecorderFor(Iter8Controller),
		Iter8Config:   cfg,
		HTTP:          &iter8Http{},
		ReleaseEvents: make(chan event.GenericEvent),
		JobManager: iter8JobManager{
			Client: mgr.GetClient(),
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Experiment")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
