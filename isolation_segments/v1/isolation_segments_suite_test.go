package isolation_segments

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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
	orgs              = 20000
	isolationSegments = 500

	orgsWithinIsolationSegments = orgs / 2 // 50%
)

var _ = BeforeSuite(func() {
	Expect(isolationSegments).To(BeNumerically(">=", testConfig.LargePageSize))

	testSetup = workflowhelpers.NewTestSuiteSetup(&testConfig)
	testSetup.Setup()
	ccdb, uaadb, ctx = helpers.OpenDbConnections(testConfig)
	helpers.ImportStoredProcedures(ccdb, ctx)

	// create orgs
	createOrgsStatement := fmt.Sprintf("SELECT FROM create_orgs(%v)", orgs)
	helpers.ExecuteStatement(ccdb, ctx, createOrgsStatement)

	// create isolation segments
	createIsolationSegmentsStatement := fmt.Sprintf("SELECT FROM create_isolation_segments(%v)", isolationSegments)
	helpers.ExecuteStatement(ccdb, ctx, createIsolationSegmentsStatement)

	// assign orgs to isolation segments; n orgs are assigned to a random isolation segment
	assignOrgsToIsolationSegmentsStatement := fmt.Sprintf("SELECT FROM assign_orgs_to_isolation_segments(%d)", orgsWithinIsolationSegments)
	helpers.ExecuteStatement(ccdb, ctx, assignOrgsToIsolationSegmentsStatement)

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

func TestIsolationSegments(t *testing.T) {
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

	timestamp := time.Now().Unix()
	jsonReporter := helpers.NewJsonReporter(fmt.Sprintf("../../test-results/isolation-segments-test-results/v1/isolation-segments-test-results-%d.json", timestamp), testConfig.CfDeploymentVersion, testConfig.CapiVersion, timestamp)

	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "IsolationSegmentsTest Suite", []Reporter{jsonReporter})
}
