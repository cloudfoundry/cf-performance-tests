package destinations

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"
	"os"

	"github.com/google/uuid"

	"github.com/cloudfoundry/cf-test-helpers/v2/cf"
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
var appGuid1 string
var appGuid2 string
var appName1 string
var appName2 string
var spaceGuid string

const test_version = "v1"

// diego seems to have a limitation here
// when binding more routes to an app the app does not start, or it will fail during staging already
const (
	routeMappings = 800
)

func setupAppAndSeedDB(appName string, spaceGuid string) string {
	data := fmt.Sprintf(`{
                           "name": "%s",
                           "relationships": {
                             "space": {
                               "data": {
                                 "guid": "%s"
                               }
                             }
                           }
                         }`, appName, spaceGuid)

	exitCode, appCreateBody := helpers.TimeCFCurlReturning(testConfig.LongTimeout, "-X", "POST", "/v3/apps", "-d", data)

	Expect(exitCode).To(Equal(0))
	Expect(appCreateBody).To(ContainSubstring("201 Created"))

	appCreateResponse := helpers.ParseCreateResponseBody(helpers.RemoveDebugOutput(appCreateBody))
	appGuid := appCreateResponse.GUID

	helpers.ImportStoredProcedures(ccdb, ctx, testConfig)

	createRouteMappingsStatement := fmt.Sprintf("create_routes_and_route_mappings_for_app('%s', '%s', '%s', %d)", appGuid, testSetup.GetOrganizationName(), spaceGuid, routeMappings)
	helpers.ExecuteStoredProcedure(ccdb, ctx, createRouteMappingsStatement, testConfig)

	log.Printf("Preparing app directory and files.")
	appDir := helpers.CreateAppFolder(appName1)
	defer os.RemoveAll(appDir)

	log.Printf("Pushing app `%s`", appName)
	pushResult := cf.Push(appName, "-i", "1", "-b", "staticfile_buildpack", "-p", appDir).Wait(120)
	if pushResult.ExitCode() != 0 {
		panic("Push not successful")
	}

	return appGuid
}

var _ = BeforeSuite(func() {
	testSetup = workflowhelpers.NewTestSuiteSetup(&testConfig)
	testSetup.Setup()
	ccdb, uaadb, ctx = helpers.OpenDbConnections(testConfig)

	spaceName := testSetup.TestSpace.SpaceName()
	spaceGuids := helpers.GetGUIDs(testSetup.AdminUserContext(), testConfig, fmt.Sprintf("/v3/spaces?names=%s", spaceName))
	spaceGuid = spaceGuids[0]

	appName1 = fmt.Sprintf("%s-app-%s", testConfig.GetNamePrefix(), uuid.NewString())
	appName2 = fmt.Sprintf("%s-app-%s", testConfig.GetNamePrefix(), uuid.NewString())

	appGuid1 = setupAppAndSeedDB(appName1, spaceGuid)
	appGuid2 = setupAppAndSeedDB(appName2, spaceGuid)

	helpers.AnalyzeDB(ccdb, ctx, testConfig)
	log.Printf("Finished seeding database.")
})

var _ = AfterSuite(func() {
	log.Printf("Deleting app `%s`\n", appName1)
	helpers.TimeCFCurlReturning(testConfig.LongTimeout, "-X", "DELETE", fmt.Sprintf("/v3/apps/%s", appGuid1))

	log.Printf("Deleting app `%s`\n", appName2)
	helpers.TimeCFCurlReturning(testConfig.LongTimeout, "-X", "DELETE", fmt.Sprintf("/v3/apps/%s", appGuid2))

	log.Printf("Starting cleanup testdata...")
	helpers.CleanupTestData(ccdb, uaadb, ctx, testConfig)
	log.Printf("Finished cleanup testdata.")
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

var _ = ReportAfterSuite("Destinations test suite", func(report types.Report) {
	helpers.GenerateReports(helpers.ConfigureJsonReporter(&testConfig, "destinations", "destinations", test_version), report)
})

func TestDestinations(t *testing.T) {
	helpers.LoadConfig(&testConfig)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Destinations Test Suite")
}
