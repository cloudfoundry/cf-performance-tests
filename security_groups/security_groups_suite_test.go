package security_groups

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
	helpers.ImportStoredProcedures(ccdb, ctx)

	// create orgs and spaces; as the number of orgs is not relevant for these tests, all spaces are created in a single org
	orgs := 1
	spacesPerOrg := spaces / orgs
	createOrgsStatement := fmt.Sprintf("SELECT FROM create_orgs(%d)", orgs)
	createSpacesStatement := fmt.Sprintf("SELECT FROM create_spaces(%d)", spacesPerOrg)
	helpers.ExecuteStatement(ccdb, ctx, createOrgsStatement)
	helpers.ExecuteStatement(ccdb, ctx, createSpacesStatement)

	// create security groups
	createSecurityGroupsStatement := fmt.Sprintf("SELECT FROM create_security_groups(%d)", securityGroups)
	helpers.ExecuteStatement(ccdb, ctx, createSecurityGroupsStatement)

	// assign security groups to spaces; n spaces have each m security groups (randomly) assigned (a security group can be assigned to multiple spaces)
	assignSecurityGroupsToSpacesStatement := fmt.Sprintf("SELECT FROM assign_security_groups_to_spaces(%d, %d)", spacesWithSecurityGroups, securityGroupsPerSpace)
	helpers.ExecuteStatement(ccdb, ctx, assignSecurityGroupsToSpacesStatement)

	// assign the regular user to all spaces
	regularUserGUID := helpers.GetUserGUID(testSetup.RegularUserContext(), testConfig)
	spacesAssignedToRegularUser := spaces
	assignUserAsSpaceDeveloper := fmt.Sprintf("SELECT FROM assign_user_as_space_developer('%s', %d)", regularUserGUID, spacesAssignedToRegularUser)
	helpers.ExecuteStatement(ccdb, ctx, assignUserAsSpaceDeveloper)
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
