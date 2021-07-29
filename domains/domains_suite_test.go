package domains

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
	sharedDomains = 10
	privateDomains = 10
)

var _ = BeforeSuite(func() {
	testSetup = workflowhelpers.NewTestSuiteSetup(&testConfig)
	testSetup.Setup()
	ccdb, uaadb, ctx = helpers.OpenDbConnections(testConfig.CcdbConnection, testConfig.UaadbConnection)

	quotaId := helpers.ExecuteSelectStatementOneRow(ccdb, ctx, "SELECT id FROM quota_definitions WHERE name = 'default'")
	var organizationIds []int

	for i := 0; i < orgs; i++ {
		guid := uuid.New()
		name := testConfig.NamePrefix + "-org-" + guid.String()
		statement := "INSERT INTO organizations (guid, name, quota_definition_id) VALUES ($1, $2, $3) RETURNING id"
		organizationId := helpers.ExecutePreparedInsertStatement(ccdb, ctx, statement, guid.String(), name, quotaId)
		organizationIds = append(organizationIds, organizationId)
	}
	for i := 0; i<sharedDomains; i++ {
		sharedDomainGuid := uuid.New()
		sharedDomainName := testConfig.NamePrefix + "-shareddomain-" + sharedDomainGuid.String()
		statement := "INSERT INTO domains (guid, name) VALUES ($1, $2) RETURNING id"
		helpers.ExecutePreparedInsertStatement(ccdb, ctx, statement, sharedDomainGuid.String(), sharedDomainName)
	}

	for i := 0; i<privateDomains; i++ {
		privateDomainGuid := uuid.New()
		privateDomainName := testConfig.NamePrefix + "-privatedomain-" + privateDomainGuid.String()
		owningOrganizationId := organizationIds[rand.Intn(len(organizationIds))]
		statement := "INSERT INTO domains (guid, name, owning_organization_id) VALUES ($1, $2, $3) RETURNING id"
		helpers.ExecutePreparedInsertStatement(ccdb, ctx, statement, privateDomainGuid.String(), privateDomainName, owningOrganizationId)
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

func TestDomains(t *testing.T) {
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
	jsonReporter := helpers.NewJsonReporter(fmt.Sprintf("../test-results/domains-test-results-%d.json", timestamp), testConfig.CfDeploymentVersion, timestamp)

	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "DomainsTest Suite", []Reporter{jsonReporter})
}

