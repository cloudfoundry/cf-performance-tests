package helpers

import (
	"encoding/json"
	"strings"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

type APIResponse struct {
	Resources []struct {
		GUID string `json:"guid"`
		Name string `json:"name"`
	} `json:"resources"`
}

func GetGUIDs(user workflowhelpers.UserContext, testConfig Config, endpoint string) []string {
	var session *Session
	var resp APIResponse
	var guids []string
	workflowhelpers.AsUser(user, testConfig.BasicTimeout, func() {
		session = cf.Cf("curl", endpoint)
		Expect(session.Wait(testConfig.BasicTimeout)).To(Exit(0))
	})
	json.Unmarshal(session.Out.Contents(), &resp)
	for _, item := range resp.Resources {
		// do not select non-test resources (e.g. the default CF orgs or security groups)
		if strings.HasPrefix(item.Name, testConfig.GetNamePrefix()) {
			guids = append(guids, item.GUID)
		}
	}
	return guids
}
