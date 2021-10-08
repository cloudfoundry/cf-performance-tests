package helpers

import (
	"encoding/json"
	"fmt"
	"github.com/onsi/ginkgo"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

type APIResponse struct {
	Pagination struct {
		TotalResults int `json:"total_results"`
	}
	Resources []struct {
		GUID     string `json:"guid"`
		Name     string `json:"name"`
		UserName string `json:"username"`
	} `json:"resources"`
}

func apiCall(user workflowhelpers.UserContext, testConfig Config, endpoint string) *APIResponse {
	var session *Session
	var resp *APIResponse
	workflowhelpers.AsUser(user, testConfig.BasicTimeout, func() {
		session = cf.Cf("curl", "--fail", endpoint).Wait(testConfig.BasicTimeout)
		Expect(session).To(Exit(0))
	})
	err := json.Unmarshal(session.Out.Contents(), &resp)
	if err != nil {
		return nil
	}
	return resp
}

func GetGUIDs(user workflowhelpers.UserContext, testConfig Config, endpoint string) []string {
	var guids []string
	resp := apiCall(user, testConfig, endpoint)
	if resp != nil {
		for _, item := range resp.Resources {
			// do not select non-test resources (e.g. the default CF orgs or security groups)
			name := item.Name
			if name == "" {
				name = item.UserName
			}
			if strings.HasPrefix(name, testConfig.GetNamePrefix()+"-") {
				guids = append(guids, item.GUID)
			}
		}
	}
	return guids
}

func GetUserGUID(user workflowhelpers.UserContext, testConfig Config) string {
	userGUIDs := GetGUIDs(user, testConfig, fmt.Sprintf("/v3/users?usernames=%s", user.Username))
	if userGUIDs != nil {
		Expect(len(userGUIDs)).To(Equal(1))
		return userGUIDs[0]
	}
	return ""
}

func WaitToFail(user workflowhelpers.UserContext, testConfig Config, endpoint string) {
	workflowhelpers.AsUser(user, testConfig.BasicTimeout, func() {
		for exitCode := -1; exitCode <= 0; {
			exitCode = cf.Cf("curl", "--fail", endpoint).Wait(testConfig.BasicTimeout).ExitCode()
		}
	})
}

func GetTotalResults(user workflowhelpers.UserContext, testConfig Config, endpoint string) int {
	var totalResults int
	resp := apiCall(user, testConfig, endpoint)
	if resp != nil {
		totalResults = resp.Pagination.TotalResults
	}
	return totalResults
}

func GetXRuntimeHeader(response []byte) float64 {
	responseString := string(response)
	regexp := regexp.MustCompile(`X-Runtime: (\d+.?\d+)`)
	matches := regexp.FindStringSubmatch(responseString)
	if len(matches) == 0 {
		panic("Response did not contain `X-Runtime` header")
	}

	runtime, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		panic("Runtime could not be parsed from string to float64")
	}
	return runtime
}

func TimeCFCurl(b ginkgo.Benchmarker, timeout time.Duration, curlArguments ...string) {

	var args = []string{"curl", "--fail", "-v"}
	args = append(args, curlArguments...)
	result := cf.Cf(args...
	).Wait(timeout)
	Expect(
		result,
	).To(Exit(0))

	runtime := GetXRuntimeHeader(result.Out.Contents())
	b.RecordValue("request time", runtime)
}