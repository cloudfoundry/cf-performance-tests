package isolation_segments

import (
	"fmt"
	"math/rand"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-performance-tests/helpers"
)

var _ = Describe("isolation segments", func() {
	Describe("GET /v3/isolation_segments", func() {
		Measure("as admin", func(b Benchmarker) {
			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
				b.Time("request time", func() {
					Expect(
						cf.Cf(
							"curl", "--fail", "/v3/isolation_segments",
						).Wait(testConfig.BasicTimeout),
					).To(Exit(0))
				})
			})
		}, testConfig.Samples)

		Measure("as regular user", func(b Benchmarker) {
			workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
				b.Time("request time", func() {
					Expect(
						cf.Cf(
							"curl", "--fail", "/v3/isolation_segments",
						).Wait(testConfig.BasicTimeout),
					).To(Exit(0))
				})
			})
		}, testConfig.Samples)

		Measure("as admin with large page size", func(b Benchmarker) {
			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.LongTimeout, func() {
				b.Time("request time", func() {
					Expect(
						cf.Cf(
							"curl", "--fail", fmt.Sprintf("/v3/isolation_segments?per_page=%d", testConfig.LargePageSize),
						).Wait(testConfig.LongTimeout),
					).To(Exit(0))
				})
			})
		}, testConfig.Samples)
	})

	Describe("GET /v3/isolation_segments/:guid/relationships/organizations", func() {
		Measure("as admin", func(b Benchmarker) {
			isolationSegmentGUIDs := helpers.GetGUIDs(testSetup.AdminUserContext(), testConfig, "/v3/isolation_segments")
			Expect(isolationSegmentGUIDs).NotTo(BeNil())
			isolationSegmentGUID := isolationSegmentGUIDs[rand.Intn(len(isolationSegmentGUIDs))]
			workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
				b.Time("request time", func() {
					Expect(
						cf.Cf(
							"curl", "--fail", fmt.Sprintf("/v3/isolation_segments/%s/relationships/organizations", isolationSegmentGUID),
						).Wait(testConfig.BasicTimeout),
					).To(Exit(0))
				})
			})
		}, testConfig.Samples)

		Measure("as regular user", func(b Benchmarker) {
			isolationSegmentGUIDs := helpers.GetGUIDs(testSetup.RegularUserContext(), testConfig, "/v3/isolation_segments")
			Expect(isolationSegmentGUIDs).NotTo(BeNil())
			isolationSegmentGUID := isolationSegmentGUIDs[rand.Intn(len(isolationSegmentGUIDs))]
			workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
				b.Time("request time", func() {
					Expect(
						cf.Cf(
							"curl", "--fail", fmt.Sprintf("/v3/isolation_segments/%s/relationships/organizations", isolationSegmentGUID),
						).Wait(testConfig.BasicTimeout),
					).To(Exit(0))
				})
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
					b.Time("request time", func() {
						Expect(
							cf.Cf(
								"curl", "--fail", fmt.Sprintf("/v3/isolation_segments/%s", isolationSegmentGUID),
							).Wait(testConfig.BasicTimeout),
						).To(Exit(0))
					})
				})
			}, testConfig.Samples)

			Measure("PATCH /v3/isolation_segments/:guid", func(b Benchmarker) {
				workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
					b.Time("request time", func() {
						data := fmt.Sprintf(`{"name":"perf-updated-isolation-segment-%s"}`, isolationSegmentGUID)
						Expect(
							cf.Cf(
								"curl", "--fail", "-X", "PATCH", "-d", data, fmt.Sprintf("/v3/isolation_segments/%s", isolationSegmentGUID),
							).Wait(testConfig.BasicTimeout),
						).To(Exit(0))
					})
				})
			}, testConfig.Samples)
		})

		Describe("as regular user", func() {
			Measure("GET /v3/isolation_segments/:guid", func(b Benchmarker) {
				isolationSegmentGUIDs := helpers.GetGUIDs(testSetup.RegularUserContext(), testConfig, "/v3/isolation_segments")
				Expect(isolationSegmentGUIDs).NotTo(BeNil())
				isolationSegmentGUID := isolationSegmentGUIDs[rand.Intn(len(isolationSegmentGUIDs))]
				workflowhelpers.AsUser(testSetup.RegularUserContext(), testConfig.BasicTimeout, func() {
					b.Time("request time", func() {
						Expect(
							cf.Cf(
								"curl", "--fail", fmt.Sprintf("/v3/isolation_segments/%s", isolationSegmentGUID),
							).Wait(testConfig.BasicTimeout),
						).To(Exit(0))
					})
				})
			}, testConfig.Samples)
		})
	})
})
