package isolation_segments

import (
	"fmt"
	"math/rand"

	"github.com/cloudfoundry/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cf-performance-tests/helpers"
)

var _ = Describe("isolation segments", func() {
	Describe("GET /v3/isolation_segments", func() {
		Measure("as admin", func(b Benchmarker) {
			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
				helpers.TimeCFCurl(b, testConfig.BasicTimeout, "/v3/isolation_segments")
			})
		}, testConfig.Samples)

		Measure("as regular user", func(b Benchmarker) {
			workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
				helpers.TimeCFCurl(b, testConfig.BasicTimeout, "/v3/isolation_segments")
			})
		}, testConfig.Samples)

		Measure(fmt.Sprintf("as admin with page size %d", testConfig.LargePageSize), func(b Benchmarker) {
			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
				helpers.TimeCFCurl(b, testConfig.LongTimeout, fmt.Sprintf("/v3/isolation_segments?per_page=%d", testConfig.LargePageSize))
			})
		}, testConfig.Samples)
	})

	Describe("GET /v3/isolation_segments/:guid/relationships/organizations", func() {
		Measure("as admin", func(b Benchmarker) {
			isolationSegmentGUIDs := helpers.GetGUIDs(testSetup.AdminUserContext(), testConfig, "/v3/isolation_segments")
			Expect(isolationSegmentGUIDs).NotTo(BeNil())
			isolationSegmentGUID := isolationSegmentGUIDs[rand.Intn(len(isolationSegmentGUIDs))]
			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
				helpers.TimeCFCurl(b, testConfig.BasicTimeout, fmt.Sprintf("/v3/isolation_segments/%s/relationships/organizations", isolationSegmentGUID))
			})
		}, testConfig.Samples)

		Measure("as regular user", func(b Benchmarker) {
			isolationSegmentGUIDs := helpers.GetGUIDs(testSetup.RegularUserContext(), testConfig, "/v3/isolation_segments")
			Expect(isolationSegmentGUIDs).NotTo(BeNil())
			isolationSegmentGUID := isolationSegmentGUIDs[rand.Intn(len(isolationSegmentGUIDs))]
			workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
				helpers.TimeCFCurl(b, testConfig.BasicTimeout, fmt.Sprintf("/v3/isolation_segments/%s/relationships/organizations", isolationSegmentGUID))
			})
		}, testConfig.Samples)
	})

	Describe("individually", func() {
		Describe("as admin", func() {
			var isolationSegmentGUID string
			BeforeEach(func() {
				isolationSegmentGUIDs := helpers.GetGUIDs(testSetup.AdminUserContext(), testConfig, "/v3/isolation_segments")
				Expect(isolationSegmentGUIDs).NotTo(BeNil())
				isolationSegmentGUID = isolationSegmentGUIDs[rand.Intn(len(isolationSegmentGUIDs))]
			})

			Measure("GET /v3/isolation_segments/:guid", func(b Benchmarker) {
				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					helpers.TimeCFCurl(b, testConfig.BasicTimeout, fmt.Sprintf("/v3/isolation_segments/%s", isolationSegmentGUID))
				})
			}, testConfig.Samples)

			Measure("PATCH /v3/isolation_segments/:guid", func(b Benchmarker) {
				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					data := fmt.Sprintf(`{"name":"%s-updated-isolation-segment-%s"}`, testConfig.GetNamePrefix(), isolationSegmentGUID)
					helpers.TimeCFCurl(b, testConfig.BasicTimeout, "-X", "PATCH", "-d", data, fmt.Sprintf("/v3/isolation_segments/%s", isolationSegmentGUID))
				})
			}, testConfig.Samples)
		})

		Describe("as regular user", func() {
			Measure("GET /v3/isolation_segments/:guid", func(b Benchmarker) {
				isolationSegmentGUIDs := helpers.GetGUIDs(testSetup.RegularUserContext(), testConfig, "/v3/isolation_segments")
				Expect(isolationSegmentGUIDs).NotTo(BeNil())
				isolationSegmentGUID := isolationSegmentGUIDs[rand.Intn(len(isolationSegmentGUIDs))]
				workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
					helpers.TimeCFCurl(b, testConfig.BasicTimeout, fmt.Sprintf("/v3/isolation_segments/%s", isolationSegmentGUID))
				})
			}, testConfig.Samples)
		})
	})
})
