package roles

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/cf-performance-tests/helpers"
	"github.com/cloudfoundry/cf-test-helpers/v2/workflowhelpers"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega/gmeasure"
)

var _ = Describe("roles", func() {
	Describe("GET /v3/roles", func() {

		Context("as admin", func() {
			It("get all roles", func() {
				experiment := gmeasure.NewExperiment("GET /v3/roles::as admin")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
					experiment.Sample(func(idx int) {
						experiment.MeasureDuration("GET /v3/roles", func() {
							helpers.TimeCFCurl(testConfig.LongTimeout, "/v3/roles")
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})

			It(fmt.Sprintf("get all roles with page size %d", testConfig.LargePageSize), func() {
				experiment := gmeasure.NewExperiment(fmt.Sprintf("GET /v3/roles::as admin with page size %d", testConfig.LargePageSize))
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
					experiment.Sample(func(idx int) {
						experiment.MeasureDuration("GET /v3/roles", func() {
							helpers.TimeCFCurl(testConfig.LongTimeout, fmt.Sprintf("/v3/roles?per_page=%d", testConfig.LargePageSize))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})
		})

		Context("as regular user", func() {
			var orgGuidsList []string
			var spaceGuidsList []string
			BeforeEach(func() {
				orgGuidsList = nil
				selectOrgGuidsStatement := fmt.Sprintf("SELECT guid FROM organizations WHERE name LIKE '%s-org-%%' ORDER BY random() LIMIT 50", testConfig.GetNamePrefix())
				orgGuids := helpers.ExecuteSelectStatement(ccdb, ctx, selectOrgGuidsStatement)
				for _, guid := range orgGuids {
					orgGuidsList = append(orgGuidsList, helpers.ConvertToString(guid))
				}
				spaceGuidsList = nil
				selectSpaceGuidsStatement := fmt.Sprintf("SELECT guid FROM spaces WHERE name LIKE '%s-space-%%' ORDER BY random() LIMIT 50", testConfig.GetNamePrefix())
				spaceGuids := helpers.ExecuteSelectStatement(ccdb, ctx, selectSpaceGuidsStatement)
				for _, guid := range spaceGuids {
					spaceGuidsList = append(spaceGuidsList, helpers.ConvertToString(guid))
				}
			})

			It("get all roles", func() {
				experiment := gmeasure.NewExperiment("GET /v3/roles::as regular user")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.LongTimeout, func() {
					experiment.Sample(func(idx int) {
						experiment.MeasureDuration("GET /v3/roles", func() {
							helpers.TimeCFCurl(testConfig.LongTimeout, "/v3/roles")
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})

			It("filter by types", func() {
				experiment := gmeasure.NewExperiment("GET /v3/roles?types=::as regular user")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.LongTimeout, func() {
					experiment.Sample(func(idx int) {
						experiment.MeasureDuration("GET /v3/roles?types=org_manager,space_developer", func() {
							helpers.TimeCFCurl(testConfig.LongTimeout, "/v3/roles?types=org_manager,space_developer")
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})

			It("filter by orgs and spaces", func() {
				experiment := gmeasure.NewExperiment("GET /v3/roles?organization_guids=&space_guids=::as regular user")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.LongTimeout, func() {
					experiment.Sample(func(idx int) {
						experiment.MeasureDuration("GET /v3/roles?organization_guids=:guids&space_guids=:guids", func() {
							helpers.TimeCFCurl(testConfig.LongTimeout, fmt.Sprintf(
								"/v3/roles?organization_guids=%v&space_guids=%v",
								strings.Join(orgGuidsList[:], ","), strings.Join(spaceGuidsList[:], ",")))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})
		})

	})
})
