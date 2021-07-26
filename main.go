package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"log"
)

const TestDataPrefix = "perf-%"

func main() {
	log.Print("Starting database test...")
	cleanupTestData("postgres://cloud_controller:fjLip8fvl0nV97OpvI7pJhSV4KQsmA@localhost:5524/cloud_controller?sslmode=disable")
	log.Print("Finished.")
}

func cleanupTestData(connection string) {
	db, err := sql.Open("pgx", connection)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()

	deleteStatements := []string{
		"DELETE FROM route_mappings USING routes WHERE routes.guid = route_mappings.route_guid AND routes.host LIKE '%s'",
		"DELETE FROM routes WHERE host LIKE '%s'",
		"DELETE FROM domain_annotations USING domains WHERE domain_annotations.resource_guid = domains.guid AND domains.name LIKE '%s'",
		"DELETE FROM domains WHERE name LIKE '%s'",
		"DELETE FROM service_bindings USING apps WHERE apps.guid = service_bindings.app_guid AND apps.name LIKE '%s'",
		"DELETE FROM route_mappings USING apps WHERE apps.guid = route_mappings.app_guid AND apps.name LIKE '%s'",
		"DELETE FROM apps WHERE name LIKE '%s'",
		"DELETE FROM service_bindings USING service_instances WHERE service_instances.guid = service_bindings.service_instance_guid AND service_instances.name LIKE '%s'",
		"DELETE FROM service_instances WHERE name LIKE '%s'",
		"DELETE FROM security_groups_spaces USING security_groups WHERE security_groups_spaces.security_group_id = security_groups.id AND security_groups.name LIKE '%s'",
		"DELETE FROM security_groups_spaces USING spaces WHERE security_groups_spaces.space_id = spaces.id AND spaces.name LIKE '%s'",
		"DELETE FROM security_groups WHERE name LIKE '%s'",
		"DELETE FROM spaces_developers USING spaces WHERE spaces_developers.space_id = spaces.id AND spaces.name LIKE '%s'",
		"DELETE FROM spaces_managers USING spaces WHERE spaces_managers.space_id = spaces.id AND spaces.name LIKE '%s'",
		"DELETE FROM spaces_auditors USING spaces WHERE spaces_auditors.space_id = spaces.id AND spaces.name LIKE '%s'",
		"DELETE FROM spaces WHERE name LIKE '%s'",
		"DELETE FROM service_plan_visibilities USING organizations WHERE service_plan_visibilities.organization_id = organizations.id AND organizations.name LIKE '%s'",
		"DELETE FROM organizations_users USING organizations WHERE organizations_users.organization_id = organizations.id AND organizations.name LIKE '%s'",
		"DELETE FROM organizations WHERE name LIKE '%s'",
	}
	for _, statement := range deleteStatements {
		executeStatement(db, ctx, fmt.Sprintf(statement, TestDataPrefix))
	}
}

func executeStatement(db *sql.DB, ctx context.Context, statement string) {
	log.Printf("Running statement: %s", statement)
	result, err := db.ExecContext(ctx, statement)
	if err != nil {
		log.Fatal(err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Statement affected %d rows.", rows)
}
