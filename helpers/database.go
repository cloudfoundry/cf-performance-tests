package helpers

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"log"
)

const TestDataPrefix = "perf-%"


func OpenDbConnections(ccdbConnection string, uaadbConnection string)(ccdb, uaadb *sql.DB, ctx context.Context){

	ccdb, err := sql.Open("pgx", ccdbConnection)
	checkError(err)

	uaadb, err = sql.Open("pgx", uaadbConnection)
	checkError(err)

	ctx = context.Background()

	return
}

func CleanupTestData(ccdb, uaadb *sql.DB, ctx context.Context) {
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
		"DELETE FROM quota_definitions WHERE name LIKE '%s'",
		"DELETE FROM events WHERE actee_name LIKE '%s'",
		"DELETE FROM service_plan_visibilities USING service_plans WHERE service_plans.id = service_plan_visibilities.service_plan_id AND service_plans.name LIKE '%s'",
		"DELETE FROM service_plans WHERE name LIKE '%s'",
		"DELETE FROM services WHERE label LIKE '%s'",
		"DELETE FROM service_brokers WHERE name LIKE '%s'",
	}
	for _, statement := range deleteStatements {
		ExecuteStatement(ccdb, ctx, fmt.Sprintf(statement, TestDataPrefix))
	}

	userGuids := ExecuteSelectStatement(uaadb, ctx, fmt.Sprintf("SELECT id FROM users WHERE username LIKE '%s'", TestDataPrefix))

	for _, userGuid := range userGuids {
		ExecuteStatement(ccdb, ctx, fmt.Sprintf("DELETE FROM users WHERE guid = '%s'", userGuid))
	}

	ExecuteStatement(uaadb, ctx, fmt.Sprintf("DELETE FROM users WHERE username LIKE '%s'", TestDataPrefix))

}

func ExecuteStatement(db *sql.DB, ctx context.Context, statement string) {
	result, err := db.ExecContext(ctx, statement)
	checkError(err)
	_, err = result.RowsAffected()
	checkError(err)
}

func ExecutePreparedInsertStatement(db *sql.DB, ctx context.Context, statement string, args ...interface{})int{
	var lastInsertId int
	stmt, err := db.PrepareContext(ctx, statement)
	checkError(err)
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, args...).Scan(&lastInsertId)
	checkError(err)
	return lastInsertId
}

func ExecuteInsertStatement(db *sql.DB, ctx context.Context, statement string) int {
	var lastInsertId int

	err := db.QueryRowContext(ctx, statement).Scan(&lastInsertId)
	checkError(err)
	return lastInsertId
}

func ExecuteSelectStatement(db *sql.DB, ctx context.Context, statement string) []string {
	rows, err := db.QueryContext(ctx, statement)
	checkError(err)
	defer rows.Close()
	results := make([]string, 0)

	for rows.Next() {
		var result string
		if err := rows.Scan(&result); err != nil {
			// Check for a scan error.
			// Query rows will be closed with defer.
			log.Fatal(err)
		}
		results = append(results, result)
	}

	// If the database is being written to ensure to check for Close
	// errors that may be returned from the driver. The query may
	// encounter an auto-commit error and be forced to rollback changes.
	rerr := rows.Close()
	checkError(rerr)

	// Rows.Err will report the last error encountered by Rows.Scan.
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return results
}

func ExecuteSelectStatementOneRow(db *sql.DB, ctx context.Context, statement string) int {
	var result int
	err := db.QueryRowContext(ctx, statement).Scan(&result)
	switch {
	case err == sql.ErrNoRows:
		log.Fatalf("query %s returned no value", statement)
	case err != nil:
		log.Fatalf("query %s failed with %v", statement, err)
	}

	return result
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
