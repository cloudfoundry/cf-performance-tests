package organizations

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"

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

const test_version = "v1"

const (
	// main test parameters:
	orgs = 100000
)

var _ = BeforeSuite(func() {
	testSetup = workflowhelpers.NewTestSuiteSetup(&testConfig)
	testSetup.Setup()
	ccdb, uaadb, ctx = helpers.OpenDbConnections(testConfig)
	helpers.ImportStoredProcedures(ccdb, ctx, testConfig)

	createOrgStatement := fmt.Sprintf("create_orgs(%d)", orgs)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createOrgStatement, testConfig)

	orgsAssignedToRegularUser := orgs / 10
	selectOrgsStatement := fmt.Sprintf("create_selected_orgs_table(%d)", orgsAssignedToRegularUser)
	helpers.ExecuteStoredProcedure(ccdb, ctx, selectOrgsStatement, testConfig)

	regularUserGUID := helpers.GetUserGUID(testSetup.RegularUserContext(), testConfig)
	assignDisjointOrgRoles := fmt.Sprintf("assign_user_disjoint_org_roles('%s')", regularUserGUID)
	helpers.ExecuteStoredProcedure(ccdb, ctx, assignDisjointOrgRoles, testConfig)

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

var _ = ReportAfterSuite("Organizations test suite", func(report types.Report) {
	helpers.GenerateReports(helpers.ConfigureJsonReporter(&testConfig, "organizations", "organizations", test_version), report)
})

func TestOrganizations(t *testing.T) {
	helpers.LoadConfig(&testConfig)
	RegisterFailHandler(Fail)
	RunSpecs(t, "OrganizationsTest Suite")
}
