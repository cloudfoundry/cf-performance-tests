package service_plans

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/google/uuid"

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
	orgs                           = 10000
	serviceOfferings               = 300
	servicePlansPublic             = 10  // results in 300 services with 10 service plans each (3k total)
	servicePlansPrivateWithoutOrgs = 10  // results in 300 services with 10 service plans each (3k total)
	servicePlansPrivateWithOrgs    = 10  // results in 300 services with 10 service plans each (3k total)
	orgsPerLimitedServicePlan      = 200 // used in `servicePlansPrivateWithOrgs`, results in 600k (3k * 200) service_plan_visibilities
	serviceInstances               = 500
)

var _ = BeforeSuite(func() {

	testSetup = workflowhelpers.NewTestSuiteSetup(&testConfig)
	testSetup.Setup()
	ccdb, uaadb, ctx = helpers.OpenDbConnections(testConfig)

	fmt.Printf("%v Starting to seed database with testdata...\n", time.Now().Format(time.RFC850))

	helpers.ImportStoredProcedures(ccdb, ctx, testConfig)
	if testConfig.DatabaseType == helpers.MysqlDb {
		helpers.DefineRandomFunction(ccdb, ctx)
	}

	serviceBrokerId := createServiceBroker(testConfig.GetNamePrefix())

	createOrgStatement := fmt.Sprintf("create_orgs(%d)", orgs)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createOrgStatement, testConfig)

	createPublicServicePlansStatement := fmt.Sprintf("create_services_and_plans(%v, %v, %v, %v, %v)",
		serviceOfferings, serviceBrokerId, servicePlansPublic, true, 0)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createPublicServicePlansStatement, testConfig)

	createPrivateServicePlansStatement := fmt.Sprintf("create_services_and_plans(%v, %v, %v, %v, %v)",
		serviceOfferings, serviceBrokerId, servicePlansPrivateWithoutOrgs, false, 0)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createPrivateServicePlansStatement, testConfig)

	createPrivateServicePlansWithOrgsStatement := fmt.Sprintf("create_services_and_plans(%v, %v, %v, %v, %v)",
		serviceOfferings, serviceBrokerId, servicePlansPrivateWithOrgs, false, orgsPerLimitedServicePlan)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createPrivateServicePlansWithOrgsStatement, testConfig)

	// create service instances incl dependent resources
	spacesPerOrg := 1
	createSpacesStatement := fmt.Sprintf("create_spaces(%d)", spacesPerOrg)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createSpacesStatement, testConfig)
	selectRandomSpaceStatement := fmt.Sprintf("SELECT id FROM spaces WHERE name LIKE '%s-space-%%' ORDER BY random() LIMIT 1", testConfig.GetNamePrefix())
	spaceId := helpers.ExecuteSelectStatementOneRow(ccdb, ctx, selectRandomSpaceStatement)
	servicePlanId := helpers.ExecuteSelectStatementOneRow(ccdb, ctx, "SELECT id FROM service_plans ORDER BY random() LIMIT 1")
	createServiceInstancesStatement := fmt.Sprintf("create_service_instances(%d, %d, %d)", spaceId, servicePlanId, serviceInstances)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createServiceInstancesStatement, testConfig)

	regularUserGUID := helpers.GetUserGUID(testSetup.RegularUserContext(), testConfig)
	orgsAssignedToRegularUser := orgs / 2
	assignUserAsOrgManager := fmt.Sprintf("assign_user_as_org_manager('%s', %d)", regularUserGUID, orgsAssignedToRegularUser)
	helpers.ExecuteStoredProcedure(ccdb, ctx, assignUserAsOrgManager, testConfig)
	spacesAssignedToRegularUser := orgs * spacesPerOrg / 2
	assignUserAsSpaceDeveloper := fmt.Sprintf("assign_user_as_space_developer('%s', %d)", regularUserGUID, spacesAssignedToRegularUser)
	helpers.ExecuteStoredProcedure(ccdb, ctx, assignUserAsSpaceDeveloper, testConfig)

	helpers.AnalyzeDB(ccdb, ctx, testConfig)
	fmt.Printf("%v Finished seeding database.\n", time.Now().Format(time.RFC850))
})

var _ = AfterSuite(func() {
	fmt.Printf("%v Starting cleanup testdata...\n", time.Now().Format(time.RFC850))
	helpers.CleanupTestData(ccdb, uaadb, ctx, testConfig)
	fmt.Printf("%v Finished cleanup testdata...\n", time.Now().Format(time.RFC850))
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

var _ = ReportAfterSuite("Service plans test suite", func(report types.Report) {
	helpers.GenerateReports(helpers.ConfigureJsonReporter(&testConfig, "Service plans"), report)
})

func TestServicePlans(t *testing.T) {
	helpers.LoadConfig(&testConfig)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Service plans Test Suite")
}

func createServiceBroker(prefix string) int {
	serviceBrokerGuid := uuid.NewString()
	serviceBrokerName := fmt.Sprintf("%s-service-broker-%s", prefix, serviceBrokerGuid)
	createServiceBrokerStatement := fmt.Sprintf(
		"INSERT INTO service_brokers (guid, name, broker_url, auth_password) VALUES ('%s', '%s', '', '')",
		serviceBrokerGuid, serviceBrokerName)
	return helpers.ExecuteInsertStatement(ccdb, ctx, createServiceBrokerStatement, testConfig)
}
