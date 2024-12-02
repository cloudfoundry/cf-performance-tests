package service_instances

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/cf-test-helpers/v2/workflowhelpers"
	. "github.com/onsi/ginkgo/v2"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"

	"github.com/cloudfoundry/cf-performance-tests/helpers"
)

var _ = Describe("service instances", func() {
	Describe("GET /v3/service_instances", func() {
		Context("as admin", func() {
			It("lists all /v3/service_instances", func() {
				experiment := gmeasure.NewExperiment("GET /v3/service_instances::as admin::list all")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						experiment.MeasureDuration("GET /v3/service_instances", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, "/v3/service_instances")
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})

			It("list all /v3/service_instances with a large page size", func() {
				experiment := gmeasure.NewExperiment(fmt.Sprintf("GET /v3/service_instances::as admin::list with a page size %d", testConfig.LargePageSize))
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						experiment.MeasureDuration("GET /v3/service_instances", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf("/v3/service_instances?per_page=%d", testConfig.LargePageSize))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})

			It("list all /v3/service_instances getting the last page", func() {
				experiment := gmeasure.NewExperiment("GET /v3/service_instances::as admin::list getting the last page")
				AddReportEntry(experiment.Name, experiment)

				var pages int
				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
					exitCode, body := helpers.TimeCFCurlReturning(testConfig.LongTimeout, "/v3/service_instances")
					Expect(exitCode).To(Equal(0))
					Expect(body).To(ContainSubstring("200 OK"))
					response := helpers.ParseResponseBody(helpers.RemoveDebugOutput(body))
					pages = response.Pagination.TotalPages
				})

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						experiment.MeasureDuration("GET /v3/service_instances", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf("/v3/service_instances?page=%d", pages))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})
		})

		Context("as regular user", func() {
			It("lists all /v3/service_instances", func() {
				experiment := gmeasure.NewExperiment("GET /v3/service_instances::as regular user::list all")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						experiment.MeasureDuration("GET /v3/service_instances", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, "/v3/service_instances")
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})
		})
	})

	Describe("GET /v3/service_instances?org_guids=", func() {
		Context("as admin", func() {
			It("filters by a single of organization_guids", func() {
				experiment := gmeasure.NewExperiment("GET /v3/service_instances?org_guids==::as admin::filter by a single organization_guid")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						orgGuidList := getRandomOrgGuids()

						experiment.MeasureDuration("GET /v3/service_instances?organization_guids=:guid", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf(
								"/v3/service_instances?organization_guids=%v", orgGuidList[0]))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})

			It("filters by a list of organization_guids with a large page size", func() {
				experiment := gmeasure.NewExperiment(fmt.Sprintf("GET /v3/service_instances?organization_guids==::as admin::filter for list of organization_guids with a page size %d", testConfig.LargePageSize))
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						orgGuidList := getRandomOrgGuids()

						experiment.MeasureDuration("GET /v3/service_instances?organization_guids=:guid", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf(
								"/v3/service_instances?per_page=%d&organization_guids=%v", testConfig.LargePageSize, strings.Join(orgGuidList[:], ",")))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})
		})
	})

	Describe("GET /v3/service_instances?space_guids=", func() {
		Context("as admin", func() {
			It("filters by a single of space_guid", func() {
				experiment := gmeasure.NewExperiment("GET /v3/service_instances?space_guids==::as admin::filter by a single space_guid")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						spaceGuidList := getRandomSpaceGuids()

						experiment.MeasureDuration("GET /v3/service_instances?space_guids=:guid", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf(
								"/v3/service_instances?per_page=%d&space_guids=%v", testConfig.LargePageSize, spaceGuidList[0]))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})

			It("filters by a list of space_guids with a large page size", func() {
				experiment := gmeasure.NewExperiment(fmt.Sprintf("GET /v3/service_instances?space_guids==::as admin::filter for list of space_guids with a page size %d", testConfig.LargePageSize))
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						spaceGuidList := getRandomSpaceGuids()

						experiment.MeasureDuration("GET /v3/service_instances?space_guids=:guid", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf(
								"/v3/service_instances?&space_guids=%v", strings.Join(spaceGuidList[:], ",")))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})
		})
	})

	Describe("GET /v3/service_instances?service_plan_guids=", func() {
		Context("as admin", func() {
			It("filters by a single service_plan_guid", func() {
				experiment := gmeasure.NewExperiment("GET /v3/service_instances?service_plan_guids==::as admin::filters by a single service_plan_guid")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						servicePlanGuidsList := getRandomServicePlanGuids()

						experiment.MeasureDuration("GET /v3/service_instances?service_plan_guids=:guid", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf(
								"/v3/service_instances?service_plan_guids=%v", servicePlanGuidsList[0]))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})

			It("filters for list of service_plan_guids with a large page size", func() {
				experiment := gmeasure.NewExperiment(fmt.Sprintf("GET /v3/service_instances?service_plan_guids==::as admin::filter for list of service_plan_guids with a page size %d", testConfig.LargePageSize))
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						servicePlanGuidsList := getRandomServicePlanGuids()

						experiment.MeasureDuration("GET /v3/service_instances?service_plan_guids=:guid", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf(
								"/v3/service_instances?per_page=%d&service_plan_guids=%v", testConfig.LargePageSize, strings.Join(servicePlanGuidsList[:], ",")))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})
		})
	})

	Describe("GET /v3/service_instances?service_plan_names=", func() {
		Context("as admin", func() {
			It("filters by a single service_plan_name", func() {
				experiment := gmeasure.NewExperiment("GET /v3/service_instances?service_plan_names==::as admin::filters by a single service_plan_name")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						servicePlanNamesList := getRandomServicePlanNames()

						experiment.MeasureDuration("GET /v3/service_instances?service_plan_names=:guid", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf(
								"/v3/service_instances?service_plan_names=%v", strings.Join(servicePlanNamesList[:], ",")))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})

			It("filters for list of service_plan_names with a large page size", func() {
				experiment := gmeasure.NewExperiment(fmt.Sprintf("GET /v3/service_instances?service_plan_names==::as admin::filter for list of service_plan_names with a page size %d", testConfig.LargePageSize))
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						servicePlanNamesList := getRandomServicePlanNames()

						experiment.MeasureDuration("GET /v3/service_instances?service_plan_names=:guid", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf(
								"/v3/service_instances?per_page=%d&service_plan_names=%v", testConfig.LargePageSize, strings.Join(servicePlanNamesList[:], ",")))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})
		})
	})
})

func getRandomOrgGuids() []string {
	return getRandomRows("guid", "organizations", fmt.Sprintf("%s-org", testConfig.GetNamePrefix()), 5)
}

func getRandomSpaceGuids() []string {
	return getRandomRows("guid", "spaces", fmt.Sprintf("%s-space", testConfig.GetNamePrefix()), 5)
}

func getRandomServicePlanGuids() []string {
	return getRandomRows("guid", "service_plans", fmt.Sprintf("%s-service-plan", testConfig.GetNamePrefix()), 5)
}

func getRandomServicePlanNames() []string {
	return getRandomRows("name", "service_plans", fmt.Sprintf("%s-service-plan", testConfig.GetNamePrefix()), 5)
}

func getRandomRows(column string, tableName string, namePrefix string, limit int) []string {
	var list []string = nil
	statement := fmt.Sprintf("SELECT %s FROM %s WHERE name LIKE '%s-%%' ORDER BY %s LIMIT %d", column, tableName, namePrefix, helpers.GetRandomFunction(testConfig), limit)

	cols := helpers.ExecuteSelectStatement(ccdb, ctx, statement)
	for _, guid := range cols {
		list = append(list, helpers.ConvertToString(guid))
	}

	Expect(len(list)).To(Equal(limit))

	return list
}
