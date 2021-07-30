package main

import (
	"github.com/cloudfoundry-incubator/cf-performance-tests/helpers"
	"log"
)

const (
	CcdbConnection = "postgres://cloud_controller:fjLip8fvl0nV97OpvI7pJhSV4KQsmA@localhost:5524/cloud_controller?sslmode=disable"
	UaadbConnection  = "postgres://uaa:2GBRCiNFiXkDLe9KSBHranhQIz9l7P@localhost:5524/uaa?sslmode=disable"
)

func main() {
	log.Print("Starting database test...")
	ccdb, uaadb, ctx := helpers.OpenDbConnections(CcdbConnection, UaadbConnection)
	helpers.CleanupTestData(ccdb, uaadb, ctx)
	log.Print("Finished.")
}
