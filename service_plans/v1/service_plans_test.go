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
		Context("as admin", func() {
			It("list all /v3/service_plans as admin efficiently", func() {
				experiment := gmeasure.NewExperiment("as admin")
				AddReportEntry(experiment.Name, experiment)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/service_plans", func() {
						workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, "/v3/service_plans")
						})
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})

			It("list all /v3/service_plans as admin efficiently with large page size", func() {
				experiment := gmeasure.NewExperiment(fmt.Sprintf("list with page size %d", testConfig.LargePageSize))
				AddReportEntry(experiment.Name, experiment)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/service_plans", func() {
						workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf("/v3/service_plans?per_page=%d", testConfig.LargePageSize))
						})
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})

		Context("as regular user", func() {
			It("lists all /v3/service_plans as a regular user efficiently", func() {
				experiment := gmeasure.NewExperiment("as user")
				AddReportEntry(experiment.Name, experiment)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/service_plans", func() {
						workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, "/v3/service_plans")
						})
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

			It("shows one /v3/service_plans/:guid as admin efficiently", func() {
				experiment := gmeasure.NewExperiment("as admin")
				AddReportEntry(experiment.Name, experiment)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/service_plans/:guid", func() {
						workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf("/v3/service_plans/%s", servicePlanGUID))
						})
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})

		Context("as regular user", func() {
			BeforeEach(func() {
				servicePlanGUID = getRandomLimitedServicePlanGuid()
			})

			It("shows one /v3/service_plans/:guid as user efficiently", func() {
				experiment := gmeasure.NewExperiment("as user")
				AddReportEntry(experiment.Name, experiment)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/service_plans/:guid", func() {
						workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf("/v3/service_plans/%s", servicePlanGUID))
						})
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})
	})

	Describe("GET /v3/service_plans/:guid/visibility", func() {
		var servicePlanGUID string
		BeforeEach(func() {
			servicePlanGUID = getRandomLimitedServicePlanGuid()
		})

		Context("as admin", func() {
			It("shows one /v3/service_plans/:guid/visibility as admin efficiently", func() {
				experiment := gmeasure.NewExperiment("as admin")
				AddReportEntry(experiment.Name, experiment)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/service_plans/:guid/visibility", func() {
						workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf("/v3/service_plans/%s/visibility", servicePlanGUID))
						})
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})

		Context("as regular user", func() {
			It("shows one /v3/service_plans/:guid/visibility as user efficiently", func() {
				experiment := gmeasure.NewExperiment("as user")
				AddReportEntry(experiment.Name, experiment)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/service_plans/:guid/visibility", func() {
						workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf("/v3/service_plans/%s/visibility", servicePlanGUID))
						})
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})
	})

	Describe("GET /v3/service_plans?service_offering_guids=", func() {
		var serviceOfferingGuidsList []string
		BeforeEach(func() {
			serviceOfferingGuidsList = nil
			serviceOfferingGuids := helpers.ExecuteSelectStatement(ccdb, ctx,
				"SELECT guid FROM services ORDER BY random() LIMIT 50")
			for _, guid := range serviceOfferingGuids {
				serviceOfferingGuidsList = append(serviceOfferingGuidsList, guid.(string))
			}
		})

		Context("as admin", func() {
			It("filters for list of service_offerings", func() {
				experiment := gmeasure.NewExperiment("as admin")
				AddReportEntry(experiment.Name, experiment)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/service_plans?service_offering_guids=:guid", func() {
						workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf(
								"/v3/service_plans?service_offering_guids=%v", strings.Join(serviceOfferingGuidsList[:], ",")))
						})
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})

			It(fmt.Sprintf("filters for list of service_offerings with page size %d", testConfig.LargePageSize), func() {
				experiment := gmeasure.NewExperiment("as admin")
				AddReportEntry(experiment.Name, experiment)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/service_plans?service_offering_guids=:guid", func() {
						workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf(
								"/v3/service_plans?service_offering_guids=%v&per_page=%d",
								strings.Join(serviceOfferingGuidsList[:], ","), testConfig.LargePageSize))
						})
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})

		Context("as regular user", func() {
			It("filters for list of service_offerings", func() {
				experiment := gmeasure.NewExperiment("as user")
				AddReportEntry(experiment.Name, experiment)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/service_plans?service_offering_guids=:guid", func() {
						workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf(
								"/v3/service_plans?service_offering_guids=%v", strings.Join(serviceOfferingGuidsList[:], ",")))
						})
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})
	})

	Describe("GET /v3/service_plans?service_instance_guids=", func() {
		var serviceInstanceGuidsList []string
		BeforeEach(func() {
			serviceInstanceGuidsList = nil
			serviceInstanceGuids := helpers.ExecuteSelectStatement(ccdb, ctx,
				"SELECT guid FROM service_instances ORDER BY random() LIMIT 50")
			for _, guid := range serviceInstanceGuids {
				serviceInstanceGuidsList = append(serviceInstanceGuidsList, guid.(string))
			}
		})
		Context("as admin", func() {
			It("filters for list of service_instances", func() {
				experiment := gmeasure.NewExperiment("as admin")
				AddReportEntry(experiment.Name, experiment)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/service_plans?service_instances_guids=:guid", func() {
						workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf(
								"/v3/service_plans?service_instance_guids=%v", strings.Join(serviceInstanceGuidsList[:], ",")))
						})
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})

			It(fmt.Sprintf("filters for list of service_instances with page size %d", testConfig.LargePageSize), func() {
				experiment := gmeasure.NewExperiment("as admin")
				AddReportEntry(experiment.Name, experiment)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/service_plans?service_instances_guids=:guid", func() {
						workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
							helpers.TimeCFCurl(testConfig.LongTimeout, fmt.Sprintf(
								"/v3/service_plans?service_instance_guids=%v&per_page=%d",
								strings.Join(serviceInstanceGuidsList[:], ","), testConfig.LargePageSize))
						})
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})
		Context("as regular user", func() {
			It("filters for list of service_instances", func() {
				experiment := gmeasure.NewExperiment("as user")
				AddReportEntry(experiment.Name, experiment)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/service_plans?service_instances_guids=:guid", func() {
						workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf(
								"/v3/service_plans?service_instance_guids=%v", strings.Join(serviceInstanceGuidsList[:], ",")))
						})
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})

			It(fmt.Sprintf("filters for list of service_instances with page size %d", testConfig.LargePageSize), func() {
				experiment := gmeasure.NewExperiment("as user")
				AddReportEntry(experiment.Name, experiment)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/service_plans?service_instances_guids=:guid", func() {
						workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
							helpers.TimeCFCurl(testConfig.LongTimeout, fmt.Sprintf(
								"/v3/service_plans?service_instance_guids=%v&per_page=%d",
								strings.Join(serviceInstanceGuidsList[:], ","), testConfig.LargePageSize))
						})
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})
	})

	Describe("GET /v3/service_plans?organization_guids=&space_guids=", func() {
		var orgGuidsList []string
		var spaceGuidsList []string
		BeforeEach(func() {
			orgGuidsList = nil
			selectOrgGuidsStatement := fmt.Sprintf("SELECT guid FROM organizations WHERE name LIKE '%s-org-%%' ORDER BY random() LIMIT 50", testConfig.GetNamePrefix())
			orgGuids := helpers.ExecuteSelectStatement(ccdb, ctx, selectOrgGuidsStatement)
			for _, guid := range orgGuids {
				orgGuidsList = append(orgGuidsList, guid.(string))
			}
			spaceGuidsList = nil
			selectSpaceGuidsStatement := fmt.Sprintf("SELECT guid FROM spaces WHERE name LIKE '%s-space-%%' ORDER BY random() LIMIT 50", testConfig.GetNamePrefix())
			spaceGuids := helpers.ExecuteSelectStatement(ccdb, ctx, selectSpaceGuidsStatement)
			for _, guid := range spaceGuids {
				spaceGuidsList = append(spaceGuidsList, guid.(string))
			}
		})

		Context("as regular user", func() {
			It("filters by org and space guids", func() {
				experiment := gmeasure.NewExperiment("as user")
				AddReportEntry(experiment.Name, experiment)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/service_plans?organization_guids=:guid&space_guids=:guid", func() {
						workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf(
								"/v3/service_plans?organization_guids=%v&space_guids=%v", strings.Join(orgGuidsList[:], ","), strings.Join(spaceGuidsList[:], ",")))
						})
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})

		})
	})
})

func getRandomLimitedServicePlanGuid() string {
	servicePlanGUIDs := helpers.ExecuteSelectStatement(ccdb, ctx,
		"SELECT guid FROM service_plans WHERE id IN (SELECT service_plan_id FROM service_plan_visibilities ORDER BY random() LIMIT 1)")
	return servicePlanGUIDs[0].(string)
}
