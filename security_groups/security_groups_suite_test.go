package security_groups

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
	orgs = 10
	spaces = 10
	securityGroups = 10
)

var _ = BeforeSuite(func() {
	testSetup = workflowhelpers.NewTestSuiteSetup(&testConfig)
	testSetup.Setup()
	ccdb, uaadb, ctx = helpers.OpenDbConnections(testConfig.CcdbConnection, testConfig.UaadbConnection)

	createOrgStatement := fmt.Sprintf(`SELECT FROM create_test_orgs(%v)`, orgs)
	createSpaceStatement := fmt.Sprintf(`SELECT FROM create_test_spaces(%v)`, spaces)
	createSecurityGroupStatement := fmt.Sprintf(`SELECT FROM create_test_security_groups(%v)`, securityGroups)
	createSecurityGroupSpaceStatement := fmt.Sprintf(`SELECT FROM create_test_security_group_spaces()`)
	helpers.ExecuteStatement(ccdb, ctx, createOrgStatement)
	helpers.ExecuteStatement(ccdb, ctx, createSpaceStatement)
	helpers.ExecuteStatement(ccdb, ctx, createSecurityGroupStatement)
	helpers.ExecuteStatement(ccdb, ctx, createSecurityGroupSpaceStatement)

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

func TestSecurityGroups(t *testing.T) {
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
	jsonReporter := helpers.NewJsonReporter(fmt.Sprintf("../test-results/security-groups-test-results-%d.json", timestamp), testConfig.CfDeploymentVersion, testConfig.CapiVersion, timestamp)

	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "SecurityGroupsTest Suite", []Reporter{jsonReporter})
}

