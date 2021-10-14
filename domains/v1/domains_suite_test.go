package domains

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"

	"github.com/cloudfoundry-incubator/cf-performance-tests/helpers"
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
	helpers.ImportStoredProcedures(ccdb, ctx)

	// create orgs
	createOrgStatement := fmt.Sprintf("SELECT FROM create_orgs(%d)", orgs)
	helpers.ExecuteStatement(ccdb, ctx, createOrgStatement)

	// create shared domains
	createSharedDomainsStatement := fmt.Sprintf("SELECT FROM create_shared_domains(%d)", sharedDomains)
	helpers.ExecuteStatement(ccdb, ctx, createSharedDomainsStatement)

	// create private domains; evenly assigned to random orgs
	createPrivateDomainsStatement := fmt.Sprintf("SELECT FROM create_private_domains(%d)", privateDomains)
	helpers.ExecuteStatement(ccdb, ctx, createPrivateDomainsStatement)

	// assign the regular user to all orgs
	regularUserGUID := helpers.GetUserGUID(testSetup.RegularUserContext(), testConfig)
	orgsAssignedToRegularUser := orgs
	assignUserAsOrgManager := fmt.Sprintf("SELECT FROM assign_user_as_org_manager('%s', %d)", regularUserGUID, orgsAssignedToRegularUser)
	helpers.ExecuteStatement(ccdb, ctx, assignUserAsOrgManager)
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
	viper.SetConfigName("config")
	viper.AddConfigPath("../../")
	viper.AddConfigPath("$HOME/.cf-performance-tests")
	err := viper.ReadInConfig()
	if err != nil {
		t.Fatalf("error loading config: %s", err.Error())
	}

	err = viper.Unmarshal(&testConfig)
	if err != nil {
		t.Fatalf("error parsing config: %s", err.Error())
	}

	err = os.MkdirAll("../../test-results/domains-test-results/v1/", os.ModePerm)
	if err != nil {
		t.Fatalf("Cannot create Directory: %s", err.Error())
	}

	timestamp := time.Now().Unix()
	jsonReporter := helpers.NewJsonReporter(fmt.Sprintf("../../test-results/domains-test-results/v1/domains-test-results-%d.json", timestamp), testConfig.CfDeploymentVersion, testConfig.CapiVersion, timestamp)

	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "DomainsTest Suite", []Reporter{jsonReporter})
}
