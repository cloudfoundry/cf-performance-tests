package audit_events

import (
	"fmt"
	"github.com/cloudfoundry/cf-performance-tests/helpers"
	"github.com/cloudfoundry/cf-test-helpers/v2/workflowhelpers"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega/gmeasure"
)

var eventTypes = "audit.user.organization_user_add,audit.user.organization_billing_manager_remove," +
	"audit.user.space_developer_add,audit.user.organization_auditor_remove,audit.user.space_developer_remove," +
	"audit.user.space_auditor_remove,audit.user.space_auditor_add,audit.user.space_manager_add," +
	"audit.user.organization_billing_manager_add,audit.user.organization_manager_remove," +
	"audit.user.organization_auditor_add,audit.user.space_manager_remove,audit.user.organization_manager_add," +
	"audit.user.organization_user_remove"
var _ = Describe("audit_events", func() {
	Describe("GET /v3/audit_events", func() {

		It("as admin", func() {
			experiment := gmeasure.NewExperiment("GET /v3/audit_events::as admin")
			AddReportEntry(experiment.Name, experiment)

			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/audit_events", func() {
						helpers.TimeCFCurl(testConfig.BasicTimeout, "/v3/audit_events")
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})

		It("as regular user", func() {
			experiment := gmeasure.NewExperiment("GET /v3/audit_events::as regular user")
			AddReportEntry(experiment.Name, experiment)

			workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/audit_events", func() {
						helpers.TimeCFCurl(testConfig.BasicTimeout, "/v3/audit_events")
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})

		It(fmt.Sprintf("as admin with page size %d", testConfig.LargePageSize), func() {
			experiment := gmeasure.NewExperiment(fmt.Sprintf("GET /v3/audit_events::as admin with page size %d", testConfig.LargePageSize))
			AddReportEntry(experiment.Name, experiment)

			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/audit_events", func() {
						helpers.TimeCFCurl(testConfig.LongTimeout, fmt.Sprintf("/v3/audit_events?per_page=%d", testConfig.LargePageSize))
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})

		It(fmt.Sprintf("as admin with types filter and page size 5"), func() {
			experiment := gmeasure.NewExperiment(fmt.Sprintf("GET /v3/audit_events::as admin with types filter & page size 5"))
			AddReportEntry(experiment.Name, experiment)

			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/audit_events", func() {
						helpers.TimeCFCurl(testConfig.LongTimeout, fmt.Sprintf("/v3/audit_events?types=%s&per_page=5&order_by=-created_at", eventTypes))
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})

		It(fmt.Sprintf("as admin with types filter and page size 50"), func() {
			experiment := gmeasure.NewExperiment(fmt.Sprintf("GET /v3/audit_events::as admin with types filter & page size 50"))
			AddReportEntry(experiment.Name, experiment)

			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/audit_events", func() {
						helpers.TimeCFCurl(testConfig.LongTimeout, fmt.Sprintf("/v3/audit_events?types=%s&per_page=50&order_by=-created_at", eventTypes))
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})

		It(fmt.Sprintf("as admin with types filter and page size %d", testConfig.LargePageSize), func() {
			experiment := gmeasure.NewExperiment(fmt.Sprintf("GET /v3/audit_events::as admin with types filter & page size %d", testConfig.LargePageSize))
			AddReportEntry(experiment.Name, experiment)

			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/audit_events", func() {
						helpers.TimeCFCurl(testConfig.LongTimeout, fmt.Sprintf("/v3/audit_events?types=%s&per_page=%d", eventTypes, testConfig.LargePageSize))
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})

		It(fmt.Sprintf("as admin with target_guids"), func() {

			selectAppGuidsStatement := fmt.Sprintf("SELECT guid FROM apps WHERE name LIKE '%s-app-%%' ORDER BY random() LIMIT 1", testConfig.GetNamePrefix())
			appGuids := helpers.ExecuteSelectStatement(ccdb, ctx, selectAppGuidsStatement)
			experiment := gmeasure.NewExperiment(fmt.Sprintf("GET /v3/audit_events::as admin with target_guids &page=1&per_page=5&order_by=-created_at"))
			AddReportEntry(experiment.Name, experiment)

			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/audit_events", func() {
						helpers.TimeCFCurl(testConfig.LongTimeout, fmt.Sprintf("/v3/audit_events?target_guids=%s&page=1&per_page=5&order_by=-created_at", appGuids))
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})

		It(fmt.Sprintf("as admin with created_ats [gt]"), func() {
			experiment := gmeasure.NewExperiment(fmt.Sprintf("GET /v3/audit_events::as admin with created_ats [gt]"))
			AddReportEntry(experiment.Name, experiment)

			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET /v3/audit_events", func() {
						helpers.TimeCFCurl(testConfig.LongTimeout, fmt.Sprintf("/v3/audit_events?types=audit.organization.update&created_ats[gt]=2022-11-14T08:13:01Z"))
					})
				}, gmeasure.SamplingConfig{N: testConfig.Samples})
			})
		})
	})
})
