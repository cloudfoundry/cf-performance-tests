package cf_performance_tests

import (
	"fmt"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Security Groups", func() {
	Describe("GET /v3/security_groups", func() {
		Measure("as admin", func(b Benchmarker) {
			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
				b.Time("request time", func() {
					Expect(cf.Cf("curl", "/v3/security_groups").Wait(testConfig.BasicTimeout)).To(Exit(0))
				})
			})
		}, 10)

		Measure("as regular user", func(b Benchmarker) {
			workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
				b.Time("request time", func() {
					Expect(cf.Cf("curl", "/v3/security_groups").Wait(testConfig.BasicTimeout)).To(Exit(0))
				})
			})
		}, 10)

		Measure("as admin with large page size", func(b Benchmarker) {
			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
				b.Time("request time", func() {
					Expect(cf.Cf("curl", fmt.Sprintf("/v3/security_groups?per_page=%d", testConfig.LargePageSize)).Wait(testConfig.LongTimeout)).To(Exit(0))
				})
			})
		}, 10)
	})
})
