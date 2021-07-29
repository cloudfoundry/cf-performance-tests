package security_groups

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/cloudfoundry-incubator/cf-performance-tests/helpers"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("security groups", func() {
	Describe("GET /v3/security_groups", func() {
		Measure("as admin", func(b Benchmarker) {
			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
				b.Time("request time", func() {
					Expect(cf.Cf("curl", "--fail", "/v3/security_groups").Wait(testConfig.BasicTimeout)).To(Exit(0))
				})
			})
		}, testConfig.Samples)

		Measure("as regular user", func(b Benchmarker) {
			workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
				b.Time("request time", func() {
					Expect(cf.Cf("curl", "--fail", "/v3/security_groups").Wait(testConfig.BasicTimeout)).To(Exit(0))
				})
			})
		}, testConfig.Samples)

		Measure("as admin with large page size", func(b Benchmarker) {
			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
				b.Time("request time", func() {
					Expect(cf.Cf("curl", "--fail", fmt.Sprintf("/v3/security_groups?per_page=%d", testConfig.LargePageSize)).Wait(testConfig.LongTimeout)).To(Exit(0))
				})
			})
		}, testConfig.Samples)

		Measure("as admin with space filter", func(b Benchmarker) {
			spaces := helpers.GetGUIDs(testSetup.AdminUserContext(), testConfig, "/v3/spaces")
			spaceGUIDs := helpers.SelectRandom(spaces, 20)
			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
				b.Time("request time", func() {
					Expect(cf.Cf("curl", "--fail", fmt.Sprintf("/v3/security_groups?running_space_guids=%s", strings.Join(spaceGUIDs, ","))).Wait(testConfig.BasicTimeout)).To(Exit(0))
				})
			})
		}, testConfig.Samples)
	})

	Describe("individually", func() {
		var securityGroups []string
		BeforeEach(func() {
			securityGroups = helpers.GetGUIDs(testSetup.AdminUserContext(), testConfig, "/v3/security_groups")
		})

		Measure("GET /v3/security_groups/:guid", func(b Benchmarker) {
			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
				sg := securityGroups[rand.Intn(len(securityGroups))]
				b.Time("request time", func() {
					Expect(cf.Cf("curl", "--fail", fmt.Sprintf("/v3/security_groups/%s", sg)).Wait(testConfig.BasicTimeout)).To(Exit(0))
				})
			})
		}, testConfig.Samples)

		PMeasure("PATCH /v3/security_groups/:guid", func(b Benchmarker) {
			updateFormat := `{"name":"updated-security-group-%s","rules":[{"protocol":"tcp","destination":"10.10.10.0/24","ports":"443,80,8080"},{"protocol":"icmp","destination":"10.10.10.0/24","type":8,"code":0,"description":"ping"}]}`
			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
				sg := securityGroups[rand.Intn(len(securityGroups))]
				b.Time("request time", func() {
					Expect(cf.Cf(
						"curl", "--fail", "-X", "PATCH",
						"-d", fmt.Sprintf(updateFormat, sg),
						fmt.Sprintf("/v3/security_groups/%s", sg)).Wait(testConfig.BasicTimeout)).To(Exit(0))
				})
			})
		}, testConfig.Samples)

		PMeasure("DELETE /v3/security_groups/:guid", func(b Benchmarker) {
			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
				sg := securityGroups[rand.Intn(len(securityGroups))]
				b.Time("request time", func() {
					Expect(cf.Cf(
						"curl", "--fail", "-X", "DELETE",
						fmt.Sprintf("/v3/security_groups/%s", sg)).Wait(testConfig.BasicTimeout)).To(Exit(0))
				})
			})
		}, testConfig.Samples)
	})
})
