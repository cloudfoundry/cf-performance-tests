package organization_quotas

import (
	"fmt"

	"github.com/cloudfoundry/cf-performance-tests/helpers"
	"github.com/cloudfoundry/cf-test-helpers/v2/workflowhelpers"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega/gmeasure"
)

var _ = Describe("organization_quotas", func() {
	Describe("GET /v3/organization_quotas", func() {

		It("as admin", func() {
			experiment := gmeasure.NewExperiment("GET /v3/organization_quotas::as admin")
			AddReportEntry(experiment.Name, experiment)

			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/organization_quotas", func() {
						helpers.TimeCFCurl(testConfig.LongTimeout, "/v3/organization_quotas")
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})

		It("as regular user", func() {
			experiment := gmeasure.NewExperiment("GET /v3/organization_quotas::as regular user")
			AddReportEntry(experiment.Name, experiment)

			workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.LongTimeout, func() {
				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/organization_quotas", func() {
						helpers.TimeCFCurl(testConfig.LongTimeout, "/v3/organization_quotas")
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})

		It(fmt.Sprintf("as regular user with page size %d", testConfig.LargePageSize), func() {
			experiment := gmeasure.NewExperiment(fmt.Sprintf("GET /v3/organization_quotas::as regular user with page size %d", testConfig.LargePageSize))
			AddReportEntry(experiment.Name, experiment)

			workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.LongTimeout, func() {
				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/organization_quotas", func() {
						helpers.TimeCFCurl(testConfig.LongTimeout, fmt.Sprintf("/v3/organization_quotas?per_page=%d", testConfig.LargePageSize))
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})
	})
})
