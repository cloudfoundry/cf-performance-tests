package isolation_segments

import (
	"fmt"
	"math/rand"

	"github.com/cloudfoundry/cf-test-helpers/v2/workflowhelpers"
	. "github.com/onsi/ginkgo/v2"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"

	"github.com/cloudfoundry/cf-performance-tests/helpers"
)

var _ = Describe("isolation segments", func() {
	Describe("GET /v3/isolation_segments", func() {

		It("as admin", func() {
			experiment := gmeasure.NewExperiment("GET /v3/isolation_segments::as admin")
			AddReportEntry(experiment.Name, experiment)

			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET isolation_segments", func() {
						helpers.TimeCFCurl(testConfig.BasicTimeout, "/v3/isolation_segments")
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})

		It("as regular user", func() {
			experiment := gmeasure.NewExperiment("GET /v3/isolation_segments::as regular user")
			AddReportEntry(experiment.Name, experiment)

			workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/isolation_segments", func() {
						helpers.TimeCFCurl(testConfig.BasicTimeout, "/v3/isolation_segments")
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})

		It(fmt.Sprintf("as admin with page size %d", testConfig.LargePageSize), func() {
			experiment := gmeasure.NewExperiment(fmt.Sprintf("GET /v3/isolation_segments::as admin with page size %d", testConfig.LargePageSize))
			AddReportEntry(experiment.Name, experiment)

			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/isolation_segments", func() {
						helpers.TimeCFCurl(testConfig.LongTimeout, fmt.Sprintf("/v3/isolation_segments?per_page=%d", testConfig.LargePageSize))
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})
	})

	Describe("GET /v3/isolation_segments/:guid/relationships/organizations", func() {
		It("as admin", func() {
			isolationSegmentGUIDs := helpers.GetGUIDs(testSetup.AdminUserContext(), testConfig, "/v3/isolation_segments")
			Expect(isolationSegmentGUIDs).NotTo(BeNil())
			isolationSegmentGUID := isolationSegmentGUIDs[rand.Intn(len(isolationSegmentGUIDs))]

			experiment := gmeasure.NewExperiment("GET /v3/isolation_segments/:guid/relationships/organizations::as admin")
			AddReportEntry(experiment.Name, experiment)

			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/isolation_segments/:guid/relationships/organizations", func() {
						helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf("/v3/isolation_segments/%s/relationships/organizations", isolationSegmentGUID))
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})

		It("as regular user", func() {
			isolationSegmentGUIDs := helpers.GetGUIDs(testSetup.AdminUserContext(), testConfig, "/v3/isolation_segments")
			Expect(isolationSegmentGUIDs).NotTo(BeNil())
			isolationSegmentGUID := isolationSegmentGUIDs[rand.Intn(len(isolationSegmentGUIDs))]

			experiment := gmeasure.NewExperiment("GET /v3/isolation_segments/:guid/relationships/organizations::as regular user")
			AddReportEntry(experiment.Name, experiment)

			workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/isolation_segments/:guid/relationships/organizations", func() {
						helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf("/v3/isolation_segments/%s/relationships/organizations", isolationSegmentGUID))
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})
	})

	Describe("individually", func() {
		Describe("as admin", func() {
			var isolationSegmentGUID string
			BeforeEach(func() {
				isolationSegmentGUIDs := helpers.GetGUIDs(testSetup.AdminUserContext(), testConfig, "/v3/isolation_segments")
				Expect(isolationSegmentGUIDs).NotTo(BeNil())
				isolationSegmentGUID = isolationSegmentGUIDs[rand.Intn(len(isolationSegmentGUIDs))]
			})

			It("gets /v3/isolation_segments/:guid as admin", func() {
				experiment := gmeasure.NewExperiment("individually::as admin::GET /v3/isolation_segments/:guid")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						experiment.MeasureDuration("GET /v3/isolation_segments/:guid", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf("/v3/isolation_segments/%s", isolationSegmentGUID))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})

			It("patches /v3/isolation_segments/:guid as admin", func() {
				experiment := gmeasure.NewExperiment("individually::as admin::PATCH /v3/isolation_segments/:guid")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						experiment.MeasureDuration("PATCH /v3/isolation_segments/:guid", func() {
							data := `{ "metadata": { "annotations": { "test": "PATCH /v3/isolation_segments/:guid" } } }`
							helpers.TimeCFCurl(testConfig.BasicTimeout, "-X", "PATCH", "-d", data, fmt.Sprintf("/v3/isolation_segments/%s", isolationSegmentGUID))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})
		})

		Describe("as regular user", func() {
			It("gets /v3/isolation_segments/:guid as regular user", func() {
				isolationSegmentGUIDs := helpers.GetGUIDs(testSetup.RegularUserContext(), testConfig, "/v3/isolation_segments")
				Expect(isolationSegmentGUIDs).NotTo(BeNil())
				isolationSegmentGUID := isolationSegmentGUIDs[rand.Intn(len(isolationSegmentGUIDs))]

				experiment := gmeasure.NewExperiment("individually::as regular user::GET /v3/isolation_segments/:guid")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						experiment.MeasureDuration("GET /v3/isolation_segments/:guid", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf("/v3/isolation_segments/%s", isolationSegmentGUID))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})
		})
	})
})
