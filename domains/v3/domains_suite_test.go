package domains

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

const test_version = "v3"

const (
	// main test parameters:
	orgs           = 20000
	sharedDomains  = 100
	privateDomains = 400

	// per-domain fan-out of the associations the list endpoint presents per
	// row; v2 leaves these empty, so a per-row (N+1) resolution stays invisible
	labelsPerDomain      = 10
	annotationsPerDomain = 10
	sharedOrgsPerDomain  = 25

	// the write tests (PATCH, DELETE) target domains that are not shared with
	// other orgs; a shared domain cannot be deleted
	unsharedPrivateDomains = 10
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

	// copy ids of orgs relevant for regular user
	selectOrgsRandomlyStatement := fmt.Sprintf("create_selected_orgs_table(%d)", orgs)
	helpers.ExecuteStoredProcedure(ccdb, ctx, selectOrgsRandomlyStatement, testConfig)

	// create shared domains
	createSharedDomainsStatement := fmt.Sprintf("create_shared_domains(%d)", sharedDomains)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createSharedDomainsStatement, testConfig)

	// create private domains; evenly assigned to random orgs
	createPrivateDomainsStatement := fmt.Sprintf("create_private_domains(%d)", privateDomains)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createPrivateDomainsStatement, testConfig)

	// give every private domain a dense set of labels and annotations
	addMetadataStatement := fmt.Sprintf("add_metadata_to_private_domains(%d, %d)", labelsPerDomain, annotationsPerDomain)
	helpers.ExecuteStoredProcedure(ccdb, ctx, addMetadataStatement, testConfig)

	// share every private domain with a dense set of orgs
	sharePrivateDomainsStatement := fmt.Sprintf("share_private_domains_with_orgs(%d)", sharedOrgsPerDomain)
	helpers.ExecuteStoredProcedure(ccdb, ctx, sharePrivateDomainsStatement, testConfig)

	// create a small pool of unshared private domains for the write tests
	createUnsharedStatement := fmt.Sprintf("create_unshared_private_domains(%d)", unsharedPrivateDomains)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createUnsharedStatement, testConfig)

	// assign the regular user to all orgs
	regularUserGUID := helpers.GetUserGUID(testSetup.RegularUserContext(), testConfig)
	orgsAssignedToRegularUser := orgs
	assignUserAsOrgManager := fmt.Sprintf("assign_user_as_org_role('%s', '%s', %d)", regularUserGUID, "organizations_managers", orgsAssignedToRegularUser)
	helpers.ExecuteStoredProcedure(ccdb, ctx, assignUserAsOrgManager, testConfig)

	helpers.AnalyzeDB(ccdb, ctx, testConfig)
})

var _ = AfterSuite(func() {
	// setup may have failed before the DB connection was opened
	if ccdb == nil {
		return
	}

	// CleanupTestData removes domains and domain_annotations, but not
	// domain_labels or organizations_private_domains; remove them first
	namePattern := fmt.Sprintf("%s-%%", testConfig.GetNamePrefix())
	helpers.ExecuteStatement(ccdb, ctx, fmt.Sprintf(
		"DELETE FROM organizations_private_domains WHERE private_domain_id IN (SELECT id FROM domains WHERE name LIKE '%s')", namePattern))
	helpers.ExecuteStatement(ccdb, ctx, fmt.Sprintf(
		"DELETE FROM domain_labels WHERE resource_guid IN (SELECT guid FROM domains WHERE name LIKE '%s')", namePattern))

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

var _ = ReportAfterSuite("Domains test suite", func(report types.Report) {
	helpers.GenerateReports(helpers.ConfigureJsonReporter(&testConfig, "domains", "domains", test_version), report)
})

func TestDomains(t *testing.T) {
	helpers.LoadConfig(&testConfig)
	RegisterFailHandler(Fail)
	RunSpecs(t, "DomainsTest Suite")
}
