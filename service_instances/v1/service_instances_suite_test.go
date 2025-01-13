package service_instances

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

const test_version = "v1"

const (
	orgs                          = 10
	spacesPerOrg                  = 20
	serviceInstancesPerSpace      = 200 // 10 x 20 x 200 = 40000 service instances
	serviceInstanceSharesPerSpace = 20  // 10 x 20 x 20 = 4000 service instance shares

	serviceOfferings        = 2
	servicePlansPerOffering = 5 // 2 x 5 = 10 service plans
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

	log.Printf("Creating service offerings and plans...")
	createPublicServicePlansStatement := fmt.Sprintf("create_services_and_plans(%v, %v, %v, %v, %v)",
		serviceOfferings, serviceBrokerId, servicePlansPerOffering, true, 0)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createPublicServicePlansStatement, testConfig)

	// create spaces
	createSpacesStatement := fmt.Sprintf("create_spaces(%d)", spacesPerOrg)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createSpacesStatement, testConfig)

	servicePlans := serviceOfferings * servicePlansPerOffering
	instancesPerPlanPerSpace := serviceInstancesPerSpace / servicePlans

	// create instances
	createServiceInstancesStatement := fmt.Sprintf("create_service_instances_for_orgs_spaces_plans(%d, %d, %d, %d, '%s')", orgs, spacesPerOrg, servicePlans, instancesPerPlanPerSpace, testConfig.GetNamePrefix())
	helpers.ExecuteStoredProcedure(ccdb, ctx, createServiceInstancesStatement, testConfig)

	// create service instance shares
	createServiceInstanceShareStatement := fmt.Sprintf("create_service_instance_shares(%d, %d, %d, '%s')", orgs, spacesPerOrg, instancesPerPlanPerSpace, testConfig.GetNamePrefix())
	helpers.ExecuteStoredProcedure(ccdb, ctx, createServiceInstanceShareStatement, testConfig)

	// assign org_manager to the user for half the number of created orgs randomly
	// assign space_developer rights to the user for all spaces within the orgs where the user received permissions
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

var _ = ReportAfterSuite("Service instances test suite", func(report types.Report) {
	helpers.GenerateReports(helpers.ConfigureJsonReporter(&testConfig, "service-instances", "service instances", test_version), report)
})

func TestServiceInstances(t *testing.T) {
	helpers.LoadConfig(&testConfig)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Service instances Test Suite")
}

func createServiceBroker(prefix string) int {
	serviceBrokerGuid := uuid.NewString()
	serviceBrokerName := fmt.Sprintf("%s-service-broker-%s", prefix, serviceBrokerGuid)
	createServiceBrokerStatement := fmt.Sprintf(
		"INSERT INTO service_brokers (guid, name, broker_url, auth_password) VALUES ('%s', '%s', '', '')",
		serviceBrokerGuid, serviceBrokerName)
	return helpers.ExecuteInsertStatement(ccdb, ctx, createServiceBrokerStatement, testConfig)
}
