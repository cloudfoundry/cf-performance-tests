package service_plans

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"log"
	"testing"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/cf-performance-tests/helpers"
)

var testConfig = helpers.NewConfig()
var testSetup *workflowhelpers.ReproducibleTestSuiteSetup
var ccdb *sql.DB
var uaadb *sql.DB
var ctx context.Context

const (
	orgs = 10000
	serviceOfferings = 300
	servicePlansPublic = 10
	servicePlansPrivateWithoutOrgs = 10
	servicePlansPrivateWithOrgs = 10
	orgsPerLimitedServicePlan = 200
)

var _ = BeforeSuite(func() {

	testSetup = workflowhelpers.NewTestSuiteSetup(&testConfig)
	testSetup.Setup()
	ccdb, uaadb, ctx = helpers.OpenDbConnections(testConfig)

	fmt.Printf("%v Starting to seed database with testdata...\n", time.Now().Format(time.RFC850))

	helpers.ImportStoredProcedures(ccdb, ctx, testConfig)

	serviceBrokerId := createServiceBroker(testConfig.GetNamePrefix())

	createOrgStatement := fmt.Sprintf("SELECT FROM create_orgs(%d)", orgs)
	helpers.ExecuteStatement(ccdb, ctx, createOrgStatement)

	createPublicServicePlansStatement := fmt.Sprintf("SELECT FROM create_services_and_plans(%v, %v, %v, %v, %v)",
		serviceOfferings, serviceBrokerId, servicePlansPublic, true, 0)
	helpers.ExecuteStatement(ccdb, ctx, createPublicServicePlansStatement)

	createPrivateServicePlansStatement := fmt.Sprintf("SELECT FROM create_services_and_plans(%v, %v, %v, %v, %v)",
		serviceOfferings, serviceBrokerId, servicePlansPrivateWithoutOrgs, false, 0)
	helpers.ExecuteStatement(ccdb, ctx, createPrivateServicePlansStatement)

	createPrivateServicePlansWithOrgsStatement := fmt.Sprintf("SELECT FROM create_services_and_plans(%v, %v, %v, %v, %v)",
		serviceOfferings, serviceBrokerId, servicePlansPrivateWithOrgs, false, orgsPerLimitedServicePlan)
	helpers.ExecuteStatement(ccdb, ctx, createPrivateServicePlansWithOrgsStatement)

	regularUserGUID := helpers.GetUserGUID(testSetup.RegularUserContext(), testConfig)
	orgsAssignedToRegularUser := orgs/2
	assignUserAsOrgManager := fmt.Sprintf("SELECT FROM assign_user_as_org_manager('%s', %d)", regularUserGUID, orgsAssignedToRegularUser)
	helpers.ExecuteStatement(ccdb, ctx, assignUserAsOrgManager)

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

func TestServicePlans(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "ServicePlansTest Suite", []Reporter{helpers.ConfigureJsonReporter(t, &testConfig, "service-plans")})
}

func createServiceBroker(prefix string) int {
	serviceBrokerGuid := uuid.NewString()
	serviceBrokerName := fmt.Sprintf("%s-service-broker-%s", prefix, serviceBrokerGuid)
	brokerUrl := fmt.Sprintf("https://bommel-%s.bommel.sap.hana.ondemand.com", serviceBrokerName)
	authPassword := "bXlfc3VwZXJfZHVwZXJfbWVoZ2FfcGFzc3dvcmQK"
	createServiceBrokerStatement := fmt.Sprintf(
		"INSERT INTO service_brokers (guid, name, broker_url, auth_password) VALUES ('%s', '%s', '%s', '%s') RETURNING id",
		serviceBrokerGuid, serviceBrokerName, brokerUrl, authPassword)
	return helpers.ExecuteInsertStatement(ccdb, ctx, createServiceBrokerStatement)
}
