package service_keys

import (
	"fmt"
	"math/rand"

	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

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

				Measure("POST /v3/service_credential_bindings", func(b Benchmarker) {
					workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
						serviceKeyName := fmt.Sprintf("%s-service-key-%s", testConfig.GetNamePrefix(), uuid.NewString())
						data := fmt.Sprintf(`{"type":"key","name":"%s","relationships":{"service_instance":{"data":{"guid":"%s"}}}}`, serviceKeyName, serviceInstanceGUID)
						exitCode, body := helpers.TimeCFCurlReturning(b, testConfig.BasicTimeout, "-X", "POST", "-d", data, "/v3/service_credential_bindings")
						Expect(exitCode).To(Equal(22))
						Expect(body).To(ContainSubstring("You have exceeded your organization's limit for service binding of type key."))
					})
				}, testConfig.Samples)
			})

			Describe("with unlimited service keys quota", func() {
				var serviceInstanceGUID string
				BeforeEach(func() {
					serviceInstanceGUIDs := helpers.GetGUIDs(testSetup.AdminUserContext(), testConfig, fmt.Sprintf("/v3/service_instances?space_guids=%s", spaceWithUnlimitedServiceKeysGUID))
					Expect(serviceInstanceGUIDs).NotTo(BeNil())
					serviceInstanceGUID = serviceInstanceGUIDs[rand.Intn(len(serviceInstanceGUIDs))]
				})

				Measure("POST /v3/service_credential_bindings", func(b Benchmarker) {
					workflowhelpers.AsUser(testSetup.AdminUserContext(), testConfig.BasicTimeout, func() {
						serviceKeyName := fmt.Sprintf("%s-service-key-%s", testConfig.GetNamePrefix(), uuid.NewString())
						data := fmt.Sprintf(`{"type":"key","name":"%s","relationships":{"service_instance":{"data":{"guid":"%s"}}}}`, serviceKeyName, serviceInstanceGUID)
						exitCode, body := helpers.TimeCFCurlReturning(b, testConfig.BasicTimeout, "-X", "POST", "-d", data, "/v3/service_credential_bindings")
						Expect(exitCode).To(Equal(0))
						Expect(body).To(ContainSubstring("202 Accepted"))
						// Note: The created VCAP::CloudController::V3::CreateBindingAsyncJob fails, as there is no real service broker to handle it.
					})
				}, testConfig.Samples)
			})
		})
	})
})
