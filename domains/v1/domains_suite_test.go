package domains

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"

	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
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
	orgs           = 20000
	sharedDomains  = 100
	privateDomains = 400
)

var _ = BeforeSuite(func() {
	Expect(sharedDomains + privateDomains).To(BeNumerically(">=", testConfig.LargePageSize))

	testSetup = workflowhelpers.NewTestSuiteSetup(&testConfig)
	testSetup.Setup()
	ccdb, uaadb, ctx = helpers.OpenDbConnections(testConfig)
	helpers.ImportStoredProcedures(ccdb, ctx, testConfig)

	// create orgs
	createOrgStatement := fmt.Sprintf("create_orgs(%d)", orgs)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createOrgStatement, testConfig)

	// create shared domains
	createSharedDomainsStatement := fmt.Sprintf("create_shared_domains(%d)", sharedDomains)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createSharedDomainsStatement, testConfig)

	// create private domains; evenly assigned to random orgs
	createPrivateDomainsStatement := fmt.Sprintf("create_private_domains(%d)", privateDomains)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createPrivateDomainsStatement, testConfig)

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

func TestDomains(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "DomainsTest Suite", []Reporter{helpers.ConfigureJsonReporter(t, &testConfig, "domains")})
}
