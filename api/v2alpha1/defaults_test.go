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

package v2alpha1_test

import (
	"github.com/iter8-tools/etc3/api/v2alpha1"
	"github.com/iter8-tools/etc3/configuration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Handler Initialization", func() {
	var (
		experiment *v2alpha1.Experiment
		cfg        configuration.Iter8Config
	)
	Context("When hanlders are set in an experiment", func() {
		BeforeEach(func() {
			experiment = v2alpha1.NewExperiment("test", "default").
				WithStrategy(v2alpha1.StrategyTypeCanary).
				WithHandlers(map[string]string{"start": "expStart", "finish": "expFinish"}).
				Build()
		})

		It("the value of the start handler should match the value in experiment when a configuration is present", func() {
			cfg = configuration.NewIter8Config().
				WithStrategy(string(v2alpha1.StrategyTypeCanary), map[string]string{"start": "cfgStart", "finish": "cfgFinish"}).
				Build()

			Expect(experiment.Spec.Strategy.Handlers).ShouldNot(BeNil())
			Expect(experiment.Spec.Strategy.Handlers.Start).ShouldNot(BeNil())
			Expect(*experiment.Spec.GetStartHandler(cfg)).Should(Equal("expStart"))
			Expect(experiment.Spec.Strategy.Handlers.Finish).ShouldNot(BeNil())
			Expect(*experiment.Spec.GetFinishHandler(cfg)).Should(Equal("expFinish"))
			Expect(experiment.Spec.GetRollbackHandler(cfg)).Should(BeNil())
			Expect(experiment.Spec.GetFailureHandler(cfg)).Should(BeNil())
		})
		It("the value of the start handler should match the value in experiment when ca onfiguration is not present", func() {
			cfg = configuration.NewIter8Config().
				Build()
			Expect(experiment.Spec.Strategy.Handlers).ShouldNot(BeNil())
			Expect(experiment.Spec.Strategy.Handlers.Start).ShouldNot(BeNil())
			Expect(*experiment.Spec.GetStartHandler(cfg)).Should(Equal("expStart"))
			Expect(experiment.Spec.Strategy.Handlers.Finish).ShouldNot(BeNil())
			Expect(*experiment.Spec.GetFinishHandler(cfg)).Should(Equal("expFinish"))
			Expect(experiment.Spec.GetRollbackHandler(cfg)).Should(BeNil())
			Expect(experiment.Spec.GetFailureHandler(cfg)).Should(BeNil())
		})
	})

	Context("When handlers are not defined in an experiment", func() {
		experiment := v2alpha1.NewExperiment("test", "default").
			WithStrategy(v2alpha1.StrategyTypeCanary).
			Build()

		It("the value is set by late initialization from the configuration", func() {
			experiment := v2alpha1.NewExperiment("test", "default").
				WithStrategy(v2alpha1.StrategyTypeCanary).
				Build()
			cfg := configuration.NewIter8Config().
				WithStrategy(string(v2alpha1.StrategyTypeCanary), map[string]string{"start": "cfgStart", "finish": "cfgFinish"}).
				Build()
			changed := experiment.Spec.InitializeHandlers(cfg)
			Expect(changed).Should(Equal(true))
			Expect(experiment.Spec.Strategy.Handlers).ShouldNot(BeNil())
			Expect(experiment.Spec.Strategy.Handlers.Start).ShouldNot(BeNil())
			Expect(*experiment.Spec.GetStartHandler(cfg)).Should(Equal("cfgxStart"))
		})

		It("the value of the start handler shoud match the value im the configuration when it is present", func() {
			cfg := configuration.NewIter8Config().
				WithStrategy(string(v2alpha1.StrategyTypeCanary), map[string]string{"start": "cfgStart", "finish": "cfgFinish"}).
				Build()
			Expect(experiment.Spec.Strategy.Handlers).Should(BeNil())
			Expect(*experiment.Spec.GetStartHandler(cfg)).Should(Equal("cfgStart"))
			Expect(*experiment.Spec.GetFinishHandler(cfg)).Should(Equal("cfgFinish"))
			Expect(experiment.Spec.GetRollbackHandler(cfg)).Should(BeNil())
			Expect(experiment.Spec.GetFailureHandler(cfg)).Should(BeNil())
		})

		It("the value of the start handler shoud be nil when when no configuration is present", func() {
			cfg := configuration.NewIter8Config().
				Build()
			Expect(experiment.Spec.Strategy.Handlers).Should(BeNil())
			Expect(experiment.Spec.GetStartHandler(cfg)).Should(BeNil())
			Expect(experiment.Spec.GetFinishHandler(cfg)).Should(BeNil())
			Expect(experiment.Spec.GetRollbackHandler(cfg)).Should(BeNil())
			Expect(experiment.Spec.GetFailureHandler(cfg)).Should(BeNil())
		})
	})
})
