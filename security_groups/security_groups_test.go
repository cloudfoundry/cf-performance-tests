package security_groups

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-performance-tests/helpers"
)

var _ = Describe("security groups", func() {
	Describe("GET /v3/security_groups", func() {
		Measure("as admin", func(b Benchmarker) {
			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
				result := cf.Cf(
					"curl", "--fail", "/v3/security_groups", "-v",
				).Wait(testConfig.BasicTimeout)
				Expect(
					result,
				).To(Exit(0))

				runtime := helpers.GetXRuntimeHeader(result.Out.Contents())
				b.RecordValue("request time", runtime)
			})
		}, testConfig.Samples)

		Measure("as regular user", func(b Benchmarker) {
			workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
				result := cf.Cf(
					"curl", "--fail", "/v3/security_groups", "-v",
				).Wait(testConfig.BasicTimeout)
				Expect(
					result,
				).To(Exit(0))

				runtime := helpers.GetXRuntimeHeader(result.Out.Contents())
				b.RecordValue("request time", runtime)
			})
		}, testConfig.Samples)

		Measure("as admin with large page size", func(b Benchmarker) {
			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
				result := cf.Cf(
					"curl", "--fail", fmt.Sprintf("/v3/security_groups?per_page=%d", testConfig.LargePageSize), "-v",
				).Wait(testConfig.LongTimeout)
				Expect(
					result,
				).To(Exit(0))

				runtime := helpers.GetXRuntimeHeader(result.Out.Contents())
				b.RecordValue("request time", runtime)
			})
		}, testConfig.Samples)

		Measure("as admin with space filter", func(b Benchmarker) {
			spaceGUIDs := helpers.GetGUIDs(testSetup.AdminUserContext(), testConfig, fmt.Sprintf("/v3/spaces?per_page=%d", testConfig.LargeElementsFilter))
			Expect(spaceGUIDs).NotTo(BeNil())
			Expect(len(spaceGUIDs)).To(Equal(testConfig.LargeElementsFilter))
			spaceGUIDs = helpers.Shuffle(spaceGUIDs)
			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
				result := cf.Cf(
					"curl", "--fail", fmt.Sprintf("/v3/security_groups?running_space_guids=%s", strings.Join(spaceGUIDs, ",")), "-v",
				).Wait(testConfig.BasicTimeout)
				Expect(
					result,
				).To(Exit(0))

				runtime := helpers.GetXRuntimeHeader(result.Out.Contents())
				b.RecordValue("request time", runtime)
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
					result := cf.Cf(
						"curl", "--fail", fmt.Sprintf("/v3/security_groups/%s", securityGroupGUID), "-v",
					).Wait(testConfig.BasicTimeout)
					Expect(
						result,
					).To(Exit(0))

					runtime := helpers.GetXRuntimeHeader(result.Out.Contents())
					b.RecordValue("request time", runtime)
				})
			}, testConfig.Samples)

			Measure("PATCH /v3/security_groups/:guid", func(b Benchmarker) {
				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					data := fmt.Sprintf(`{"name":"perf-updated-security-group-%s"}`, securityGroupGUID)
					result := cf.Cf(
						"curl", "--fail", "-X", "PATCH", "-d", data, fmt.Sprintf("/v3/security_groups/%s", securityGroupGUID), "-v",
					).Wait(testConfig.BasicTimeout)
					Expect(
						result,
					).To(Exit(0))

					runtime := helpers.GetXRuntimeHeader(result.Out.Contents())
					b.RecordValue("request time", runtime)
				})
			}, testConfig.Samples)

			Measure("DELETE /v3/security_groups/:guid", func(b Benchmarker) {
				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					result := cf.Cf(
						"curl", "--fail", "-X", "DELETE", fmt.Sprintf("/v3/security_groups/%s", securityGroupGUID), "-v",
					).Wait(testConfig.BasicTimeout)
					Expect(
						result,
					).To(Exit(0))

					runtime := helpers.GetXRuntimeHeader(result.Out.Contents())
					b.RecordValue("request time", runtime)

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
					result := cf.Cf(
						"curl", "--fail", fmt.Sprintf("/v3/security_groups/%s", securityGroupGUID), "-v",
					).Wait(testConfig.BasicTimeout)
					Expect(
						result,
					).To(Exit(0))

					runtime := helpers.GetXRuntimeHeader(result.Out.Contents())
					b.RecordValue("request time", runtime)
				})
			}, testConfig.Samples)
		})
	})
})
