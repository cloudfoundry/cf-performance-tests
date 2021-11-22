package service_plans

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/cloudfoundry-incubator/cf-performance-tests/helpers"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("service plans", func() {
	Describe("GET /v3/service_plans", func() {
		Context("as admin", func() {
			Measure("list all", func(b Benchmarker) {
				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					helpers.TimeCFCurl(b, testConfig.BasicTimeout, "/v3/service_plans")
				})
			}, testConfig.Samples)
			Measure(fmt.Sprintf("list with page size %d", testConfig.LargePageSize), func(b Benchmarker) {
				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
					helpers.TimeCFCurl(b, testConfig.LongTimeout, fmt.Sprintf("/v3/service_plans?per_page=%d", testConfig.LargePageSize))
				})
			}, testConfig.Samples)
		})
		Context("as regular user", func() {
			Measure("list all", func(b Benchmarker) {
				workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.LongTimeout, func() {
					helpers.TimeCFCurl(b, testConfig.LongTimeout, "/v3/service_plans")
				})
			}, testConfig.Samples)
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
			Measure("show one", func(b Benchmarker) {
				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					helpers.TimeCFCurl(b, testConfig.BasicTimeout, fmt.Sprintf("/v3/service_plans/%s", servicePlanGUID))
				})
			}, testConfig.Samples)

		})
		Context("as regular user", func() {
			BeforeEach(func() {
				servicePlanGUID = getRandomLimitedServicePlanGuid()
			})
			Measure("show one", func(b Benchmarker) {
				workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.LongTimeout, func() {
					helpers.TimeCFCurl(b, testConfig.LongTimeout, fmt.Sprintf("/v3/service_plans/%s", servicePlanGUID))
				})
			}, testConfig.Samples)
		})
	})

	Describe("GET /v3/service_plans/:guid/visibility", func() {
		var servicePlanGUID string
		BeforeEach(func() {
			servicePlanGUID = getRandomLimitedServicePlanGuid()
		})
		Context("as admin", func() {
			Measure("show visibility", func(b Benchmarker) {
				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					helpers.TimeCFCurl(b, testConfig.BasicTimeout, fmt.Sprintf("/v3/service_plans/%s/visibility", servicePlanGUID))
				})
			}, testConfig.Samples)
		})
		Context("as regular user", func() {
			Measure("show visibility", func(b Benchmarker) {
				workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
					helpers.TimeCFCurl(b, testConfig.BasicTimeout, fmt.Sprintf("/v3/service_plans/%s/visibility", servicePlanGUID))
				})
			}, testConfig.Samples)
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
			Measure("filter for list of service_offerings", func(b Benchmarker) {
				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					helpers.TimeCFCurl(b, testConfig.BasicTimeout, fmt.Sprintf(
						"/v3/service_plans?service_offering_guids=%v", strings.Join(serviceOfferingGuidsList[:], ",")))
				})
			}, testConfig.Samples)
			Measure(fmt.Sprintf("filter for list of service_offerings with page size %d", testConfig.LargePageSize), func(b Benchmarker) {
				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
					helpers.TimeCFCurl(b, testConfig.LongTimeout, fmt.Sprintf(
						"/v3/service_plans?service_offering_guids=%v&per_page=%d",
						strings.Join(serviceOfferingGuidsList[:], ","), testConfig.LargePageSize))
				})
			}, testConfig.Samples)
		})
		Context("as regular user", func() {
			Measure("filter for list of service_offerings", func(b Benchmarker) {
				workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
					helpers.TimeCFCurl(b, testConfig.BasicTimeout, fmt.Sprintf(
						"/v3/service_plans?service_offering_guids=%v", strings.Join(serviceOfferingGuidsList[:], ",")))
				})
			}, testConfig.Samples)
		})
	})
})

func getRandomLimitedServicePlanGuid() string {
	servicePlanGUIDs := helpers.ExecuteSelectStatement(ccdb, ctx,
		"SELECT guid FROM service_plans WHERE id IN (SELECT service_plan_id FROM service_plan_visibilities ORDER BY random() LIMIT 1)")
	return servicePlanGUIDs[0].(string)
}

