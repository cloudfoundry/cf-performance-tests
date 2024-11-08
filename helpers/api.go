package helpers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudfoundry/cf-test-helpers/v2/cf"
	"github.com/cloudfoundry/cf-test-helpers/v2/workflowhelpers"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

type APIResponse struct {
	Pagination struct {
		TotalPages   int `json:"total_pages"`
		TotalResults int `json:"total_results"`
	} `json:"pagination"`
	Resources []struct {
		GUID     string `json:"guid"`
		Name     string `json:"name"`
		UserName string `json:"username"`
	} `json:"resources"`
}

type APICreateResponse struct {
	GUID string `json:"guid"`
}

type DestinationsCreateResponse struct {
	Destinations []struct {
		GUID string `json:"guid"`
	} `json:"destinations"`
}

func ParseDestinationsCreateResponseBody(body []byte) *DestinationsCreateResponse {
	var resp *DestinationsCreateResponse
	err := json.Unmarshal(body, &resp)

	if err != nil {

		return nil
	}
	return resp
}

func ParseCreateResponseBody(body []byte) *APICreateResponse {
	var resp *APICreateResponse
	err := json.Unmarshal(body, &resp)

	if err != nil {

		return nil
	}
	return resp
}

func RemoveDebugOutput(body []byte) []byte {
	var jsons [][]byte
	start := -1
	level := 0

	for i, r := range body {
		if r == '{' {
			if start == -1 {
				start = i
			}
			level++
		} else if r == '}' {
			level--
			if level == 0 {
				jsons = append(jsons, body[start:i+1])
				start = -1
			}
		}
	}

	// For POST there will be 2 jsons in the output. Return the second one, which is the response.
	if len(jsons) > 1 {
		return jsons[1]
	} else { // For GET there is just one json
		return jsons[0]
	}
}

func ParseResponseBody(body []byte) *APIResponse {
	var resp *APIResponse
	err := json.Unmarshal(body, &resp)
	if err != nil {
		return nil
	}
	return resp
}

func ApiCall(user workflowhelpers.UserContext, testConfig Config, endpoint string) *APIResponse {
	var session *Session
	workflowhelpers.AsUser(user, testConfig.BasicTimeout, func() {
		session = cf.Cf("curl", "--fail", endpoint).Wait(testConfig.BasicTimeout)
		Expect(session).To(Exit(0))
	})
	return ParseResponseBody(session.Out.Contents())
}

func GetGUIDs(user workflowhelpers.UserContext, testConfig Config, endpoint string) []string {
	var guids []string
	resp := ApiCall(user, testConfig, endpoint)
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

func CreateAppFolder(appName string) string {
	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "app")
	if err != nil {
		panic(err)
	}

	log.Printf("Created a temporary directory: %s", tempDir)

	// Create a new file in the temporary directory
	indexFile, err := os.Create(fmt.Sprintf("%s/%s", tempDir, "index.html"))
	if err != nil {
		panic(err)
	}

	log.Printf("Created a temporary file: %s", indexFile.Name())

	msg := []byte("Hello, World!")
	if _, err := indexFile.Write(msg); err != nil {
		panic(err)
	}

	// You need to close the file after writing
	if err := indexFile.Close(); err != nil {
		panic(err)
	}

	// Get the directory of the temporary file
	dir := filepath.Dir(indexFile.Name())

	return dir
}
