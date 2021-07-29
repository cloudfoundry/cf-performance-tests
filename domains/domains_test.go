package domains

import (
	"fmt"
	"math/rand"

	"github.com/cloudfoundry-incubator/cf-performance-tests/helpers"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)


var _ = Describe("domains", func() {
	Describe("GET /v3/domains", func() {
		Measure("as admin", func(b Benchmarker) {
			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
				b.Time("request time", func() {
					Expect(cf.Cf("curl", "--fail", "/v3/domains").Wait(testConfig.BasicTimeout)).To(Exit(0))
				})
			})
		}, testConfig.Samples)

		Measure("as regular user", func(b Benchmarker) {
			workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
				b.Time("request time", func() {
					Expect(cf.Cf("curl", "--fail", "/v3/domains").Wait(testConfig.BasicTimeout)).To(Exit(0))
				})
			})
		}, testConfig.Samples)

		Measure("as admin with large page size", func(b Benchmarker) {
			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
				b.Time("request time", func() {
					Expect(cf.Cf("curl", "--fail", fmt.Sprintf("/v3/domains?per_page=%d", testConfig.LargePageSize)).Wait(testConfig.LongTimeout)).To(Exit(0))
				})
			})
		}, testConfig.Samples)
	})

	Describe("GET /v3/organizations/:guid/domains", func() {
		var orgs []string
		BeforeEach(func() {
			orgs = helpers.GetGUIDs(testSetup.AdminUserContext(), testConfig, "/v3/organizations")
		})
		Measure("as admin", func(b Benchmarker) {
			org := orgs[rand.Intn(len(orgs))]
			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
				b.Time("request time", func() {
					Expect(cf.Cf("curl", "--fail", fmt.Sprintf("/v3/organizations/%s/domains", org)).Wait(testConfig.BasicTimeout)).To(Exit(0))
				})
			})
		}, testConfig.Samples)
	})

	Describe("individually", func() {
		var domains []string
		BeforeEach(func() {
			domains = helpers.GetGUIDs(testSetup.AdminUserContext(), testConfig, "/v3/domains")
		})

		Measure("GET /v3/domains/:guid", func(b Benchmarker) {
			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
				domain := domains[rand.Intn(len(domains))]
				b.Time("request time", func() {
					Expect(cf.Cf("curl", "--fail", fmt.Sprintf("/v3/domains/%s", domain)).Wait(testConfig.BasicTimeout)).To(Exit(0))
				})
			})
		}, testConfig.Samples)

		PMeasure("PATCH /v3/domains/:guid", func(b Benchmarker) {
			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
				domain := domains[rand.Intn(len(domains))]
				b.Time("request time", func() {
					Expect(cf.Cf(
						"curl", "--fail", "-X", "PATCH",
						"-d", `{ "metadata": { "annotations": { "test": "PATCH /v3/domains/:guid" } } }`,
						fmt.Sprintf("/v3/domains/%s", domain)).Wait(testConfig.BasicTimeout)).To(Exit(0))
				})
			})
		}, testConfig.Samples)
	})
})
