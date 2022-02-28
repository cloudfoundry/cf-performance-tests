package service_keys

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"

	"github.com/cloudfoundry/cf-test-helpers/workflowhelpers"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
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
	createServiceInstancesStatement := fmt.Sprintf("SELECT FROM create_service_instances(%d, %d, %d)", spaceWithUnlimitedServiceKeysId, servicePlanId, serviceInstancesPerSpace)
	helpers.ExecuteStatement(ccdb, ctx, createServiceInstancesStatement)

	createServiceInstancesStatement = fmt.Sprintf("SELECT FROM create_service_instances(%d, %d, %d)", spaceWithExhaustedServiceKeysId, servicePlanId, serviceInstancesPerSpace)
	helpers.ExecuteStatement(ccdb, ctx, createServiceInstancesStatement)

	// create service keys
	createServiceKeysStatement := fmt.Sprintf("SELECT FROM create_service_keys_for_service_instances(%d, %d)", spaceWithUnlimitedServiceKeysId, serviceKeysPerServiceInstance)
	helpers.ExecuteStatement(ccdb, ctx, createServiceKeysStatement)

	createServiceKeysStatement = fmt.Sprintf("SELECT FROM create_service_keys_for_service_instances(%d, %d)", spaceWithExhaustedServiceKeysId, serviceKeysPerServiceInstance)
	helpers.ExecuteStatement(ccdb, ctx, createServiceKeysStatement)
})

func createService() int {
	serviceGuid := uuid.NewString()
	serviceName := fmt.Sprintf("%s-service-%s", prefix, serviceGuid)
	createServiceStatement := fmt.Sprintf(
		"INSERT INTO services (guid, label, description, bindable) VALUES ('%s', '%s', '', true) RETURNING id",
		serviceGuid, serviceName)
	return helpers.ExecuteInsertStatement(ccdb, ctx, createServiceStatement)
}

func createServicePlan(serviceId int) int {
	servicePlanGuid := uuid.NewString()
	servicePlanName := fmt.Sprintf("%s-service-plan-%s", prefix, servicePlanGuid)
	createServicePlanStatement := fmt.Sprintf(
		"INSERT INTO service_plans (guid, name, description, free, service_id, unique_id) VALUES ('%s', '%s', '', false, %d, 0) RETURNING id",
		servicePlanGuid, servicePlanName, serviceId)
	return helpers.ExecuteInsertStatement(ccdb, ctx, createServicePlanStatement)
}

func createQuotaDefinition(totalServiceKeys int) int {
	quotaDefinitionGuid := uuid.NewString()
	quotaDefinitionName := fmt.Sprintf("%s-quota-definition-%s", prefix, quotaDefinitionGuid)
	totalServices := serviceInstancesPerSpace
	createQuotaDefinitionStatement := fmt.Sprintf(
		"INSERT INTO quota_definitions (guid, name, non_basic_services_allowed, total_services, memory_limit, total_routes, total_service_keys) VALUES ('%s', '%s', false, %d, 0, 0, %d) RETURNING id",
		quotaDefinitionGuid, quotaDefinitionName, totalServices, totalServiceKeys)
	return helpers.ExecuteInsertStatement(ccdb, ctx, createQuotaDefinitionStatement)
}

func createOrg(quotaDefinitionId int) int {
	orgGuid := uuid.NewString()
	orgName := fmt.Sprintf("%s-org-%s", prefix, orgGuid)
	createOrgStatement := fmt.Sprintf(
		"INSERT INTO organizations (guid, name, quota_definition_id) VALUES ('%s', '%s', %d) RETURNING id",
		orgGuid, orgName, quotaDefinitionId)
	return helpers.ExecuteInsertStatement(ccdb, ctx, createOrgStatement)
}

func createSpace(orgId int) (int, string) {
	spaceGuid := uuid.NewString()
	spaceName := fmt.Sprintf("%s-space-%s", prefix, spaceGuid)
	createSpaceStatement := fmt.Sprintf(
		"INSERT INTO spaces (guid, name, organization_id) VALUES ('%s', '%s', %d) RETURNING id",
		spaceGuid, spaceName, orgId)
	return helpers.ExecuteInsertStatement(ccdb, ctx, createSpaceStatement), spaceGuid
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

func TestServiceKeys(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "ServiceKeysTest Suite", []Reporter{helpers.ConfigureJsonReporter(t, &testConfig, "service-keys")})
}
