package helpers

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"testing"
	"time"
)

type User struct {
	Username     string
	Password     string
	Client       string
	ClientSecret string
}

type Users struct {
	Admin    User
	Existing User
}

type Config struct {
	API                 string
	UseHttp             bool   `mapstructure:"use_http"`
	SkipSslValidation   bool   `mapstructure:"skip_ssl_validation"`
	CfDeploymentVersion string `mapstructure:"cf_deployment_version"`
	CapiVersion         string `mapstructure:"capi_version"`
	LargePageSize       int    `mapstructure:"large_page_size"`
	LargeElementsFilter int    `mapstructure:"large_elements_filter"`
	Samples             int
	BasicTimeout        time.Duration `mapstructure:"basic_timeout"`
	LongTimeout         time.Duration `mapstructure:"long_timeout"`
	Users               Users
	CcdbConnection      string `mapstructure:"ccdb_connection"`
	UaadbConnection     string `mapstructure:"uaadb_connection"`
	ResultsFolder       string `mapstructure:"results_folder"`
}

func NewConfig() Config {
	return Config{
		LargePageSize:       500,
		LargeElementsFilter: 100,
		Samples:             5,
		BasicTimeout:        30 * time.Second,
		LongTimeout:         120 * time.Second,
	}
}

func (config Config) GetAdminUser() string             { return config.Users.Admin.Username }
func (config Config) GetAdminPassword() string         { return config.Users.Admin.Password }
func (config Config) GetUseExistingOrganization() bool { return false }
func (config Config) GetUseExistingSpace() bool        { return false }
func (config Config) GetExistingOrganization() string  { return "" }
func (config Config) GetExistingSpace() string         { return "" }
func (config Config) GetUseExistingUser() bool {
	return config.Users.Existing.Username != "" && config.Users.Existing.Password != ""
}
func (config Config) GetExistingUser() string             { return config.Users.Existing.Username }
func (config Config) GetExistingUserPassword() string     { return config.Users.Existing.Password }
func (config Config) GetShouldKeepUser() bool             { return false }
func (config Config) GetConfigurableTestPassword() string { return "" }
func (config Config) GetAdminClient() string              { return config.Users.Admin.Client }
func (config Config) GetAdminClientSecret() string        { return config.Users.Admin.ClientSecret }
func (config Config) GetExistingClient() string           { return config.Users.Existing.Client }
func (config Config) GetExistingClientSecret() string     { return config.Users.Existing.ClientSecret }
func (config Config) GetApiEndpoint() string {
	if config.UseHttp {
		return fmt.Sprintf("http://%s", config.API)
	} else {
		return fmt.Sprintf("https://%s", config.API)
	}
}
func (config Config) GetSkipSSLValidation() bool                     { return config.SkipSslValidation }
func (config Config) GetNamePrefix() string                          { return "perf" }
func (config Config) GetScaledTimeout(t time.Duration) time.Duration { return t }
func (config Config) GetResultsFolder() string { return config.ResultsFolder }

func ConfigureJsonReporter(t *testing.T, testConfig *Config, testSuiteName string) *JsonReporter {
	viper.SetConfigName("config")
	viper.AddConfigPath("../../")
	viper.AddConfigPath("$HOME/.cf-performance-tests")
	viper.SetDefault("results_folder", "../../test-results")
	err := viper.ReadInConfig()
	if err != nil {
		t.Fatalf("error loading config: %s", err.Error())
	}

	err = viper.Unmarshal(testConfig)
	if err != nil {
		t.Fatalf("error parsing config: %s", err.Error())
	}

	resultsFolder := fmt.Sprintf("%s/%s-test-results/v1", testConfig.GetResultsFolder(), testSuiteName)
	err = os.MkdirAll(resultsFolder, os.ModePerm)
	if err != nil {
		t.Fatalf("Cannot create Directory: %s", err.Error())
	}

	timestamp := time.Now().Unix()
	return NewJsonReporter(fmt.Sprintf("%s/%s-test-results-%d.json", resultsFolder, testSuiteName, timestamp), testConfig.CfDeploymentVersion, testConfig.CapiVersion, timestamp)
}