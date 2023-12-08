package service_plans

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/cloudfoundry/cf-test-helpers/v2/workflowhelpers"
	. "github.com/onsi/ginkgo/v2"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"

	"github.com/cloudfoundry/cf-performance-tests/helpers"
)

var _ = Describe("service plans", func() {
	Describe("GET /v3/service_plans", func() {
		It("lists all /v3/service_plans as admin", func() {
			experiment := gmeasure.NewExperiment("GET /v3/service_plans::as admin::list all")
			AddReportEntry(experiment.Name, experiment)

			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/service_plans", func() {
						helpers.TimeCFCurl(testConfig.BasicTimeout, "/v3/service_plans")
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})

		It("list all /v3/service_plans as admin with large page size", func() {
			experiment := gmeasure.NewExperiment(fmt.Sprintf("GET /v3/service_plans::as admin::list with page size %d", testConfig.LargePageSize))
			AddReportEntry(experiment.Name, experiment)

			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/service_plans", func() {
						helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf("/v3/service_plans?per_page=%d", testConfig.LargePageSize))
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})

		It("lists all /v3/service_plans as regular user", func() {
			experiment := gmeasure.NewExperiment("GET /v3/service_plans::as regular user::list all")
			AddReportEntry(experiment.Name, experiment)

			workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/service_plans", func() {
						helpers.TimeCFCurl(testConfig.BasicTimeout, "/v3/service_plans")
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})
	})

	Describe("GET /v3/service_plans/:guid", func() {
		var servicePlanGUID string
		Context("as admin", func() {
			BeforeEach(func() {
				servicePlanGUIDs := helpers.GetGUIDs(testSetup.AdminUserContext(), testConfig, "/v3/service_plans")
				Expect(servicePlanGUIDs).NotTo(BeNil())
				servicePlanGUID = servicePlanGUIDs[rand.Intn(len(servicePlanGUIDs))]
			})

			It("shows one /v3/service_plans/:guid as admin", func() {
				experiment := gmeasure.NewExperiment("GET /v3/service_plans/:guid::as admin::show one")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						experiment.MeasureDuration("GET /v3/service_plans/:guid", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf("/v3/service_plans/%s", servicePlanGUID))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})
		})

		Context("as regular user", func() {
			BeforeEach(func() {
				servicePlanGUID = getRandomLimitedServicePlanGuid()
			})

			It("shows one /v3/service_plans/:guid as user", func() {
				experiment := gmeasure.NewExperiment("GET /v3/service_plans/:guid::as regular user::show one")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						experiment.MeasureDuration("GET /v3/service_plans/:guid", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf("/v3/service_plans/%s", servicePlanGUID))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})
		})
	})

	Describe("GET /v3/service_plans/:guid/visibility", func() {
		var servicePlanGUID string
		BeforeEach(func() {
			servicePlanGUID = getRandomLimitedServicePlanGuid()
		})

		Context("as admin", func() {
			It("shows one /v3/service_plans/:guid/visibility as admin", func() {
				experiment := gmeasure.NewExperiment("GET /v3/service_plans/:guid/visibility::as admin::show visibility")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						experiment.MeasureDuration("GET /v3/service_plans/:guid/visibility", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf("/v3/service_plans/%s/visibility", servicePlanGUID))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})
		})

		Context("as regular user", func() {
			It("shows one /v3/service_plans/:guid/visibility as user", func() {
				experiment := gmeasure.NewExperiment("GET /v3/service_plans/:guid/visibility::as regular user::show visibility")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						experiment.MeasureDuration("GET /v3/service_plans/:guid/visibility", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf("/v3/service_plans/%s/visibility", servicePlanGUID))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})
		})
	})

	Describe("GET /v3/service_plans?service_offering_guids=", func() {
		var serviceOfferingGuidsList []string
		BeforeEach(func() {
			serviceOfferingGuidsList = nil
			serviceOfferingsStatement := fmt.Sprintf("SELECT guid FROM services ORDER BY %s LIMIT 50", helpers.GetRandomFunction(testConfig))
			serviceOfferingGuids := helpers.ExecuteSelectStatement(ccdb, ctx, serviceOfferingsStatement)
			for _, guid := range serviceOfferingGuids {
				serviceOfferingGuidsList = append(serviceOfferingGuidsList, helpers.ConvertToString(guid))
			}
		})

		Context("as admin", func() {
			It("filters for list of service_offerings", func() {
				experiment := gmeasure.NewExperiment("GET /v3/service_plans?service_offering_guids=::as admin::filter for list of service_offerings")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						experiment.MeasureDuration("GET /v3/service_plans?service_offering_guids=:guid", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf(
								"/v3/service_plans?service_offering_guids=%v", strings.Join(serviceOfferingGuidsList[:], ",")))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})

			It(fmt.Sprintf("filters for list of service_offerings with page size %d", testConfig.LargePageSize), func() {
				experiment := gmeasure.NewExperiment(fmt.Sprintf("GET /v3/service_plans?service_offering_guids=::as admin::filter for list of service_offerings with page size %d", testConfig.LargePageSize))
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						experiment.MeasureDuration("GET /v3/service_plans?service_offering_guids=:guid", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf(
								"/v3/service_plans?service_offering_guids=%v&per_page=%d",
								strings.Join(serviceOfferingGuidsList[:], ","), testConfig.LargePageSize))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})
		})

		Context("as regular user", func() {
			It("filters for list of service_offerings", func() {
				experiment := gmeasure.NewExperiment("GET /v3/service_plans?service_offering_guids=::as regular user::filter for list of service_offerings")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						experiment.MeasureDuration("GET /v3/service_plans?service_offering_guids=:guid", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf(
								"/v3/service_plans?service_offering_guids=%v", strings.Join(serviceOfferingGuidsList[:], ",")))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})
		})
	})

	Describe("GET /v3/service_plans?service_instance_guids=", func() {
		var serviceInstanceGuidsList []string
		BeforeEach(func() {
			serviceInstanceGuidsList = nil
			serviceInstanceStatement := fmt.Sprintf("SELECT guid FROM service_instances ORDER BY %s LIMIT 50", helpers.GetRandomFunction(testConfig))
			serviceInstanceGuids := helpers.ExecuteSelectStatement(ccdb, ctx, serviceInstanceStatement)
			for _, guid := range serviceInstanceGuids {
				serviceInstanceGuidsList = append(serviceInstanceGuidsList, helpers.ConvertToString(guid))
			}
		})

		Context("as admin", func() {
			It("filters for list of service_instances", func() {
				experiment := gmeasure.NewExperiment("GET /v3/service_plans?service_instance_guids=::as admin::filter for list of service_instances")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						experiment.MeasureDuration("GET /v3/service_plans?service_instances_guids=:guid", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf(
								"/v3/service_plans?service_instance_guids=%v", strings.Join(serviceInstanceGuidsList[:], ",")))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})

			It(fmt.Sprintf("filters for list of service_instances with page size %d", testConfig.LargePageSize), func() {
				experiment := gmeasure.NewExperiment(fmt.Sprintf("GET /v3/service_plans?service_instance_guids=::as admin::filter for list of service_instances with page size %d", testConfig.LargePageSize))
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						experiment.MeasureDuration("GET /v3/service_plans?service_instances_guids=:guid", func() {
							helpers.TimeCFCurl(testConfig.LongTimeout, fmt.Sprintf(
								"/v3/service_plans?service_instance_guids=%v&per_page=%d",
								strings.Join(serviceInstanceGuidsList[:], ","), testConfig.LargePageSize))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})
		})

		Context("as regular user", func() {
			It("filters for list of service_instances", func() {
				experiment := gmeasure.NewExperiment("GET /v3/service_plans?service_instance_guids=::as regular user::filter for list of service_instances")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						experiment.MeasureDuration("GET /v3/service_plans?service_instances_guids=:guid", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf(
								"/v3/service_plans?service_instance_guids=%v", strings.Join(serviceInstanceGuidsList[:], ",")))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})

			It(fmt.Sprintf("filters for list of service_instances with page size %d", testConfig.LargePageSize), func() {
				experiment := gmeasure.NewExperiment(fmt.Sprintf("GET /v3/service_plans?service_instance_guids=::as regular user::filter for list of service_instances with page size %d", testConfig.LargePageSize))
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						experiment.MeasureDuration("GET /v3/service_plans?service_instances_guids=:guid", func() {
							helpers.TimeCFCurl(testConfig.LongTimeout, fmt.Sprintf(
								"/v3/service_plans?service_instance_guids=%v&per_page=%d",
								strings.Join(serviceInstanceGuidsList[:], ","), testConfig.LargePageSize))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})
		})
	})

	Describe("GET /v3/service_plans?organization_guids=&space_guids=", func() {
		var orgGuidsList []string
		var spaceGuidsList []string
		BeforeEach(func() {
			orgGuidsList = nil
			selectOrgGuidsStatement := fmt.Sprintf("SELECT guid FROM organizations WHERE name LIKE '%s-org-%%' ORDER BY %s LIMIT 50", testConfig.GetNamePrefix(), helpers.GetRandomFunction(testConfig))
			orgGuids := helpers.ExecuteSelectStatement(ccdb, ctx, selectOrgGuidsStatement)
			for _, guid := range orgGuids {
				orgGuidsList = append(orgGuidsList, helpers.ConvertToString(guid))
			}
			spaceGuidsList = nil
			selectSpaceGuidsStatement := fmt.Sprintf("SELECT guid FROM spaces WHERE name LIKE '%s-space-%%' ORDER BY %s LIMIT 50", testConfig.GetNamePrefix(), helpers.GetRandomFunction(testConfig))
			spaceGuids := helpers.ExecuteSelectStatement(ccdb, ctx, selectSpaceGuidsStatement)
			for _, guid := range spaceGuids {
				spaceGuidsList = append(spaceGuidsList, helpers.ConvertToString(guid))
			}
		})

		Context("as regular user", func() {
			It("filters by org and space guids", func() {
				experiment := gmeasure.NewExperiment("GET /v3/service_plans?organization_guids=&space_guids=::as regular user::filter by org and space guids")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						experiment.MeasureDuration("GET /v3/service_plans?organization_guids=:guid&space_guids=:guid", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf(
								"/v3/service_plans?organization_guids=%v&space_guids=%v", strings.Join(orgGuidsList[:], ","), strings.Join(spaceGuidsList[:], ",")))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})

		})
	})
})

func getRandomLimitedServicePlanGuid() string {
	servicePlanGUIDsStatement := fmt.Sprintf("SELECT s_p.guid FROM service_plans s_p INNER JOIN service_plan_visibilities s_p_v ON s_p.id = s_p_v.service_plan_id ORDER BY %s LIMIT 1", helpers.GetRandomFunction(testConfig))
	servicePlanGUIDs := helpers.ExecuteSelectStatement(ccdb, ctx, servicePlanGUIDsStatement)
	return helpers.ConvertToString(servicePlanGUIDs[0])
}
