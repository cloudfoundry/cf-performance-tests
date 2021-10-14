package security_groups

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/cloudfoundry-incubator/cf-performance-tests/helpers"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("security groups", func() {
	Describe("GET /v3/security_groups", func() {
		Measure("as admin", func(b Benchmarker) {
			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
				helpers.TimeCFCurl(b, testConfig.BasicTimeout, "/v3/security_groups")
			})
		}, testConfig.Samples)

		Measure("as regular user", func(b Benchmarker) {
			workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
				helpers.TimeCFCurl(b, testConfig.BasicTimeout, "/v3/security_groups")
			})
		}, testConfig.Samples)

		Measure(fmt.Sprintf("as admin with page size %d", testConfig.LargePageSize), func(b Benchmarker) {
			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
				helpers.TimeCFCurl(b, testConfig.LongTimeout, fmt.Sprintf("/v3/security_groups?per_page=%d", testConfig.LargePageSize))
			})
		}, testConfig.Samples)

		Measure(fmt.Sprintf("as admin with space filter containing %d spaces", testConfig.LargeElementsFilter), func(b Benchmarker) {
			spaceGUIDs := helpers.GetGUIDs(testSetup.AdminUserContext(), testConfig, fmt.Sprintf("/v3/spaces?per_page=%d", testConfig.LargeElementsFilter))
			Expect(spaceGUIDs).NotTo(BeNil())
			Expect(len(spaceGUIDs)).To(Equal(testConfig.LargeElementsFilter))
			spaceGUIDs = helpers.Shuffle(spaceGUIDs)
			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
				helpers.TimeCFCurl(b, testConfig.BasicTimeout, fmt.Sprintf("/v3/security_groups?running_space_guids=%s", strings.Join(spaceGUIDs, ",")))
			})
		}, testConfig.Samples)
	})

	Describe("individually", func() {
		Describe("as admin", func() {
			var securityGroupGUID string
			BeforeEach(func() {
				securityGroupGUIDs := helpers.GetGUIDs(testSetup.AdminUserContext(), testConfig, "/v3/security_groups")
				Expect(securityGroupGUIDs).NotTo(BeNil())
				securityGroupGUID = securityGroupGUIDs[rand.Intn(len(securityGroupGUIDs))]
			})

			Measure("GET /v3/security_groups/:guid", func(b Benchmarker) {
				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					helpers.TimeCFCurl(b, testConfig.BasicTimeout, fmt.Sprintf("/v3/security_groups/%s", securityGroupGUID))
				})
			}, testConfig.Samples)

			Measure("PATCH /v3/security_groups/:guid", func(b Benchmarker) {
				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					data := fmt.Sprintf(`{"name":"perf-updated-security-group-%s"}`, securityGroupGUID)
					helpers.TimeCFCurl(b, testConfig.BasicTimeout, "-X", "PATCH", "-d", data, fmt.Sprintf("/v3/security_groups/%s", securityGroupGUID))
				})
			}, testConfig.Samples)

			Measure("DELETE /v3/security_groups/:guid", func(b Benchmarker) {
				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					helpers.TimeCFCurl(b, testConfig.BasicTimeout, "-X", "DELETE", fmt.Sprintf("/v3/security_groups/%s", securityGroupGUID))

					// Wait until "GET /v3/security_groups/:guid" fails.
					helpers.WaitToFail(testSetup.AdminUserContext(), testConfig, fmt.Sprintf("/v3/security_groups/%s", securityGroupGUID))
				})
			}, testConfig.Samples)
		})

		Describe("as regular user", func() {
			Measure("GET /v3/security_groups/:guid", func(b Benchmarker) {
				securityGroupGUIDs := helpers.GetGUIDs(testSetup.RegularUserContext(), testConfig, "/v3/security_groups")
				Expect(securityGroupGUIDs).NotTo(BeNil())
				securityGroupGUID := securityGroupGUIDs[rand.Intn(len(securityGroupGUIDs))]
				workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
					helpers.TimeCFCurl(b, testConfig.BasicTimeout, fmt.Sprintf("/v3/security_groups/%s", securityGroupGUID))
				})
			}, testConfig.Samples)
		})
	})
})
