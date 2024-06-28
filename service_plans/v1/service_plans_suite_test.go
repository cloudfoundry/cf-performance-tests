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
var orgsWithAccessIDs []string
var orgsFilter string

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
	serviceBrokerId := createServiceBroker(testConfig.GetNamePrefix())

	log.Printf("Stored procedures imported")

	createOrgStatement := fmt.Sprintf("create_orgs(%d)", orgs)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createOrgStatement, testConfig)

	orgsAssignedToRegularUser := orgs / 2

	selectOrgsRandomlyStatement := fmt.Sprintf("create_selected_orgs_table(%d)", orgsAssignedToRegularUser)
	helpers.ExecuteStoredProcedure(ccdb, ctx, selectOrgsRandomlyStatement, testConfig)

	log.Printf("Creating public service plans...")
	createPublicServicePlansStatement := fmt.Sprintf("create_services_and_plans(%v, %v, %v, %v, %v)",
		serviceOfferings, serviceBrokerId, servicePlansPublic, true, 0)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createPublicServicePlansStatement, testConfig)

	log.Printf("Creating private service plans without visibilities...")
	createPrivateServicePlansStatement := fmt.Sprintf("create_services_and_plans(%v, %v, %v, %v, %v)",
		serviceOfferings, serviceBrokerId, servicePlansPrivateWithoutOrgs, false, 0)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createPrivateServicePlansStatement, testConfig)

	log.Printf("Creating private plans with visibilities...")
	createPrivateServicePlansWithOrgsStatement := fmt.Sprintf("create_services_and_plans(%v, %v, %v, %v, %v)",
		serviceOfferings, serviceBrokerId, servicePlansPrivateWithOrgs, false, orgsPerLimitedServicePlan)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createPrivateServicePlansWithOrgsStatement, testConfig)

	// create service instances incl dependent resources
	spacesPerOrg := 1
	createSpacesStatement := fmt.Sprintf("create_spaces(%d)", spacesPerOrg)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createSpacesStatement, testConfig)

	//choose one single space and one single service plan randomly
	selectRandomSpaceStatement := fmt.Sprintf("SELECT spaces.id FROM spaces JOIN selected_orgs ON spaces.organization_id = selected_orgs.id ORDER BY %s LIMIT 1", helpers.GetRandomFunction(testConfig))
	spaceId := helpers.ExecuteSelectStatementOneRow(ccdb, ctx, selectRandomSpaceStatement)

	selectRandomServicePlanStatement := fmt.Sprintf("SELECT s_p_v.service_plan_id FROM service_plan_visibilities AS s_p_v JOIN selected_orgs AS s_o ON s_p_v.organization_id = s_o.id ORDER BY %s LIMIT 1", helpers.GetRandomFunction(testConfig))
	servicePlanId := helpers.ExecuteSelectStatementOneRow(ccdb, ctx, selectRandomServicePlanStatement)

	createServiceInstancesStatement := fmt.Sprintf("create_service_instances(%d, %d, %d)", spaceId, servicePlanId, serviceInstances)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createServiceInstancesStatement, testConfig)

	//assign org_manager to the user for half the number of created orgs randomly
	//assign space_developer rights to the user for all spaces within the orgs where the user received permissions
	regularUserGUID := helpers.GetUserGUID(testSetup.RegularUserContext(), testConfig)

	assignUserAsOrgManager := fmt.Sprintf("assign_user_as_org_role('%s', '%s', %d)", regularUserGUID, "organizations_managers", orgsAssignedToRegularUser)
	helpers.ExecuteStoredProcedure(ccdb, ctx, assignUserAsOrgManager, testConfig)
	spacesAssignedToRegularUser := orgs * spacesPerOrg / 2
	assignUserAsSpaceDeveloper := fmt.Sprintf("assign_user_as_space_role('%s', '%s', %d)", regularUserGUID, "spaces_developers", spacesAssignedToRegularUser)
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
	helpers.GenerateReports(helpers.ConfigureJsonReporter(&testConfig, "service-plans", "service plans"), report)
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
