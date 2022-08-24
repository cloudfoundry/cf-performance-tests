package security_groups

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

var _ = Describe("security groups", func() {
	Describe("GET /v3/security_groups", func() {

		It("gets /v3/security_groups as admin efficiently", func() {
			experiment := gmeasure.NewExperiment("as admin")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("GET /v3/security_groups", func() {
					workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
						helpers.V2TimeCFCurl(testConfig.BasicTimeout, "/v3/security_groups")
					})
				})
			}, gmeasure.SamplingConfig{N: testConfig.Samples})
		})

		It("gets /v3/security_groups as a regular user efficiently", func() {
			experiment := gmeasure.NewExperiment("as user")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("GET /v3/security_groups", func() {
					workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
						helpers.V2TimeCFCurl(testConfig.BasicTimeout, "/v3/security_groups")
					})
				})
			}, gmeasure.SamplingConfig{N: testConfig.Samples})
		})

		It(fmt.Sprintf("gets /v3/security_groups as admin with page size %d efficiently", testConfig.LargePageSize), func() {
			experiment := gmeasure.NewExperiment(fmt.Sprintf("as admin with page size %d", testConfig.LargePageSize))
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration(fmt.Sprintf("GET /v3/security_groups"), func() {
					workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
						helpers.V2TimeCFCurl(testConfig.LongTimeout, fmt.Sprintf("/v3/security_groups?per_page=%d", testConfig.LargePageSize))
					})
				})
			}, gmeasure.SamplingConfig{N: testConfig.Samples})
		})

		It(fmt.Sprintf("gets /v3/security_groups as admin with space filter containing %d spaces", testConfig.LargeElementsFilter), func() {
			spaceGUIDs := helpers.GetGUIDs(testSetup.AdminUserContext(), testConfig, fmt.Sprintf("/v3/spaces?per_page=%d&label_selector=%s", testConfig.LargeElementsFilter, testConfig.TestResourcePrefix))
			Expect(spaceGUIDs).NotTo(BeNil())
			Expect(len(spaceGUIDs)).To(Equal(testConfig.LargeElementsFilter))
			spaceGUIDs = helpers.Shuffle(spaceGUIDs)

			experiment := gmeasure.NewExperiment(fmt.Sprintf("as admin with space filter containing %d spaces", testConfig.LargeElementsFilter))
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration(fmt.Sprintf("GET /v3/security_groups"), func() {
					workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
						helpers.V2TimeCFCurl(testConfig.LongTimeout, fmt.Sprintf("/v3/security_groups?running_space_guids=%s", strings.Join(spaceGUIDs, ",")))
					})
				})
			}, gmeasure.SamplingConfig{N: testConfig.Samples})
		})
	})

	Describe("individually", func() {
		Describe("as admin", func() {
			var securityGroupGUID string
			BeforeEach(func() {
				securityGroupGUIDs := helpers.GetGUIDs(testSetup.AdminUserContext(), testConfig, "/v3/security_groups")
				Expect(securityGroupGUIDs).NotTo(BeNil())
				securityGroupGUID = securityGroupGUIDs[rand.Intn(len(securityGroupGUIDs))]
			})

			It("gets /v3/security_groups/:guid as admin efficiently", func() {
				experiment := gmeasure.NewExperiment("as admin")
				AddReportEntry(experiment.Name, experiment)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/security_groups/:guid", func() {
						workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
							helpers.V2TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf("/v3/security_groups/%s", securityGroupGUID))
						})
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})

			It("patches /v3/security_groups/:guid as admin efficiently", func() {
				experiment := gmeasure.NewExperiment("as admin")
				AddReportEntry(experiment.Name, experiment)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("PATCH /v3/security_groups/:guid", func() {
						workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
							data := fmt.Sprintf(`{"name":"%s-updated-security-group-%s"}`, testConfig.GetNamePrefix(), securityGroupGUID)
							helpers.V2TimeCFCurl(testConfig.BasicTimeout, "-X", "PATCH", "-d", data, fmt.Sprintf("/v3/security_groups/%s", securityGroupGUID))
						})
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})

			It("deletes /v3/security_groups/:guid as admin efficiently", func() {
				experiment := gmeasure.NewExperiment("as admin")
				AddReportEntry(experiment.Name, experiment)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("DELETE /v3/security_groups/:guid", func() {
						workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
							securityGroupGUIDs := helpers.GetGUIDs(testSetup.AdminUserContext(), testConfig, "/v3/security_groups")
							Expect(securityGroupGUIDs).NotTo(BeNil())
							securityGroupGUID = securityGroupGUIDs[rand.Intn(len(securityGroupGUIDs))]

							helpers.V2TimeCFCurl(testConfig.BasicTimeout, "-X", "DELETE", fmt.Sprintf("/v3/security_groups/%s", securityGroupGUID))

							helpers.WaitToFail(testSetup.AdminUserContext(), testConfig, fmt.Sprintf("/v3/security_groups/%s", securityGroupGUID))
						})
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})

		Describe("as regular user", func() {
			It("gets /v3/security_groups/:guid as regular user efficiently", func() {
				securityGroupGUIDs := helpers.GetGUIDs(testSetup.RegularUserContext(), testConfig, "/v3/security_groups")
				Expect(securityGroupGUIDs).NotTo(BeNil())
				securityGroupGUID := securityGroupGUIDs[rand.Intn(len(securityGroupGUIDs))]

				experiment := gmeasure.NewExperiment("as regular user")
				AddReportEntry(experiment.Name, experiment)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/security_groups/:guid", func() {
						workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
							helpers.V2TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf("/v3/security_groups/%s", securityGroupGUID))
						})
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})
	})
})
