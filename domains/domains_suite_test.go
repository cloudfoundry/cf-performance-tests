package domains

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/cloudfoundry-incubator/cf-performance-tests/helpers"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

var testConfig helpers.Config = helpers.NewConfig()
var testSetup *workflowhelpers.ReproducibleTestSuiteSetup
var ccdb *sql.DB
var uaadb *sql.DB
var ctx context.Context

const (
	orgs           = 10
	sharedDomains  = 10
	privateDomains = 10
)

var _ = BeforeSuite(func() {
	testSetup = workflowhelpers.NewTestSuiteSetup(&testConfig)
	testSetup.Setup()
	ccdb, uaadb, ctx = helpers.OpenDbConnections(testConfig.CcdbConnection, testConfig.UaadbConnection)

	//quotaId := helpers.ExecuteSelectStatementOneRow(ccdb, ctx, "SELECT id FROM quota_definitions WHERE name = 'default'")
	// required create functions and pass it to the database once. Then we can just call the stored db functions here.

	createOrgStatement := fmt.Sprintf(`SELECT FROM create_test_orgs(%v)`, orgs)
	createSharedDomainStatement := fmt.Sprintf(`SELECT FROM create_test_shared_domains(%v)`, sharedDomains)
	createPrivateDomainStatement := fmt.Sprintf(`SELECT FROM create_test_private_domains(%v)`, privateDomains)
	helpers.ExecuteStatement(ccdb, ctx, createOrgStatement)
	helpers.ExecuteStatement(ccdb, ctx, createSharedDomainStatement)
	helpers.ExecuteStatement(ccdb, ctx, createPrivateDomainStatement)

})

var _ = AfterSuite(func() {

	helpers.CleanupTestData(ccdb, uaadb, ctx)

	err := ccdb.Close()
	if err != nil {
		log.Print(err)
	}

	err = uaadb.Close()
	if err != nil {
		log.Print(err)
	}
})

func TestDomains(t *testing.T) {
	viper.SetConfigName("config")
	viper.AddConfigPath("..")
	viper.AddConfigPath("$HOME/.cf-performance-tests")
	err := viper.ReadInConfig()
	if err != nil {
		t.Fatalf("error loading config: %s", err.Error())
	}

	err = viper.Unmarshal(&testConfig)
	if err != nil {
		t.Fatalf("error parsing config: %s", err.Error())
	}

	timestamp := time.Now().Unix()
	jsonReporter := helpers.NewJsonReporter(fmt.Sprintf("../test-results/domains-test-results-%d.json", timestamp), testConfig.CfDeploymentVersion, timestamp)

	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "DomainsTest Suite", []Reporter{jsonReporter})
}
