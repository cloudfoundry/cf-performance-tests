package users

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"

	"github.com/google/uuid"

	"github.com/cloudfoundry/cf-test-helpers/v2/workflowhelpers"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cf-performance-tests/helpers"
)

var testConfig = helpers.NewConfig()
var testSetup *workflowhelpers.ReproducibleTestSuiteSetup
var ccdb *sql.DB
var uaadb *sql.DB
var ctx context.Context
var org_guid = uuid.NewString()
var space_guid = uuid.NewString()

const test_version = "v1"

const (
	// main test parameters:
	users = 10000
)

var _ = BeforeSuite(func() {
	testSetup = workflowhelpers.NewTestSuiteSetup(&testConfig)
	testSetup.Setup()
	ccdb, uaadb, ctx = helpers.OpenDbConnections(testConfig)
	helpers.ImportStoredProcedures(ccdb, ctx, testConfig)

	// create users with org and space roles
	createUsersWithOrgAndSpaceRolesStatement := fmt.Sprintf("create_users_with_org_and_space_roles('%s', '%s', %d)", org_guid, space_guid, users)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createUsersWithOrgAndSpaceRolesStatement, testConfig)

	helpers.AnalyzeDB(ccdb, ctx, testConfig)
})

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

var _ = ReportAfterSuite("Users test suite", func(report types.Report) {
	helpers.GenerateReports(helpers.ConfigureJsonReporter(&testConfig, "users", "users", test_version), report)
})

func TestUsers(t *testing.T) {
	helpers.LoadConfig(&testConfig)
	RegisterFailHandler(Fail)
	RunSpecs(t, "UsersTest Suite")
}
