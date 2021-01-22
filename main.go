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

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	v2alpha1 "github.com/iter8-tools/etc3/api/v2alpha1"
	"github.com/iter8-tools/etc3/configuration"
	"github.com/iter8-tools/etc3/controllers"
	// +kubebuilder:scaffold:imports
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
	_ = clientgoscheme.AddToScheme(scheme)

	_ = v2alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

type iter8Http struct{}

func (m *iter8Http) Post(url, contentType string, body []byte) (resp []byte, statuscocode int, err error) {
	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, body, "", "  ")
	setupLog.Info("post request", "URL", url)
	setupLog.Info(string(prettyJSON.Bytes()))
	raw, err := http.Post(url, contentType, bytes.NewBuffer(body))
	setupLog.Info("post result", "raw", raw)
	setupLog.Info("post error", "err", err)
	if err != nil {
		return nil, 500, err
	}

	defer raw.Body.Close()
	setupLog.Info("reading Body")
	b, err := ioutil.ReadAll(raw.Body)
	return b, raw.StatusCode, err
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "4715d5e4.iter8.tools",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	restCfg, err := config.GetConfig()
	if err != nil {
		setupLog.Error(err, "unable to configure manager")
	}

	cfg := configuration.Iter8Config{}
	if err := configuration.ReadConfig(path.Join(os.Getenv("DEFAULTS_DIR"), "defaults.yaml"), &cfg); err != nil {
		setupLog.Error(err, "unable to configure manager")
		os.Exit(1)
	}
	if err := validateConfig(cfg); err != nil {
		setupLog.Error(err, "unable to configure manager")
		os.Exit(1)
	}
	setupLog.Info("read config", "cfg", cfg)

	if err = (&controllers.ExperimentReconciler{
		Client:        mgr.GetClient(),
		Log:           ctrl.Log.WithName("controllers").WithName("Experiment"),
		Scheme:        mgr.GetScheme(),
		RestConfig:    restCfg,
		EventRecorder: mgr.GetEventRecorderFor(Iter8Controller),
		Iter8Config:   cfg,
		HTTP:          &iter8Http{},
		ReleaseEvents: make(chan event.GenericEvent),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Experiment")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func validateConfig(cfg configuration.Iter8Config) error {
	// validate EnvironmentTypes
	for _, expType := range cfg.ExperimentTypes {
		ok := false
		for _, validValue := range v2alpha1.ValidStrategyTypes {
			if expType.Name == string(validValue) {
				ok = true
			}
		}
		if !ok {
			return fmt.Errorf("Invalid experiment type: %s, valid types are: %v", expType.Name, v2alpha1.ValidStrategyTypes)
		}
	}
	return nil
}
