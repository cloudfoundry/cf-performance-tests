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

	createOrgStatement := fmt.Sprintf("create_orgs(%d)", orgs)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createOrgStatement, testConfig)

	log.Printf("Getting orgs...")
	selectOrgsRandomlyStatement := fmt.Sprintf("SELECT id FROM organizations WHERE name LIKE '%s-org-%%' ORDER BY %s LIMIT %d", testConfig.GetNamePrefix(), helpers.GetRandomFunction(testConfig), orgs / 2)
	ids := helpers.ExecuteSelectStatement(ccdb, ctx, selectOrgsRandomlyStatement)

	orgsWithAccessIDs = make([]string, len(ids))
	for i, v := range ids {
		if id, ok := v.(int64); ok {
			orgsWithAccessIDs[i] = strconv.FormatInt(id, 10)
		}
	}

	log.Printf("Creating public service plans...")
	createPublicServicePlansStatement := fmt.Sprintf("create_services_and_plans(%v, %v, %v, %v, %v, ARRAY[]::integer[])",
		serviceOfferings, serviceBrokerId, servicePlansPublic, true, 0)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createPublicServicePlansStatement, testConfig)

	log.Printf("Creating private service plans without visibilities...")
	createPrivateServicePlansStatement := fmt.Sprintf("create_services_and_plans(%v, %v, %v, %v, %v, ARRAY[]::integer[])",
		serviceOfferings, serviceBrokerId, servicePlansPrivateWithoutOrgs, false, 0)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createPrivateServicePlansStatement, testConfig)

	log.Printf("Creating private plans with visibilities...")
	createPrivateServicePlansWithOrgsStatement := fmt.Sprintf("create_services_and_plans(%v, %v, %v, %v, %v, ARRAY[%s]::integer[])",
		serviceOfferings, serviceBrokerId, servicePlansPrivateWithOrgs, false, orgsPerLimitedServicePlan, strings.Join(orgsWithAccessIDs, ", "))
	helpers.ExecuteStoredProcedure(ccdb, ctx, createPrivateServicePlansWithOrgsStatement, testConfig)

	// create service instances incl dependent resources
	spacesPerOrg := 1
	createSpacesStatement := fmt.Sprintf("create_spaces(%d)", spacesPerOrg)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createSpacesStatement, testConfig)

	//choose one single space and one single service plan randomly
	//then create 500 service instances of that service plan in that space
	//PROBLEM: the selected plan might be of kind "private without orgs" -> user will still see the service instances but cannot see the plans relevant for test: GET /v3/service_plans?service_instance_guids=
	//selectRandomSpaceStatement := fmt.Sprintf("SELECT id FROM spaces WHERE name LIKE '%s-space-%%' ORDER BY %s LIMIT 1", testConfig.GetNamePrefix(), helpers.GetRandomFunction(testConfig))
	selectRandomSpaceStatement := fmt.Sprintf("SELECT id FROM spaces WHERE name LIKE '%s-space-%%' AND organization_id = ANY(ARRAY[%s]::integer[]) ORDER BY %s LIMIT 1", testConfig.GetNamePrefix(), strings.Join(orgsWithAccessIDs, ", "), helpers.GetRandomFunction(testConfig))

	spaceId := helpers.ExecuteSelectStatementOneRow(ccdb, ctx, selectRandomSpaceStatement)
	//selectRandomServicePlanStatement := fmt.Sprintf("SELECT id FROM service_plans ORDER BY %s LIMIT 1", helpers.GetRandomFunction(testConfig))
	selectRandomServicePlanStatement := fmt.Sprintf("SELECT service_plan_id FROM service_plan_visibilities WHERE organization_id = ANY(ARRAY[%s]::integer[]) ORDER BY %s LIMIT 1", strings.Join(orgsWithAccessIDs, ", "), helpers.GetRandomFunction(testConfig))
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
	assignUserAsOrgManager := fmt.Sprintf("assign_user_as_org_role('%s', '%s', %d, ARRAY[%s]::integer[])", regularUserGUID, "organizations_managers", orgsAssignedToRegularUser, strings.Join(orgsWithAccessIDs, ", "))
	helpers.ExecuteStoredProcedure(ccdb, ctx, assignUserAsOrgManager, testConfig)
	spacesAssignedToRegularUser := orgs * spacesPerOrg / 2
	assignUserAsSpaceDeveloper := fmt.Sprintf("assign_user_as_space_role('%s', '%s', %d, ARRAY[%s]::integer[])", regularUserGUID, "spaces_developers", spacesAssignedToRegularUser, strings.Join(orgsWithAccessIDs, ", "))
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
