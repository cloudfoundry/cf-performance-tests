package service_keys

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"

	"github.com/cloudfoundry/cf-test-helpers/v2/workflowhelpers"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cf-performance-tests/helpers"
)

var testConfig = helpers.NewConfig()
var prefix string
var testSetup *workflowhelpers.ReproducibleTestSuiteSetup
var ccdb *sql.DB
var uaadb *sql.DB
var ctx context.Context

var spaceWithUnlimitedServiceKeysGUID string
var spaceWithExhaustedServiceKeysGUID string

const (
	// main test parameters:
	serviceInstancesPerSpace      = 5000 // i.e. 10000 in 2 spaces
	serviceKeysPerServiceInstance = 20   // i.e. 100000 per space/org
)

var _ = BeforeSuite(func() {
	Expect(serviceInstancesPerSpace).To(BeNumerically(">=", testConfig.LargeElementsFilter))

	testSetup = workflowhelpers.NewTestSuiteSetup(&testConfig)
	testSetup.Setup()
	prefix = testConfig.GetNamePrefix()
	ccdb, uaadb, ctx = helpers.OpenDbConnections(testConfig)
	helpers.ImportStoredProcedures(ccdb, ctx, testConfig)

	// create service and service plan
	serviceId := createService()
	servicePlanId := createServicePlan(serviceId)

	// create quota, org and space
	quotaDefinitionWithUnlimitedServiceKeysId := createQuotaDefinition(-1)
	quotaDefinitionWithExhaustedServiceKeysId := createQuotaDefinition(0)

	orgWithUnlimitedServiceKeysId := createOrg(quotaDefinitionWithUnlimitedServiceKeysId)
	orgWithExhaustedServiceKeysId := createOrg(quotaDefinitionWithExhaustedServiceKeysId)

	var spaceWithUnlimitedServiceKeysId, spaceWithExhaustedServiceKeysId int
	spaceWithUnlimitedServiceKeysId, spaceWithUnlimitedServiceKeysGUID = createSpace(orgWithUnlimitedServiceKeysId)
	spaceWithExhaustedServiceKeysId, spaceWithExhaustedServiceKeysGUID = createSpace(orgWithExhaustedServiceKeysId)

	// create service instances
	createServiceInstancesStatement := fmt.Sprintf("create_service_instances(%d, %d, %d)", spaceWithUnlimitedServiceKeysId, servicePlanId, serviceInstancesPerSpace)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createServiceInstancesStatement, testConfig)

	createServiceInstancesStatement = fmt.Sprintf("create_service_instances(%d, %d, %d)", spaceWithExhaustedServiceKeysId, servicePlanId, serviceInstancesPerSpace)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createServiceInstancesStatement, testConfig)

	// create service keys
	createServiceKeysStatement := fmt.Sprintf("create_service_keys_for_service_instances(%d, %d)", spaceWithUnlimitedServiceKeysId, serviceKeysPerServiceInstance)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createServiceKeysStatement, testConfig)

	createServiceKeysStatement = fmt.Sprintf("create_service_keys_for_service_instances(%d, %d)", spaceWithExhaustedServiceKeysId, serviceKeysPerServiceInstance)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createServiceKeysStatement, testConfig)

	helpers.AnalyzeDB(ccdb, ctx, testConfig)
})

func createService() int {
	serviceGuid := uuid.NewString()
	serviceName := fmt.Sprintf("%s-service-%s", prefix, serviceGuid)
	createServiceStatement := fmt.Sprintf(
		"INSERT INTO services (guid, label, description, bindable) VALUES ('%s', '%s', '', true)",
		serviceGuid, serviceName)
	return helpers.ExecuteInsertStatement(ccdb, ctx, createServiceStatement, testConfig)
}

func createServicePlan(serviceId int) int {
	servicePlanGuid := uuid.NewString()
	servicePlanName := fmt.Sprintf("%s-service-plan-%s", prefix, servicePlanGuid)
	createServicePlanStatement := fmt.Sprintf(
		"INSERT INTO service_plans (guid, name, description, free, service_id, unique_id) VALUES ('%s', '%s', '', false, %d, 0)",
		servicePlanGuid, servicePlanName, serviceId)
	return helpers.ExecuteInsertStatement(ccdb, ctx, createServicePlanStatement, testConfig)
}

func createQuotaDefinition(totalServiceKeys int) int {
	quotaDefinitionGuid := uuid.NewString()
	quotaDefinitionName := fmt.Sprintf("%s-quota-definition-%s", prefix, quotaDefinitionGuid)
	totalServices := serviceInstancesPerSpace
	createQuotaDefinitionStatement := fmt.Sprintf(
		"INSERT INTO quota_definitions (guid, name, non_basic_services_allowed, total_services, memory_limit, total_routes, total_service_keys) VALUES ('%s', '%s', false, %d, 0, 0, %d)",
		quotaDefinitionGuid, quotaDefinitionName, totalServices, totalServiceKeys)
	return helpers.ExecuteInsertStatement(ccdb, ctx, createQuotaDefinitionStatement, testConfig)
}

func createOrg(quotaDefinitionId int) int {
	orgGuid := uuid.NewString()
	orgName := fmt.Sprintf("%s-org-%s", prefix, orgGuid)
	createOrgStatement := fmt.Sprintf(
		"INSERT INTO organizations (guid, name, quota_definition_id) VALUES ('%s', '%s', %d)",
		orgGuid, orgName, quotaDefinitionId)
	return helpers.ExecuteInsertStatement(ccdb, ctx, createOrgStatement, testConfig)
}

func createSpace(orgId int) (int, string) {
	spaceGuid := uuid.NewString()
	spaceName := fmt.Sprintf("%s-space-%s", prefix, spaceGuid)
	createSpaceStatement := fmt.Sprintf(
		"INSERT INTO spaces (guid, name, organization_id) VALUES ('%s', '%s', %d)",
		spaceGuid, spaceName, orgId)
	return helpers.ExecuteInsertStatement(ccdb, ctx, createSpaceStatement, testConfig), spaceGuid
}

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

var _ = ReportAfterSuite("Service keys test suite", func(report types.Report) {
	helpers.GenerateReports(helpers.ConfigureJsonReporter(&testConfig, "service-keys"), report)
})

func TestServiceKeys(t *testing.T) {
	helpers.LoadConfig(&testConfig)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Service keys Test Suite")
}
