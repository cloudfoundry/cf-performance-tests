package users

import (
	"fmt"

	"github.com/cloudfoundry/cf-performance-tests/helpers"
	"github.com/cloudfoundry/cf-test-helpers/v2/workflowhelpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"
)

var _ = Describe("users", func() {
	Describe("GET /v3/organizations/:guid/users", func() {

		Context("as admin", func() {
			It("get all users in org", func() {
				experiment := gmeasure.NewExperiment("GET /v3/organizations/:guid/users::as admin")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
					experiment.Sample(func(idx int) {
						experiment.MeasureDuration("GET /v3/organizations/:guid/users", func() {
							_, body := helpers.TimeCFCurlReturning(testConfig.LongTimeout, fmt.Sprintf("/v3/organizations/%s/users", org_guid))
							response := helpers.ParseResponseBody(helpers.RemoveDebugOutput(body))
							Expect(response.Pagination.TotalResults).To(Equal(users))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})
		})

	})

	Describe("GET /v3/spaces/:guid/users", func() {

		Context("as admin", func() {
			It("get all users in space", func() {
				experiment := gmeasure.NewExperiment("GET /v3/spaces/:guid/users::as admin")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
					experiment.Sample(func(idx int) {
						experiment.MeasureDuration("GET /v3/spaces/:guid/users", func() {
							_, body := helpers.TimeCFCurlReturning(testConfig.LongTimeout, fmt.Sprintf("/v3/spaces/%s/users", space_guid))
							response := helpers.ParseResponseBody(helpers.RemoveDebugOutput(body))
							Expect(response.Pagination.TotalResults).To(Equal(users))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})
		})

	})
})
