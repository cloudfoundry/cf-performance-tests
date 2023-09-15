package helpers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudfoundry/cf-test-helpers/v2/cf"
	"github.com/cloudfoundry/cf-test-helpers/v2/workflowhelpers"
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

func RemoveDebugOutput(body []byte) []byte {
	scanner := bufio.NewScanner(bytes.NewReader(body))
	var buffer bytes.Buffer
	write := false

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "{") {
			write = true
		}
		if write {
			buffer.WriteString(line + "\n")
		}
	}

	return buffer.Bytes()
}

func ParseResponseBody(body []byte) *APIResponse {
	var resp *APIResponse
	err := json.Unmarshal(body, &resp)
	if err != nil {
		return nil
	}
	return resp
}

func apiCall(user workflowhelpers.UserContext, testConfig Config, endpoint string) *APIResponse {
	var session *Session
	workflowhelpers.AsUser(user, testConfig.BasicTimeout, func() {
		session = cf.Cf("curl", "--fail", endpoint).Wait(testConfig.BasicTimeout)
		Expect(session).To(Exit(0))
	})
	return ParseResponseBody(session.Out.Contents())
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

func TimeCFCurl(timeout time.Duration, curlArguments ...string) {
	exitCode, _ := TimeCFCurlReturning(timeout, curlArguments...)
	Expect(exitCode).To(Equal(0))
}

func TimeCFCurlReturning(timeout time.Duration, curlArguments ...string) (int, []byte) {
	var args = []string{"curl", "--fail", "-v"}
	args = append(args, curlArguments...)
	result := cf.Cf(args...).Wait(timeout)

	return result.ExitCode(), result.Out.Contents()
}
