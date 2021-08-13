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

package v2beta1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CurrentIterations", func() {
	Context("Iteration Utilities", func() {
		It("Work as Expected", func() {
			By("Creating an experiment")
			experiment := NewExperiment("test", "default").Build()

			By("Verifying that no iterations have been completed")
			Expect(experiment.Status.GetCompletedIterations()).Should(Equal(int32(0)))

			By("Incrementing the number of completed iterations")
			experiment.Status.IncrementCompletedIterations()
			experiment.Status.IncrementCompletedIterations()

			By("Checking that the number incremented")
			Expect(experiment.Status.GetCompletedIterations()).Should(Equal(int32(2)))
		})
	})
})

var _ = Describe("Winner Determination", func() {
	var experiment *Experiment
	BeforeEach(func() {
		experiment = NewExperiment("test", "default").
			WithVersion("baseline").WithVersion("candidate").WithVersion("winner").
			WithBaselineVersion("baseline", nil).
			WithCandidateVersion("candiate", nil).
			WithCandidateVersion("winner", nil).
			Build()
	})
	Context("When no Status.Analysis is present", func() {
		Specify("Version recommended for promotion is current baseline", func() {
			experiment.Status.SetVersionRecommendedForPromotion(experiment.Spec.VersionInfo.Baseline.Name)
			Expect(*experiment.Status.VersionRecommendedForPromotion).Should(Equal(experiment.Spec.VersionInfo.Baseline.Name))
		})
	})
	Context("When no winner assessment is present in Status.Analysis", func() {
		Specify("Version recommended for promotion is current baseline", func() {
			experiment.Status.Analysis = &Analysis{}
			experiment.Status.SetVersionRecommendedForPromotion(experiment.Spec.VersionInfo.Baseline.Name)
			Expect(*experiment.Status.VersionRecommendedForPromotion).Should(Equal(experiment.Spec.VersionInfo.Baseline.Name))
		})
	})
	Context("When no winner found", func() {
		Specify("Version recommended for promotion is current baseline", func() {
			winner := "winner"
			experiment.Status.Analysis = &Analysis{
				WinnerAssessment: &WinnerAssessmentAnalysis{
					AnalysisMetaData: AnalysisMetaData{},
					Data: WinnerAssessmentData{
						WinnerFound: false,
						Winner:      &winner,
					},
				},
			}
			experiment.Status.SetVersionRecommendedForPromotion(experiment.Spec.VersionInfo.Baseline.Name)
			Expect(*experiment.Status.VersionRecommendedForPromotion).Should(Equal(experiment.Spec.VersionInfo.Baseline.Name))
		})
	})
	Context("When winner found", func() {
		Specify("Version recommended for promotion is winner", func() {
			winner := "winner"
			experiment.Status.Analysis = &Analysis{
				WinnerAssessment: &WinnerAssessmentAnalysis{
					AnalysisMetaData: AnalysisMetaData{},
					Data: WinnerAssessmentData{
						WinnerFound: true,
						Winner:      &winner,
					},
				},
			}
			experiment.Status.SetVersionRecommendedForPromotion(experiment.Spec.VersionInfo.Baseline.Name)
			Expect(*experiment.Status.VersionRecommendedForPromotion).Should(Equal("winner"))
		})
	})

	var _ = Describe("TestingPattern", func() {
		var jqe string = "expr"
		Context("When experiment has 1 version, no reward", func() {
			It("TestingPattern should be SLOValidation", func() {
				experiment = NewExperiment("test", "default").
					WithVersion("baseline").
					Build()
				Expect(experiment.TestingPattern()).To(Equal(TestingPatternSLOValidation))
			})
		})
		Context("When experiment has 2 versions, no reward", func() {
			It("TestingPattern should be SLOValidation", func() {
				experiment = NewExperiment("test", "default").
					WithVersion("baseline").WithVersion("candidate").
					Build()
				Expect(experiment.TestingPattern()).To(Equal(TestingPatternSLOValidation))
			})
		})
		Context("When experiment has 3 versions, no reward", func() {
			It("TestingPattern should be SLOValidation", func() {
				experiment = NewExperiment("test", "default").
					WithVersion("v1").WithVersion("v2").WithVersion("v3").
					Build()
				Expect(experiment.TestingPattern()).To(Equal(TestingPatternSLOValidation))
			})
		})

		Context("When experiment has 2 version, 1 reward, no objectives", func() {
			It("TestingPattern should be TestingPatternAB", func() {
				experiment = NewExperiment("test", "default").
					WithVersion("v1").WithVersion("v2").
					WithReward(*NewMetric("reward", "default").WithJQExpression(&jqe).Build(), PreferredDirectionHigher).
					Build()
				Expect(experiment.TestingPattern()).To(Equal(TestingPatternAB))
			})
		})
		Context("When experiment has 3 version, reward, no objectives", func() {
			It("TestingPattern should be TestingPatternABN", func() {
				experiment = NewExperiment("test", "default").
					WithVersion("v1").WithVersion("v2").WithVersion("v3").
					WithReward(*NewMetric("reward", "default").WithJQExpression(&jqe).Build(), PreferredDirectionHigher).
					Build()
				Expect(experiment.TestingPattern()).To(Equal(TestingPatternABN))
			})
		})
		Context("When experiment has 2 version, reward, objective", func() {
			It("TestingPattern should be TestingPatternHybridAB", func() {
				experiment = NewExperiment("test", "default").
					WithVersion("v1").WithVersion("v2").
					WithReward(*NewMetric("reward", "default").WithJQExpression(&jqe).Build(), PreferredDirectionHigher).
					WithObjective(*NewMetric("objective", "default").WithJQExpression(&jqe).Build(), nil, nil, false).
					Build()
				Expect(experiment.TestingPattern()).To(Equal(TestingPatternHybridAB))
			})
		})
		Context("When experiment has 3 version, reward, objective", func() {
			It("TestingPattern should be TestingPatternHybridABN", func() {
				experiment = NewExperiment("test", "default").
					WithVersion("v1").WithVersion("v2").WithVersion("v3").
					WithReward(*NewMetric("reward", "default").WithJQExpression(&jqe).Build(), PreferredDirectionHigher).
					WithObjective(*NewMetric("objective", "default").WithJQExpression(&jqe).Build(), nil, nil, false).
					Build()
				Expect(experiment.TestingPattern()).To(Equal(TestingPatternHybridABN))
			})
		})
	})
})
