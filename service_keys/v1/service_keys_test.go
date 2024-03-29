package service_keys

import (
	"fmt"
	"math/rand"

	"github.com/cloudfoundry/cf-test-helpers/v2/workflowhelpers"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"

	"github.com/cloudfoundry/cf-performance-tests/helpers"
)

var _ = Describe("service keys", func() {
	Describe("individually", func() {
		Describe("as admin", func() {
			Describe("with exhausted service keys quota", func() {
				var serviceInstanceGUID string
				BeforeEach(func() {
					serviceInstanceGUIDs := helpers.GetGUIDs(testSetup.AdminUserContext(), testConfig, fmt.Sprintf("/v3/service_instances?space_guids=%s", spaceWithExhaustedServiceKeysGUID))
					Expect(serviceInstanceGUIDs).NotTo(BeNil())
					serviceInstanceGUID = serviceInstanceGUIDs[rand.Intn(len(serviceInstanceGUIDs))]
				})

				It("posts /v3/service_credential_bindings as admin  ", func() {
					experiment := gmeasure.NewExperiment("individually::as admin::with exhausted service keys quota::POST /v3/service_credential_bindings")
					AddReportEntry(experiment.Name, experiment)

					workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
						experiment.Sample(func(idx int) {
							experiment.MeasureDuration("POST /v3/service_credential_bindings", func() {
								serviceKeyName := fmt.Sprintf("%s-service-key-%s", testConfig.GetNamePrefix(), uuid.NewString())
								data := fmt.Sprintf(`{"type":"key","name":"%s","relationships":{"service_instance":{"data":{"guid":"%s"}}}}`, serviceKeyName, serviceInstanceGUID)

								exitCode, body := helpers.TimeCFCurlReturning(testConfig.BasicTimeout, "-X", "POST", "-d", data, "/v3/service_credential_bindings")
								Expect(exitCode).To(Equal(22))
								Expect(body).To(ContainSubstring("You have exceeded your organization's limit for service binding of type key."))
							})
						}, gmeasure.SamplingConfig{N: testConfig.Samples})
					})
				})
			})

			Describe("with unlimited service keys quota", func() {
				var serviceInstanceGUID string
				BeforeEach(func() {
					serviceInstanceGUIDs := helpers.GetGUIDs(testSetup.AdminUserContext(), testConfig, fmt.Sprintf("/v3/service_instances?space_guids=%s", spaceWithUnlimitedServiceKeysGUID))
					Expect(serviceInstanceGUIDs).NotTo(BeNil())
					serviceInstanceGUID = serviceInstanceGUIDs[rand.Intn(len(serviceInstanceGUIDs))]
				})

				It("posts /v3/service_credential_bindings as admin  ", func() {
					experiment := gmeasure.NewExperiment("individually::as admin::with unlimited service keys quota::POST /v3/service_credential_bindings")

					AddReportEntry(experiment.Name, experiment)

					workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
						experiment.Sample(func(idx int) {
							experiment.MeasureDuration("POST /v3/service_credential_bindings", func() {
								serviceKeyName := fmt.Sprintf("%s-service-key-%s", testConfig.GetNamePrefix(), uuid.NewString())
								data := fmt.Sprintf(`{"type":"key","name":"%s","relationships":{"service_instance":{"data":{"guid":"%s"}}}}`, serviceKeyName, serviceInstanceGUID)

								exitCode, body := helpers.TimeCFCurlReturning(testConfig.BasicTimeout, "-X", "POST", "-d", data, "/v3/service_credential_bindings")
								Expect(exitCode).To(Equal(0))
								Expect(body).To(ContainSubstring("202 Accepted"))
								// Note: The created VCAP::CloudController::V3::CreateBindingAsyncJob fails, as there is no real service broker to handle it.
							})
						}, gmeasure.SamplingConfig{N: testConfig.Samples})
					})
				})
			})
		})
	})
})
