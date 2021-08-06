package helpers

import (
	"fmt"
	"time"
)

type Users struct {
	Admin    User
	Existing User
}

type User struct {
	Username     string
	Password     string
	Client       string
	ClientSecret string
}

type Config struct {
	API                 string
	CfDeploymentVersion string `mapstructure:"cf_deployment_version"`
	CapiVersion			string `mapstructure:"capi_version"`
	UseHttp             bool   `mapstructure:"use_http"`
	SkipSslValidation   bool   `mapstructure:"skip_ssl_validation"`
	NamePrefix          string `mapstructure:"name_prefix"`

	LargePageSize int `mapstructure:"large_page_size"`
	Samples       int

	Users Users

	BasicTimeout time.Duration `mapstructure:"basic_timeout"`
	LongTimeout  time.Duration `mapstructure:"long_timeout"`
	CcdbConnection  string `mapstructure:"ccdb_connection"`
	UaadbConnection string `mapstructure:"uaadb_connection"`
}

func NewConfig() Config {
	return Config{
		UseHttp:           false,
		SkipSslValidation: false,
		NamePrefix:        "CPT",
		BasicTimeout:      30 * time.Second,
		LongTimeout:       120 * time.Second,
		LargePageSize:     500,
		Samples:           10,
	}
}

func (config Config) GetSkipSSLValidation() bool                     { return config.SkipSslValidation }
func (config Config) GetExistingOrganization() string                { return "" }
func (config Config) GetUseExistingOrganization() bool               { return false }
func (config Config) GetExistingSpace() string                       { return "" }
func (config Config) GetUseExistingSpace() bool                      { return false }
func (config Config) GetAddExistingUserToExistingSpace() bool        { return false }
func (config Config) GetUseExistingUser() bool                       { return false }
func (config Config) GetExistingUser() string                        { return config.Users.Existing.Username }
func (config Config) GetExistingUserPassword() string                { return config.Users.Existing.Password }
func (config Config) GetExistingClient() string                      { return config.Users.Existing.Client }
func (config Config) GetExistingClientSecret() string                { return config.Users.Existing.ClientSecret }
func (config Config) GetAdminUser() string                           { return config.Users.Admin.Username }
func (config Config) GetAdminPassword() string                       { return config.Users.Admin.Password }
func (config Config) GetAdminClient() string                         { return config.Users.Admin.Client }
func (config Config) GetAdminClientSecret() string                   { return config.Users.Admin.ClientSecret }
func (config Config) GetConfigurableTestPassword() string            { return "" }
func (config Config) GetNamePrefix() string                          { return config.NamePrefix }
func (config Config) GetScaledTimeout(t time.Duration) time.Duration { return t }
func (config Config) GetShouldKeepUser() bool                        { return false }

func (config Config) GetApiEndpoint() string {
	if config.UseHttp {
		return fmt.Sprintf("http://%s", config.API)
	} else {
		return fmt.Sprintf("https://%s", config.API)
	}
}
