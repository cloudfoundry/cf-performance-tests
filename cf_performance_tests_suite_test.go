package cf_performance_tests

import (
	"testing"

	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

var testConfig Config = NewConfig()
var testSetup *workflowhelpers.ReproducibleTestSuiteSetup

var _ = BeforeSuite(func() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.cf-performance-tests")
	err := viper.ReadInConfig()
	Expect(err).NotTo(HaveOccurred())

	err = viper.Unmarshal(&testConfig)
	Expect(err).NotTo(HaveOccurred())

	testSetup = workflowhelpers.NewTestSuiteSetup(&testConfig)
	testSetup.Setup()
})

func TestCfPerformanceTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CfPerformanceTests Suite")
}
