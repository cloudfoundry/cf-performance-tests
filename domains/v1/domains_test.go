package domains

import (
	"fmt"
	"math/rand"

	"github.com/cloudfoundry/cf-performance-tests/helpers"
	"github.com/cloudfoundry/cf-test-helpers/v2/workflowhelpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"
)

var _ = Describe("domains", func() {
	Describe("GET /v3/domains", func() {

		It("gets /v3/domains as admin efficiently", func() {
			experiment := gmeasure.NewExperiment("as admin")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("GET /v3/domains", func() {
					workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
						helpers.TimeCFCurl(testConfig.BasicTimeout, "/v3/domains")
					})
				})
			}, gmeasure.SamplingConfig{N: testConfig.Samples})
		})

		It("gets /v3/domains as a regular user efficiently", func() {
			experiment := gmeasure.NewExperiment("as user")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("GET /v3/domains", func() {
					workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
						helpers.TimeCFCurl(testConfig.BasicTimeout, "/v3/domains")
					})
				})
			}, gmeasure.SamplingConfig{N: testConfig.Samples})
		})

		It(fmt.Sprintf("gets /v3/domains as admin with page size %d efficiently", testConfig.LargePageSize), func() {
			experiment := gmeasure.NewExperiment(fmt.Sprintf("as admin with page size %d", testConfig.LargePageSize))
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration(fmt.Sprintf("GET /v3/domains"), func() {
					workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
						helpers.TimeCFCurl(testConfig.LongTimeout, fmt.Sprintf("/v3/domains?per_page=%d", testConfig.LargePageSize))
					})
				})
			}, gmeasure.SamplingConfig{N: testConfig.Samples})
		})
	})

	Describe("GET /v3/organizations/:guid/domains", func() {
		It("gets /v3/organizations/:guid/domains as admin efficiently", func() {
			orgGUIDs := helpers.GetGUIDs(testSetup.AdminUserContext(), testConfig, "/v3/organizations")
			Expect(orgGUIDs).NotTo(BeNil())
			orgGUID := orgGUIDs[rand.Intn(len(orgGUIDs))]

			experiment := gmeasure.NewExperiment("as admin")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("GET /v3/organizations/:guid/domains", func() {
					workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
						helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf("/v3/organizations/%s/domains", orgGUID))
					})
				})
			}, gmeasure.SamplingConfig{N: testConfig.Samples})
		})

		It("gets /v3/organizations/:guid/domains as regular user efficiently", func() {
			orgGUIDs := helpers.GetGUIDs(testSetup.RegularUserContext(), testConfig, "/v3/organizations")
			Expect(orgGUIDs).NotTo(BeNil())
			orgGUID := orgGUIDs[rand.Intn(len(orgGUIDs))]

			experiment := gmeasure.NewExperiment("as regular user")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("GET /v3/organizations/:guid/domains", func() {
					workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
						helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf("/v3/organizations/%s/domains", orgGUID))
					})
				})
			}, gmeasure.SamplingConfig{N: testConfig.Samples})
		})
	})

	Describe("individually", func() {
		Describe("as admin", func() {
			var domainGUID string
			BeforeEach(func() {
				domainGUIDs := helpers.GetGUIDs(testSetup.AdminUserContext(), testConfig, "/v3/domains")
				Expect(domainGUIDs).NotTo(BeNil())
				domainGUID = domainGUIDs[rand.Intn(len(domainGUIDs))]
			})

			It("gets /v3/domains/:guid as admin efficiently", func() {
				experiment := gmeasure.NewExperiment("as admin")
				AddReportEntry(experiment.Name, experiment)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/domains/:guid", func() {
						workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf("/v3/domains/%s", domainGUID))
						})
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})

			It("patches /v3/domains/:guid as admin efficiently", func() {
				experiment := gmeasure.NewExperiment("as admin")
				AddReportEntry(experiment.Name, experiment)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("PATCH /v3/domains/:guid", func() {
						workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
							data := `{ "metadata": { "annotations": { "test": "PATCH /v3/domains/:guid" } } }`
							helpers.TimeCFCurl(testConfig.BasicTimeout, "-X", "PATCH", "-d", data, fmt.Sprintf("/v3/domains/%s", domainGUID))
						})
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})

			It("deletes /v3/domains/:guid as admin efficiently", func() {
				experiment := gmeasure.NewExperiment("as admin")
				AddReportEntry(experiment.Name, experiment)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("DELETE /v3/domains/:guid", func() {
						workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
							domainGUIDs := helpers.GetGUIDs(testSetup.AdminUserContext(), testConfig, "/v3/domains")
							Expect(domainGUIDs).NotTo(BeNil())
							domainGUID = domainGUIDs[rand.Intn(len(domainGUIDs))]

							helpers.TimeCFCurl(testConfig.BasicTimeout, "-X", "DELETE", fmt.Sprintf("/v3/domains/%s", domainGUID))

							helpers.WaitToFail(testSetup.AdminUserContext(), testConfig, fmt.Sprintf("/v3/domains/%s", domainGUID))
						})
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})

		Describe("as regular user", func() {
			It("gets /v3/domains/:guid as regular user efficiently", func() {
				domainGUIDs := helpers.GetGUIDs(testSetup.RegularUserContext(), testConfig, "/v3/domains")
				Expect(domainGUIDs).NotTo(BeNil())
				domainGUID := domainGUIDs[rand.Intn(len(domainGUIDs))]

				experiment := gmeasure.NewExperiment("as regular user")
				AddReportEntry(experiment.Name, experiment)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/domains/:guid", func() {
						workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf("/v3/domains/%s", domainGUID))
						})
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})
	})
})
