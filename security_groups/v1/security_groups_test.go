package security_groups

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/cf-test-helpers/v2/workflowhelpers"
	. "github.com/onsi/ginkgo/v2"

	"github.com/onsi/gomega/gmeasure"

	"github.com/cloudfoundry/cf-performance-tests/helpers"
)

var _ = Describe("security groups", func() {
	Describe("GET /v3/security_groups", func() {

		It("as admin", func() {
			experiment := gmeasure.NewExperiment("GET /v3/security_groups::as admin")
			AddReportEntry(experiment.Name, experiment)

			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/security_groups", func() {
						helpers.TimeCFCurl(testConfig.BasicTimeout, "/v3/security_groups")
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})

		It("as regular user", func() {
			experiment := gmeasure.NewExperiment("GET /v3/security_groups::as regular user")
			AddReportEntry(experiment.Name, experiment)

			workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/security_groups", func() {
						helpers.TimeCFCurl(testConfig.BasicTimeout, "/v3/security_groups")
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})

		It(fmt.Sprintf("as admin with page size %d", testConfig.LargePageSize), func() {
			experiment := gmeasure.NewExperiment(fmt.Sprintf("GET /v3/security_groups::as admin with page size %d", testConfig.LargePageSize))
			AddReportEntry(experiment.Name, experiment)

			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/security_groups", func() {
						helpers.TimeCFCurl(testConfig.LongTimeout, fmt.Sprintf("/v3/security_groups?per_page=%d", testConfig.LargePageSize))
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})

		//selected space guids could contain spaces without any security groups -> only half of spaces have security groups assigned
		It(fmt.Sprintf("as admin with space filter containing %d spaces", testConfig.LargeElementsFilter), func() {
			experiment := gmeasure.NewExperiment(fmt.Sprintf("GET /v3/security_groups::as admin with space filter containing %d spaces", testConfig.LargeElementsFilter))
			AddReportEntry(experiment.Name, experiment)

			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
				experiment.Sample(func(idx int) {
					spaceGUIDs := getRandomSpacesWithSecurityGroups()

					experiment.MeasureDuration("GET /v3/security_groups", func() {
						helpers.TimeCFCurl(testConfig.LongTimeout, fmt.Sprintf("/v3/security_groups?running_space_guids=%s", strings.Join(spaceGUIDs, ",")))
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})
	})

	Describe("individually", func() {
		Describe("as admin", func() {
			It("gets /v3/security_groups/:guid as admin", func() {
				experiment := gmeasure.NewExperiment("individually::as admin::GET /v3/security_groups/:guid")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						securityGroupGUID := getRandomSecurityGroup()

						experiment.MeasureDuration("GET /v3/security_groups/:guid", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf("/v3/security_groups/%s", securityGroupGUID))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})

			It("patches /v3/security_groups/:guid as admin", func() {
				experiment := gmeasure.NewExperiment("individually::as admin::PATCH /v3/security_groups/:guid")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						securityGroupGUID := getRandomSecurityGroup()

						experiment.MeasureDuration("PATCH /v3/security_groups/:guid", func() {
							data := fmt.Sprintf(`{"name":"%s-updated-security-group-%s"}`, testConfig.GetNamePrefix(), securityGroupGUID)
							helpers.TimeCFCurl(testConfig.BasicTimeout, "-X", "PATCH", "-d", data, fmt.Sprintf("/v3/security_groups/%s", securityGroupGUID))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})

			It("deletes /v3/security_groups/:guid as admin", func() {
				experiment := gmeasure.NewExperiment("individually::as admin::DELETE /v3/security_groups/:guid")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						securityGroupGUID := getRandomSecurityGroup()

						experiment.MeasureDuration("DELETE /v3/security_groups/:guid", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, "-X", "DELETE", fmt.Sprintf("/v3/security_groups/%s", securityGroupGUID))
						})

						helpers.WaitToFail(testSetup.AdminUserContext(), testConfig, fmt.Sprintf("/v3/security_groups/%s", securityGroupGUID))
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})
		})

		Describe("as regular user", func() {
			It("gets /v3/security_groups/:guid as regular user", func() {
				experiment := gmeasure.NewExperiment("individually::as regular user::GET /v3/security_groups/:guid")
				AddReportEntry(experiment.Name, experiment)

				workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
					experiment.Sample(func(idx int) {
						securityGroupGUID := getRandomSecurityGroup()

						experiment.MeasureDuration("GET /v3/security_groups/:guid", func() {
							helpers.TimeCFCurl(testConfig.BasicTimeout, fmt.Sprintf("/v3/security_groups/%s", securityGroupGUID))
						})
					}, gmeasure.SamplingConfig{N: testConfig.Samples})
				})
			})
		})
	})
})

func getRandomSecurityGroup() string {
	var securityGroupGuid string
	securityGroupStatement := fmt.Sprintf("SELECT guid FROM security_groups JOIN security_groups_spaces ON security_groups.id = security_groups_spaces.security_group_id ORDER BY %s LIMIT 1", helpers.GetRandomFunction(testConfig))
	securityGroupGuids := helpers.ExecuteSelectStatement(ccdb, ctx, securityGroupStatement)
	for _, guid := range securityGroupGuids {
		securityGroupGuid = helpers.ConvertToString(guid)
	}

	return helpers.ConvertToString(securityGroupGuid)
}

func getRandomSpacesWithSecurityGroups() []string {
	var spaceGuidsList []string = nil
	spaceStatement := fmt.Sprintf("SELECT guid FROM spaces JOIN security_groups_spaces ON spaces.id = security_groups_spaces.space_id ORDER BY %s LIMIT %d", helpers.GetRandomFunction(testConfig), testConfig.LargeElementsFilter)
	spaceGuids := helpers.ExecuteSelectStatement(ccdb, ctx, spaceStatement)

	for _, guid := range spaceGuids {
		spaceGuidsList = append(spaceGuidsList, helpers.ConvertToString(guid))
	}

	return spaceGuidsList
}
