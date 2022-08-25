package isolation_segments

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

const (
	// main test parameters:
	orgs              = 20000
	isolationSegments = 500

	orgsWithinIsolationSegments = orgs / 2 // 50%
)

var _ = BeforeSuite(func() {
	Expect(isolationSegments).To(BeNumerically(">=", testConfig.LargePageSize))

	testSetup = workflowhelpers.NewTestSuiteSetup(&testConfig)
	testSetup.Setup()
	ccdb, uaadb, ctx = helpers.OpenDbConnections(testConfig)
	helpers.ImportStoredProcedures(ccdb, ctx, testConfig)

	// create orgs
	createOrgsStatement := fmt.Sprintf("create_orgs(%v)", orgs)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createOrgsStatement, testConfig)

	// create isolation segments
	createIsolationSegmentsStatement := fmt.Sprintf("create_isolation_segments(%v)", isolationSegments)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createIsolationSegmentsStatement, testConfig)

	// assign orgs to isolation segments; n orgs are assigned to a random isolation segment
	assignOrgsToIsolationSegmentsStatement := fmt.Sprintf("assign_orgs_to_isolation_segments(%d)", orgsWithinIsolationSegments)
	helpers.ExecuteStoredProcedure(ccdb, ctx, assignOrgsToIsolationSegmentsStatement, testConfig)

	// assign the regular user to all orgs
	regularUserGUID := helpers.GetUserGUID(testSetup.RegularUserContext(), testConfig)
	orgsAssignedToRegularUser := orgs
	assignUserAsOrgManager := fmt.Sprintf("assign_user_as_org_manager('%s', %d)", regularUserGUID, orgsAssignedToRegularUser)
	helpers.ExecuteStoredProcedure(ccdb, ctx, assignUserAsOrgManager, testConfig)

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

var _ = ReportAfterSuite("Isolation segments test suite", func(report types.Report) {
	helpers.GenerateReports(helpers.ConfigureJsonReporter(&testConfig, "isolationSegments"), report)
})

func TestIsolationSegments(t *testing.T) {
	helpers.LoadConfig(&testConfig)
	RegisterFailHandler(Fail)
	RunSpecs(t, "IsolationSegments Suite")
}
