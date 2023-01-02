package audit_events

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"log"
	"testing"

	"github.com/cloudfoundry/cf-test-helpers/v2/workflowhelpers"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cf-performance-tests/helpers"
)

var testConfig = helpers.NewConfig()
var prefix string
var testSetup *workflowhelpers.ReproducibleTestSuiteSetup
var ccdb *sql.DB
var uaadb *sql.DB
var ctx context.Context

const (
	// main test parameters:
	orgs   = 4
	spaces = 1
)

var _ = BeforeSuite(func() {

	testSetup = workflowhelpers.NewTestSuiteSetup(&testConfig)
	testSetup.Setup()
	prefix = testConfig.GetNamePrefix()
	ccdb, uaadb, ctx = helpers.OpenDbConnections(testConfig)
	helpers.ImportStoredProcedures(ccdb, ctx, testConfig)

	// create orgs
	createOrgStatement := fmt.Sprintf("create_orgs(%d)", orgs)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createOrgStatement, testConfig)

	//create spaces
	createSpacesStatement := fmt.Sprintf("create_spaces(%d)", spaces)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createSpacesStatement, testConfig)

	// create events helper table
	createEventsTableStatement := fmt.Sprintf("create_event_types_table()")
	helpers.ExecuteStoredProcedure(ccdb, ctx, createEventsTableStatement, testConfig)

	// create events
	createEventStatement := fmt.Sprintf("create_events()")
	helpers.ExecuteStoredProcedure(ccdb, ctx, createEventStatement, testConfig)

	// create apps table entries
	createApps(2)

	// assign the regular user to all orgs
	regularUserGUID := helpers.GetUserGUID(testSetup.RegularUserContext(), testConfig)
	orgsAssignedToRegularUser := orgs
	assignUserAsOrgManager := fmt.Sprintf("assign_user_as_org_manager('%s', %d)", regularUserGUID, orgsAssignedToRegularUser)
	helpers.ExecuteStoredProcedure(ccdb, ctx, assignUserAsOrgManager, testConfig)

	helpers.AnalyzeDB(ccdb, ctx, testConfig)
})

func createApps(numApps int) {
	i := 1
	for i < numApps {
		i++
		appGuid := uuid.NewString()
		appName := fmt.Sprintf("%s-app-%s", prefix, appGuid)
		createSpaceStatement := fmt.Sprintf(
			"INSERT INTO apps (id,guid, name) VALUES ('%d','%s', '%s')",
			i, appGuid, appName)
		helpers.ExecuteInsertStatement(ccdb, ctx, createSpaceStatement, testConfig)
	}
}

var _ = AfterSuite(func() {
	helpers.CleanupTestData(ccdb, uaadb, ctx, testConfig)

	err := ccdb.Close()
	if err != nil {
		log.Print(err)
	}

	if uaadb != nil {
		err = uaadb.Close()
		if err != nil {
			log.Print(err)
		}
	}
})

var _ = ReportAfterSuite("Audit events test suite", func(report types.Report) {
	helpers.GenerateReports(helpers.ConfigureJsonReporter(&testConfig, "audit-events", "audit events"), report)
})

func TestAuditEvents(t *testing.T) {
	helpers.LoadConfig(&testConfig)
	RegisterFailHandler(Fail)
	RunSpecs(t, "AuditEventsTest Suite")
}
