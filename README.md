# cf-performance-tests
Performance tests for the Cloud Foundry API (Cloud Controller).

## Goals
These tests are intended to:
* Help debug slow endpoints
* Analyse performance impact of changes to Cloud Controller codebase
* Ensure that query times do not scale exponentially with database size

## Anti-goals
These tests are not intended to:
* Test parallelism of a specific webserver
* Load test the Cloud Controller
* Assist with scaling decisions of CAPI deployments

## Running tests
Tests in this repository are written using [Ginkgo](https://onsi.github.io/ginkgo/) using the [Measure](https://pkg.go.dev/github.com/onsi/ginkgo#Measure) spec definition to time API calls across multiple attempts, tracking the minimum, maximum durations as well as the standard deviation.

The test suite uses [Viper](https://github.com/spf13/viper) for configuration of parameters such as API endpoint, credentials etc. Viper will look for a configuration file in both the `$HOME` directory and the working directory that tests are invoked from. See the [Config struct](helpers/config.go) for available configuration parameters.

To run the tests, create a configuration file that Viper can find, e.g. `config.yml` in the project's root folder:
```yaml
api: "<CF API endpoint>"
use_http: false  (the default value)
skip_ssl_validation: false (the default value)
cf_deployment_version: "<used for generated report>"
capi_version: "<used for generated report>"
large_page_size: 500  (the default value)
large_elements_filter: 100  (the default value)
samples: 5  (the default value)
basictimeout: 30  (the default value)
longtimeout: 120  (the default value)
users:
  admin:
    username: "<admin username>"
    password: "<admin password>"
  existing:  (optional block)
    username: "<non-admin username>"
    password: "<non-admin password>"
ccdb_connection: "<connection string for CCDB>"
uaadb_connection: "<connection string for UAADB>"  (optional, used to cleanup the created test user)
```
The `name_prefix` string must match the prefix of the test resources names. Note that some performance tests delete lists of resources. Using a `name_prefix` ensures that only test resources are deleted.

Then run:
```bash
ginkgo -r
```
