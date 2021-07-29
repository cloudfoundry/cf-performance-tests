package security_groups

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/cloudfoundry-incubator/cf-performance-tests/helpers"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	"github.com/google/uuid"
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

	quotaId := helpers.ExecuteSelectStatementOneRow(ccdb, ctx, "SELECT id FROM quota_definitions WHERE name = 'default'")
	var organizationIds []int
	var spaceIds []int
	var securityGroupIds []int

	for i := 0; i < orgs; i++ {
		guid := uuid.New()
		name := testConfig.NamePrefix + "-org-" + guid.String()
		statement := "INSERT INTO organizations (guid, name, quota_definition_id) VALUES ($1, $2, $3) RETURNING id"
		organizationId := helpers.ExecutePreparedInsertStatement(ccdb, ctx, statement, guid.String(), name, quotaId)
		organizationIds = append(organizationIds, organizationId)
	}
	// TODO create orgs return orgs id; create spaces return space id ; 1 org = 1 space;

	for _, orgId := range organizationIds{
		for i := 0; i<spaces; i++ {
			spaceGuid := uuid.New()
			spaceName := testConfig.NamePrefix + "-space-" + spaceGuid.String()
			statement := "INSERT INTO spaces (guid, name, organization_id) VALUES ($1, $2, $3) RETURNING id"
			spaceId := helpers.ExecutePreparedInsertStatement(ccdb, ctx, statement, spaceGuid.String(), spaceName, orgId)
			spaceIds = append(spaceIds, spaceId)
		}
	}

	for i := 0; i<securityGroups; i++ {
		securityGroupsGuid := uuid.New()
		securityGroupName := testConfig.NamePrefix + "-securitygroup-" + securityGroupsGuid.String()
		securityRule := `[
  {
	"protocol": "icmp",
	"destination": "0.0.0.0/0",
	"type": 0,
	"code": 0
  },
  {
	"protocol": "tcp",
	"destination": "10.0.11.0/24",
	"ports": "80,443",
	"log": true,
	"description": "Allow http and https traffic to ZoneA"
  }
]`
		statement := "INSERT INTO security_groups (guid, name, rules) VALUES ($1, $2, $3) RETURNING id"
		securityGroupId := helpers.ExecutePreparedInsertStatement(ccdb, ctx, statement, securityGroupsGuid.String(), securityGroupName, securityRule)
		securityGroupIds = append(securityGroupIds, securityGroupId)
	}

	for _, spaceId := range spaceIds{
		for i := 0; i<5; i++ {
			securityGroupId := securityGroupIds[rand.Intn(len(securityGroupIds))]
			statement := "INSERT INTO security_groups_spaces (security_group_id, space_id) VALUES ($1, $2) RETURNING security_groups_spaces_pk"
			helpers.ExecutePreparedInsertStatement(ccdb, ctx, statement, securityGroupId, spaceId)
		}
	}

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
	jsonReporter := helpers.NewJsonReporter(fmt.Sprintf("../test-results/security-groups-test-results-%d.json", timestamp), testConfig.CfDeploymentVersion, timestamp)

	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "SecurityGroupsTest Suite", []Reporter{jsonReporter})
}

