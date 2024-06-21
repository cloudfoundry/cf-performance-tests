package security_groups

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
	spaces         = 500
	securityGroups = 500

	spacesWithSecurityGroups = spaces / 2         // 50%
	securityGroupsPerSpace   = securityGroups / 2 // 50%
)

var _ = BeforeSuite(func() {
	Expect(spaces).To(BeNumerically(">=", testConfig.LargeElementsFilter))
	Expect(securityGroups).To(BeNumerically(">=", testConfig.LargePageSize))

	testSetup = workflowhelpers.NewTestSuiteSetup(&testConfig)
	testSetup.Setup()
	ccdb, uaadb, ctx = helpers.OpenDbConnections(testConfig)
	helpers.ImportStoredProcedures(ccdb, ctx, testConfig)

	// create orgs and spaces; as the number of orgs is not relevant for these tests, all spaces are created in a single org
	orgs := 1
	spacesPerOrg := spaces / orgs
	createOrgsStatement := fmt.Sprintf("create_orgs(%d)", orgs)
	createSpacesStatement := fmt.Sprintf("create_spaces(%d)", spacesPerOrg)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createOrgsStatement, testConfig)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createSpacesStatement, testConfig)

	// create security groups
	createSecurityGroupsStatement := fmt.Sprintf("create_security_groups(%d)", securityGroups)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createSecurityGroupsStatement, testConfig)

	// assign security groups to spaces; n spaces have each m security groups (randomly) assigned (a security group can be assigned to multiple spaces)
	assignSecurityGroupsToSpacesStatement := fmt.Sprintf("assign_security_groups_to_spaces(%d, %d)", spacesWithSecurityGroups, securityGroupsPerSpace)
	helpers.ExecuteStoredProcedure(ccdb, ctx, assignSecurityGroupsToSpacesStatement, testConfig)

	// assign the regular user to all spaces
	regularUserGUID := helpers.GetUserGUID(testSetup.RegularUserContext(), testConfig)
	spacesAssignedToRegularUser := spaces
	assignUserAsSpaceDeveloper := fmt.Sprintf("assign_user_as_space_role('%s', '%s', %d, NULL)", regularUserGUID, "spaces_developers", spacesAssignedToRegularUser)
	helpers.ExecuteStoredProcedure(ccdb, ctx, assignUserAsSpaceDeveloper, testConfig)

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

var _ = ReportAfterSuite("Security groups test suite", func(report types.Report) {
	helpers.GenerateReports(helpers.ConfigureJsonReporter(&testConfig, "security-groups", "security groups"), report)
})

func TestSecurityGroups(t *testing.T) {
	helpers.LoadConfig(&testConfig)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Security groups Suite")
}
