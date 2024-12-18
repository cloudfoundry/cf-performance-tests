package service_plans

import (
	"fmt"
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

		It("lists all /v3/service_plans as admin with large page size", func() {
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

			workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.LongTimeout, func() {
				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/service_plans", func() {
						helpers.TimeCFCurl(testConfig.LongTimeout, "/v3/service_plans")
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})
	})

	Describe("GET /v3/service_plans/:guid", func() {
		Context("as admin", func() {
			It("shows one /v3/service_plans/:guid as admin", func() {
				experiment := gmeasure.NewExperiment("GET /v3/service_plans/:guid::as admin::show one")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						servicePlanGUID := getRandomLimitedServicePlanGuid()

						experiment.MeasureDuration("GET /v3/service_plans/:guid", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf("/v3/service_plans/%s", servicePlanGUID))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})
		})

		Context("as regular user", func() {
			It("shows one /v3/service_plans/:guid as regular user", func() {
				experiment := gmeasure.NewExperiment("GET /v3/service_plans/:guid::as regular user::show one")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						servicePlanGUID := getRandomLimitedServicePlanGuid()

						experiment.MeasureDuration("GET /v3/service_plans/:guid", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf("/v3/service_plans/%s", servicePlanGUID))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})
		})
	})

	Describe("GET /v3/service_plans/:guid/visibility", func() {
		Context("as admin", func() {
			It("shows one /v3/service_plans/:guid/visibility as admin", func() {
				experiment := gmeasure.NewExperiment("GET /v3/service_plans/:guid/visibility::as admin::show visibility")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						var servicePlanGUID = getRandomLimitedServicePlanGuid()

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
						var servicePlanGUID = getRandomLimitedServicePlanGuid()

						experiment.MeasureDuration("GET /v3/service_plans/:guid/visibility", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf("/v3/service_plans/%s/visibility", servicePlanGUID))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})
		})
	})

	Describe("GET /v3/service_plans?service_offering_guids=", func() {
		Context("as admin", func() {
			It("filters for list of service_offerings", func() {
				experiment := gmeasure.NewExperiment("GET /v3/service_plans?service_offering_guids=::as admin::filter for list of service_offerings")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						serviceOfferingGuidsList := getRandomServiceOfferingGUIDs()

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
						serviceOfferingGuidsList := getRandomServiceOfferingGUIDs()

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
						serviceOfferingGuidsList := getRandomServiceOfferingGUIDs()

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
		Context("as admin", func() {
			It("filters for list of service_instances", func() {
				experiment := gmeasure.NewExperiment("GET /v3/service_plans?service_instance_guids=::as admin::filter for list of service_instances")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						serviceInstanceGuidsList := getRandomServiceInstanceGUIDs()

						experiment.MeasureDuration("GET /v3/service_plans?service_instances_guids=:guid", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf(
								"/v3/service_plans?service_instance_guids=%v", strings.Join(serviceInstanceGuidsList[:], ",")))
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
						serviceInstanceGuidsList := getRandomServiceInstanceGUIDs()

						experiment.MeasureDuration("GET /v3/service_plans?service_instances_guids=:guid", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf(
								"/v3/service_plans?service_instance_guids=%v", strings.Join(serviceInstanceGuidsList[:], ",")))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})
		})
	})

	Describe("GET /v3/service_plans?organization_guids=&space_guids=", func() {
		Context("as regular user", func() {
			It("filters by org and space guids", func() {
				experiment := gmeasure.NewExperiment("GET /v3/service_plans?organization_guids=&space_guids=::as regular user::filter by org and space guids")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						var orgGuidsList []string = nil
						selectOrgGuidsStatement := fmt.Sprintf("SELECT organizations.guid FROM organizations JOIN selected_orgs ON organizations.id = selected_orgs.id ORDER BY %s LIMIT 50", helpers.GetRandomFunction(testConfig))
						orgGuids := helpers.ExecuteSelectStatement(ccdb, ctx, selectOrgGuidsStatement)
						for _, guid := range orgGuids {
							orgGuidsList = append(orgGuidsList, helpers.ConvertToString(guid))
						}
						Expect(len(orgGuidsList)).To(Equal(50))

						var spaceGuidsList []string = nil
						selectSpaceGuidsStatement := fmt.Sprintf("SELECT spaces.guid FROM spaces JOIN selected_orgs ON spaces.organization_id = selected_orgs.id ORDER BY %s LIMIT 50", helpers.GetRandomFunction(testConfig))
						spaceGuids := helpers.ExecuteSelectStatement(ccdb, ctx, selectSpaceGuidsStatement)
						for _, guid := range spaceGuids {
							spaceGuidsList = append(spaceGuidsList, helpers.ConvertToString(guid))
						}
						Expect(len(spaceGuidsList)).To(Equal(50))

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
	// currently all service plan visibilities are for orgs the user has access to
	servicePlanGUIDsStatement := fmt.Sprintf("SELECT s_p.guid FROM service_plans s_p INNER JOIN service_plan_visibilities s_p_v ON s_p.id = s_p_v.service_plan_id WHERE s_p.name LIKE '%s-service-plan-%%' ORDER BY %s LIMIT 1", testConfig.GetNamePrefix(), helpers.GetRandomFunction(testConfig))
	servicePlanGUIDs := helpers.ExecuteSelectStatement(ccdb, ctx, servicePlanGUIDsStatement)
	return helpers.ConvertToString(servicePlanGUIDs[0])
}

func getRandomServiceInstanceGUIDs() []string {
	var serviceInstanceGuidsList []string = nil

	// all service instances are being created in a space the user has access to
	serviceInstanceStatement := fmt.Sprintf("SELECT guid FROM service_instances WHERE name LIKE '%s-service-instance-%%' ORDER BY %s LIMIT 200", testConfig.GetNamePrefix(), helpers.GetRandomFunction(testConfig))
	serviceInstanceGuids := helpers.ExecuteSelectStatement(ccdb, ctx, serviceInstanceStatement)
	for _, guid := range serviceInstanceGuids {
		serviceInstanceGuidsList = append(serviceInstanceGuidsList, helpers.ConvertToString(guid))
	}

	Expect(len(serviceInstanceGuidsList)).To(Equal(200))

	return serviceInstanceGuidsList
}

func getRandomServiceOfferingGUIDs() []string {
	var serviceOfferingGuidsList []string = nil
	// join service plan visibilities because currently all service plan visibilities are for orgs the user has access to
	serviceOfferingsStatement := fmt.Sprintf("SELECT services.guid FROM services JOIN service_plans ON services.id = service_plans.service_id JOIN service_plan_visibilities ON service_plans.id = service_plan_visibilities.service_plan_id WHERE service_plans.name LIKE '%s-service-plan-%%' ORDER BY %s LIMIT 50", testConfig.GetNamePrefix(), helpers.GetRandomFunction(testConfig))

	serviceOfferingGuids := helpers.ExecuteSelectStatement(ccdb, ctx, serviceOfferingsStatement)
	for _, guid := range serviceOfferingGuids {
		serviceOfferingGuidsList = append(serviceOfferingGuidsList, helpers.ConvertToString(guid))
	}

	Expect(len(serviceOfferingGuidsList)).To(Equal(50))

	return serviceOfferingGuidsList
}
