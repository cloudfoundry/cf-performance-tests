package service_plans

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"
	"strconv"
	"strings"

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

	selectOrgsRandomlyStatement := fmt.Sprintf("create_selected_orgs_table(%d)", orgs / 2)
	helpers.ExecuteStoredProcedure(ccdb, ctx, selectOrgsRandomlyStatement, testConfig)

	// instead of fetching random org ids from the organizations table, create a table with the stored procedure containing org ids
	// then get all the org ids from that table and use it in below functions (for postgres)
	// for mysql the procedures can read directly from the DB
	// unclear if this works for all tests
	log.Printf("Getting orgs...")
	selectOrgsStatement := fmt.Sprintf("SELECT id FROM selected_orgs ORDER BY %s LIMIT %d", helpers.GetRandomFunction(testConfig), orgs / 2)
	ids := helpers.ExecuteSelectStatement(ccdb, ctx, selectOrgsStatement)

	orgsWithAccessIDs = make([]string, len(ids))
	for i, v := range ids {
		if id, ok := v.(int64); ok {
			orgsWithAccessIDs[i] = strconv.FormatInt(id, 10)
		} else if id, ok := v.([]uint8); ok {
			orgsWithAccessIDs[i] = string(id)
		}
	}

	// mysql does not support passing arrays into procedures
	if testConfig.DatabaseType == helpers.PsqlDb {
		orgsFilter = fmt.Sprintf("ARRAY[%s]::integer[]", strings.Join(orgsWithAccessIDs, ", "))
	} else {
		orgsFilter = "NULL"
	}

	log.Printf("Creating public service plans...")
	createPublicServicePlansStatement := fmt.Sprintf("create_services_and_plans(%v, %v, %v, %v, %v, NULL)",
		serviceOfferings, serviceBrokerId, servicePlansPublic, true, 0)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createPublicServicePlansStatement, testConfig)

	log.Printf("Creating private service plans without visibilities...")
	createPrivateServicePlansStatement := fmt.Sprintf("create_services_and_plans(%v, %v, %v, %v, %v, NULL)",
		serviceOfferings, serviceBrokerId, servicePlansPrivateWithoutOrgs, false, 0)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createPrivateServicePlansStatement, testConfig)

	log.Printf("Creating private plans with visibilities...")
	createPrivateServicePlansWithOrgsStatement := fmt.Sprintf("create_services_and_plans(%v, %v, %v, %v, %v, %s)",
		serviceOfferings, serviceBrokerId, servicePlansPrivateWithOrgs, false, orgsPerLimitedServicePlan, orgsFilter)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createPrivateServicePlansWithOrgsStatement, testConfig)

	// create service instances incl dependent resources
	spacesPerOrg := 1
	createSpacesStatement := fmt.Sprintf("create_spaces(%d)", spacesPerOrg)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createSpacesStatement, testConfig)

	//choose one single space and one single service plan randomly
	//then create 500 service instances of that service plan in that space
	//PROBLEM: the selected plan might be of kind "private without orgs" -> user will still see the service instances but cannot see the plans relevant for test: GET /v3/service_plans?service_instance_guids=
	//selectRandomSpaceStatement := fmt.Sprintf("SELECT id FROM spaces WHERE name LIKE '%s-space-%%' ORDER BY %s LIMIT 1", testConfig.GetNamePrefix(), helpers.GetRandomFunction(testConfig))
	selectRandomSpaceStatement := fmt.Sprintf("SELECT spaces.id FROM spaces JOIN selected_orgs ON spaces.organization_id = selected_orgs.id ORDER BY %s LIMIT 1", helpers.GetRandomFunction(testConfig))

	spaceId := helpers.ExecuteSelectStatementOneRow(ccdb, ctx, selectRandomSpaceStatement)
	selectRandomServicePlanStatement := fmt.Sprintf("SELECT s_p_v.service_plan_id FROM service_plan_visibilities AS s_p_v JOIN selected_orgs AS s_o ON s_p_v.organization_id = s_o.id ORDER BY %s LIMIT 1", helpers.GetRandomFunction(testConfig))
	servicePlanId := helpers.ExecuteSelectStatementOneRow(ccdb, ctx, selectRandomServicePlanStatement)
	createServiceInstancesStatement := fmt.Sprintf("create_service_instances(%d, %d, %d)", spaceId, servicePlanId, serviceInstances)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createServiceInstancesStatement, testConfig)

	//assign org_manager to the user for half the number of created orgs randomly
	//assign space_developer rights to the user for half the number of created spaces randomly
	//-> actually the user can see service instances if he has a space role, but no org role. But this case cannot exist, because the API won't let you create a space role if no org role exists
	//BUT: it might be that the user won't get any role in the space and org where the service instances have been created, in which case the user won't see any service instances. Relevant for tests case: GET /v3/service_plans?service_instance_guids=
	//also relevant for: /v3/service_plans?organization_guids=:guid&space_guids (number of orgs and spaces with assigned roles can vary)
	//also relevant for: /v3/service_plans?service_offering_guids= (randomly selected service offerings in the test might not be visible to the user

	//TODO: make sure that the plan used for the service instances is orgsPerLimitedServicePlan
	//		make sure that the service instances get created in an org where the plan has been enabled and where the user has access to (either org or space role)
	// this should fix the two tests with service instances filter

	//TODO: make sure that the number of visible service plans is always the same for a regular user
	//probably select orgs randomly assign org role and then assign space role in that org's space?
	regularUserGUID := helpers.GetUserGUID(testSetup.RegularUserContext(), testConfig)
	orgsAssignedToRegularUser := orgs / 2
	assignUserAsOrgManager := fmt.Sprintf("assign_user_as_org_role('%s', '%s', %d, %s)", regularUserGUID, "organizations_managers", orgsAssignedToRegularUser, orgsFilter)
	helpers.ExecuteStoredProcedure(ccdb, ctx, assignUserAsOrgManager, testConfig)
	spacesAssignedToRegularUser := orgs * spacesPerOrg / 2
	assignUserAsSpaceDeveloper := fmt.Sprintf("assign_user_as_space_role('%s', '%s', %d, %s)", regularUserGUID, "spaces_developers", spacesAssignedToRegularUser, orgsFilter)
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
