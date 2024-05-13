package destinations

import (
	"fmt"

	"github.com/cloudfoundry/cf-test-helpers/v2/workflowhelpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"
	"github.com/google/uuid"

	"github.com/cloudfoundry/cf-performance-tests/helpers"
)

var _ = Describe("destinations", func() {
	Describe("as admin", func() {
		var routeGUIDs []string

		BeforeEach(func() {
			routeGUIDs = nil
			domainGUID := helpers.ApiCall(testSetup.AdminUserContext(), testConfig, "/v3/domains").Resources[0].GUID

			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
				// create one route per sample
				for range make([]struct{}, testConfig.Samples) {
					host := fmt.Sprintf("%s-host-%s", testConfig.GetNamePrefix(), uuid.NewString())
					data := fmt.Sprintf(`{
                                           "host": "%s",
                                           "relationships": {
                                             "domain": {
                                               "data": {
                                                 "guid": "%s"
                                               }
                                             },
                                             "space": {
                                               "data": {
                                                 "guid": "%s"
                                               }
                                             }
                                           }
                                         }`, host, domainGUID, spaceGuid)

					exitCode, body := helpers.TimeCFCurlReturning(testConfig.BasicTimeout, "-X", "POST", "-d", data, "/v3/routes")

					Expect(exitCode).To(Equal(0))
					Expect(body).To(ContainSubstring("201 Created"))

					response := helpers.ParseCreateResponseBody(helpers.RemoveDebugOutput(body))
					routeGUIDs = append(routeGUIDs, response.GUID)
				}
			})
		})

		Context("using a single app", func() {
			It("creates a new destination", func() {
				experiment := gmeasure.NewExperiment("POST /v3/routes/:guid/destinations::as admin::using single app")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
					experiment.Sample(func(idx int) {
						experiment.MeasureDuration("POST /v3/routes/:guid/destinations", func() {
							data := fmt.Sprintf(`{"destinations": [ { "app": { "guid": "%s"} } ] }`, appGuid1)
							exitCode, body := helpers.TimeCFCurlReturning(testConfig.LongTimeout, "-X", "POST", "-d", data, fmt.Sprintf("/v3/routes/%s/destinations", routeGUIDs[idx]))

							Expect(exitCode).To(Equal(0))
							Expect(body).To(ContainSubstring("200 OK"))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})

			It("replaces destinations of a route", func() {
				experiment := gmeasure.NewExperiment("PATCH /v3/routes/:guid/destinations::as admin::using single app")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
					experiment.Sample(func(idx int) {
						experiment.MeasureDuration("PATCH /v3/routes/:guid/destinations", func() {
							data := fmt.Sprintf(`{"destinations": [ { "app": { "guid": "%s"} } ] }`, appGuid1)
							exitCode, body := helpers.TimeCFCurlReturning(testConfig.LongTimeout, "-X", "PATCH", "-d", data, fmt.Sprintf("/v3/routes/%s/destinations", routeGUIDs[idx]))

							Expect(exitCode).To(Equal(0))
							Expect(body).To(ContainSubstring("200 OK"))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})

			It("removes a destination from a route", func() {
				experiment := gmeasure.NewExperiment("DELETE /v3/routes/:guid/destinations/:guid::as admin::using single app")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
					experiment.Sample(func(idx int) {
						data := fmt.Sprintf(`{"destinations": [ { "app": { "guid": "%s"} } ] }`, appGuid1)
						exitCode, body := helpers.TimeCFCurlReturning(testConfig.LongTimeout, "-X", "POST", "-d", data, fmt.Sprintf("/v3/routes/%s/destinations", routeGUIDs[idx]))

						Expect(exitCode).To(Equal(0))
						Expect(body).To(ContainSubstring("200 OK"))

						response := helpers.ParseDestinationsCreateResponseBody(helpers.RemoveDebugOutput(body))
						destinationGuid := response.Destinations[0].GUID

						experiment.MeasureDuration("DELETE /v3/routes/:guid/destinations/:guid", func() {
							exitCode, body := helpers.TimeCFCurlReturning(testConfig.LongTimeout, "-X", "DELETE", fmt.Sprintf("/v3/routes/%s/destinations/%s", routeGUIDs[idx], destinationGuid))

							Expect(exitCode).To(Equal(0))
							Expect(body).To(ContainSubstring("204 No Content"))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})
		})

		Context("using two apps", func() {
			It("creates new destinations", func() {
				experiment := gmeasure.NewExperiment("POST /v3/routes/:guid/destinations::as admin::using two apps")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
					experiment.Sample(func(idx int) {
						experiment.MeasureDuration("POST /v3/routes/:guid/destinations", func() {
							data := fmt.Sprintf(`{"destinations": [ { "app": { "guid": "%s"} }, { "app": { "guid": "%s"} } ] }`, appGuid1, appGuid2)
							exitCode, body := helpers.TimeCFCurlReturning(testConfig.LongTimeout, "-X", "POST", "-d", data, fmt.Sprintf("/v3/routes/%s/destinations", routeGUIDs[idx]))

							Expect(exitCode).To(Equal(0))
							Expect(body).To(ContainSubstring("200 OK"))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})

			It("replaces destinations of a route", func() {
				experiment := gmeasure.NewExperiment("PATCH /v3/routes/:guid/destinations::as admin::using two apps")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
					experiment.Sample(func(idx int) {
						experiment.MeasureDuration("PATCH /v3/routes/:guid/destinations", func() {
							data := fmt.Sprintf(`{"destinations": [ { "app": { "guid": "%s"} }, { "app": { "guid": "%s"} } ] }`, appGuid1, appGuid2)
							exitCode, body := helpers.TimeCFCurlReturning(testConfig.LongTimeout, "-X", "PATCH", "-d", data, fmt.Sprintf("/v3/routes/%s/destinations", routeGUIDs[idx]))

							Expect(exitCode).To(Equal(0))
							Expect(body).To(ContainSubstring("200 OK"))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})
		})
	})
})
