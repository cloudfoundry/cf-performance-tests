package domains

import (
	"fmt"

	"github.com/cloudfoundry/cf-performance-tests/helpers"
	"github.com/cloudfoundry/cf-test-helpers/v2/workflowhelpers"
	. "github.com/onsi/ginkgo/v2"

	"github.com/onsi/gomega/gmeasure"
)

var _ = Describe("domains", func() {
	Describe("GET /v3/domains", func() {

		It("as admin", func() {
			experiment := gmeasure.NewExperiment("GET /v3/domains::as admin")
			AddReportEntry(experiment.Name, experiment)

			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/domains", func() {
						helpers.TimeCFCurl(testConfig.BasicTimeout, "/v3/domains")
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})

		It("as regular user", func() {
			experiment := gmeasure.NewExperiment("GET /v3/domains::as regular user")
			AddReportEntry(experiment.Name, experiment)

			workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/domains", func() {
						helpers.TimeCFCurl(testConfig.BasicTimeout, "/v3/domains")
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})

		It(fmt.Sprintf("as admin with page size %d", testConfig.LargePageSize), func() {
			experiment := gmeasure.NewExperiment(fmt.Sprintf("GET /v3/domains::as admin with page size %d", testConfig.LargePageSize))
			AddReportEntry(experiment.Name, experiment)

			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/domains", func() {
						helpers.TimeCFCurl(testConfig.LongTimeout, fmt.Sprintf("/v3/domains?per_page=%d", testConfig.LargePageSize))
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})
	})

	Describe("GET /v3/organizations/:guid/domains", func() {
		It("as admin", func() {
			experiment := gmeasure.NewExperiment("GET /v3/organizations/:guid/domains::as admin")
			AddReportEntry(experiment.Name, experiment)

			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
				experiment.Sample(func(idx int) {
					orgGUID := getRandomOrgWithPrivateDomain()

					experiment.MeasureDuration("GET /v3/organizations/:guid/domains", func() {
						helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf("/v3/organizations/%s/domains", orgGUID))
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})

		It("as regular user", func() {
			experiment := gmeasure.NewExperiment("GET /v3/organizations/:guid/domains::as regular user")
			AddReportEntry(experiment.Name, experiment)

			workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
				experiment.Sample(func(idx int) {
					orgGUID := getRandomOrgWithPrivateDomain()

					experiment.MeasureDuration("GET /v3/organizations/:guid/domains", func() {
						helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf("/v3/organizations/%s/domains", orgGUID))
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})
	})

	Describe("individually", func() {
		Describe("as admin", func() {
			It("gets /v3/domains/:guid as admin", func() {
				experiment := gmeasure.NewExperiment("individually::as admin::GET /v3/domains/:guid")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						domainGUID := getRandomPrivateDomain()

						experiment.MeasureDuration("GET /v3/domains/:guid", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf("/v3/domains/%s", domainGUID))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})

			It("patches /v3/domains/:guid as admin", func() {
				experiment := gmeasure.NewExperiment("individually::as admin::PATCH /v3/domains/:guid")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						domainGUID := getRandomPrivateDomain()

						experiment.MeasureDuration("PATCH /v3/domains/:guid", func() {
							data := `{ "metadata": { "annotations": { "test": "PATCH /v3/domains/:guid" } } }`
							helpers.TimeCFCurl(testConfig.BasicTimeout, "-X", "PATCH", "-d", data, fmt.Sprintf("/v3/domains/%s", domainGUID))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})

			It("deletes /v3/domains/:guid as admin", func() {
				experiment := gmeasure.NewExperiment("individually::as admin::DELETE /v3/domains/:guid")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						domainGUID := getRandomPrivateDomain()

						experiment.MeasureDuration("DELETE /v3/domains/:guid", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, "-X", "DELETE", fmt.Sprintf("/v3/domains/%s", domainGUID))
						})

						helpers.WaitToFail(testSetup.AdminUserContext(), testConfig, fmt.Sprintf("/v3/domains/%s", domainGUID))
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})
		})

		Describe("as regular user", func() {
			It("gets /v3/domains/:guid as regular user", func() {
				experiment := gmeasure.NewExperiment("individually::as regular user::GET /v3/domains/:guid")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						domainGUID := getRandomPrivateDomain()

						experiment.MeasureDuration("GET /v3/domains/:guid", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf("/v3/domains/%s", domainGUID))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})
		})
	})
})

func getRandomOrgWithPrivateDomain() string {
	var orgGuid string
	orgStatement := fmt.Sprintf("SELECT organizations.guid FROM organizations JOIN domains ON organizations.id = domains.owning_organization_id WHERE domains.owning_organization_id IS NOT null ORDER BY %s LIMIT 1", helpers.GetRandomFunction(testConfig))
	orgGuids := helpers.ExecuteSelectStatement(ccdb, ctx, orgStatement)
	for _, guid := range orgGuids {
		orgGuid = helpers.ConvertToString(guid)
	}

	return helpers.ConvertToString(orgGuid)
}

func getRandomPrivateDomain() string {
	var domainGuid string
	domainStatement := fmt.Sprintf("SELECT guid FROM domains WHERE domains.owning_organization_id IS NOT null ORDER BY %s LIMIT 1", helpers.GetRandomFunction(testConfig))
	domainGuids := helpers.ExecuteSelectStatement(ccdb, ctx, domainStatement)
	for _, guid := range domainGuids {
		domainGuid = helpers.ConvertToString(guid)
	}

	return helpers.ConvertToString(domainGuid)
}
