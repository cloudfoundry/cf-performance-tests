package domains

import (
	"fmt"
	"math/rand"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-performance-tests/helpers"
)

var _ = Describe("domains", func() {
	Describe("GET /v3/domains", func() {
		Measure("as admin", func(b Benchmarker) {
			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
				result := cf.Cf(
					"curl", "--fail", "/v3/domains", "-v",
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
					"curl", "--fail", "/v3/domains", "-v",
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
					"curl", "--fail", fmt.Sprintf("/v3/domains?per_page=%d", testConfig.LargePageSize), "-v",
				).Wait(testConfig.LongTimeout)
				Expect(
					result,
				).To(Exit(0))

				runtime := helpers.GetXRuntimeHeader(result.Out.Contents())
				b.RecordValue("request time", runtime)
			})
		}, testConfig.Samples)
	})

	Describe("GET /v3/organizations/:guid/domains", func() {
		Measure("as admin", func(b Benchmarker) {
			orgGUIDs := helpers.GetGUIDs(testSetup.AdminUserContext(), testConfig, "/v3/organizations")
			Expect(orgGUIDs).NotTo(BeNil())
			orgGUID := orgGUIDs[rand.Intn(len(orgGUIDs))]
			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
				result := cf.Cf(
					"curl", "--fail", fmt.Sprintf("/v3/organizations/%s/domains", orgGUID), "-v",
				).Wait(testConfig.BasicTimeout)
				Expect(
					result,
				).To(Exit(0))

				runtime := helpers.GetXRuntimeHeader(result.Out.Contents())
				b.RecordValue("request time", runtime)
			})
		}, testConfig.Samples)

		Measure("as regular user", func(b Benchmarker) {
			orgGUIDs := helpers.GetGUIDs(testSetup.RegularUserContext(), testConfig, "/v3/organizations")
			Expect(orgGUIDs).NotTo(BeNil())
			orgGUID := orgGUIDs[rand.Intn(len(orgGUIDs))]
			workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
				result := cf.Cf(
					"curl", "--fail", fmt.Sprintf("/v3/organizations/%s/domains", orgGUID), "-v",
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
			var domainGUID string
			BeforeEach(func() {
				domainGUIDs := helpers.GetGUIDs(testSetup.AdminUserContext(), testConfig, "/v3/domains")
				Expect(domainGUIDs).NotTo(BeNil())
				domainGUID = domainGUIDs[rand.Intn(len(domainGUIDs))]
			})

			Measure("GET /v3/domains/:guid", func(b Benchmarker) {
				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					result := cf.Cf(
						"curl", "--fail", fmt.Sprintf("/v3/domains/%s", domainGUID), "-v",
					).Wait(testConfig.BasicTimeout)
					Expect(
						result,
					).To(Exit(0))

					runtime := helpers.GetXRuntimeHeader(result.Out.Contents())
					b.RecordValue("request time", runtime)
				})
			}, testConfig.Samples)

			Measure("PATCH /v3/domains/:guid", func(b Benchmarker) {
				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					data := `{ "metadata": { "annotations": { "test": "PATCH /v3/domains/:guid" } } }`
					result := cf.Cf(
						"curl", "--fail", "-X", "PATCH", "-d", data, fmt.Sprintf("/v3/domains/%s", domainGUID), "-v",
					).Wait(testConfig.BasicTimeout)
					Expect(
						result,
					).To(Exit(0))

					runtime := helpers.GetXRuntimeHeader(result.Out.Contents())
					b.RecordValue("request time", runtime)
				})
			}, testConfig.Samples)

			Measure("DELETE /v3/domains/:guid", func(b Benchmarker) {
				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					result := cf.Cf(
						"curl", "--fail", "-X", "DELETE", fmt.Sprintf("/v3/domains/%s", domainGUID), "-v",
					).Wait(testConfig.BasicTimeout)
					Expect(
						result,
					).To(Exit(0))

					runtime := helpers.GetXRuntimeHeader(result.Out.Contents())
					b.RecordValue("request time", runtime)

					// Wait until "GET /v3/domains/:guid" fails.
					helpers.WaitToFail(testSetup.AdminUserContext(), testConfig, fmt.Sprintf("/v3/domains/%s", domainGUID))
				})
			}, testConfig.Samples)
		})

		Describe("as regular user", func() {
			Measure("GET /v3/domains/:guid", func(b Benchmarker) {
				domainGUIDs := helpers.GetGUIDs(testSetup.RegularUserContext(), testConfig, "/v3/domains")
				Expect(domainGUIDs).NotTo(BeNil())
				domainGUID := domainGUIDs[rand.Intn(len(domainGUIDs))]
				workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
					result := cf.Cf(
						"curl", "--fail", fmt.Sprintf("/v3/domains/%s", domainGUID), "-v",
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
