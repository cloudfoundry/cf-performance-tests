package helpers

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"runtime"
	"text/template"
	"time"

	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"

	_ "github.com/go-sql-driver/mysql"
)

func OpenDbConnections(testConfig Config) (ccdb, uaadb *sql.DB, ctx context.Context) {
	log.Printf("opening db connection to %s", testConfig.DatabaseType)
	driverName := ""
	switch testConfig.DatabaseType {
	case psql_db:
		driverName = "pgx"
	case mysql_db:
		driverName = "mysql"
	default:
		log.Fatalf("Invalid 'database_type' parameter: %s", testConfig.DatabaseType)
	}

	ccdb, err := sql.Open(driverName, testConfig.CcdbConnection)
	checkError(err)

	if testConfig.UaadbConnection != "" {
		uaadb, err = sql.Open(driverName, testConfig.UaadbConnection)
		checkError(err)
	}

	ctx = context.Background()

	return
}

func ImportStoredProcedures(ccdb *sql.DB, ctx context.Context, testConfig Config) {
	if testConfig.DatabaseType == mysql_db {
		InitializeMySql(ccdb, ctx, testConfig)
		return
	}

	type StoredProceduresSQLTemplate struct {
		Prefix string
	}
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Failed to retrieve current file location")
	}

	sqlFunctionsTemplate, err := ioutil.ReadFile(path.Join(path.Dir(filename), "../scripts/pgsql_functions.tmpl.sql"))
	if err != nil {
		log.Fatal(err)
	}

	tmpl, err := template.New("sql_functions").Parse(string(sqlFunctionsTemplate))
	if err != nil {
		log.Fatal(err)
	}

	sqlFunctionsTemplateResult := new(bytes.Buffer)
	err = tmpl.Execute(sqlFunctionsTemplateResult, StoredProceduresSQLTemplate{testConfig.GetNamePrefix()})
	if err != nil {
		log.Fatal(err)
	}

	ExecuteStatement(ccdb, ctx, sqlFunctionsTemplateResult.String())
}

func CleanupTestData(ccdb, uaadb *sql.DB, ctx context.Context, testConfig Config) {
	deleteStatements := []string{
		"DELETE FROM route_mappings USING routes WHERE routes.guid = route_mappings.route_guid AND routes.host LIKE '%s'",
		"DELETE FROM routes WHERE host LIKE '%s'",
		"DELETE FROM domain_annotations USING domains WHERE domain_annotations.resource_guid = domains.guid AND domains.name LIKE '%s'",
		"DELETE FROM domains WHERE name LIKE '%s'",
		"DELETE FROM service_bindings USING apps WHERE apps.guid = service_bindings.app_guid AND apps.name LIKE '%s'",
		"DELETE FROM route_mappings USING apps WHERE apps.guid = route_mappings.app_guid AND apps.name LIKE '%s'",
		"DELETE FROM apps WHERE name LIKE '%s'",
		"DELETE FROM service_keys WHERE name LIKE '%s'",
		"DELETE FROM service_bindings USING service_instances WHERE service_instances.guid = service_bindings.service_instance_guid AND service_instances.name LIKE '%s'",
		"DELETE FROM service_instances WHERE name LIKE '%s'",
		"DELETE FROM security_groups_spaces USING security_groups WHERE security_groups_spaces.security_group_id = security_groups.id AND security_groups.name LIKE '%s'",
		"DELETE FROM security_groups_spaces USING spaces WHERE security_groups_spaces.space_id = spaces.id AND spaces.name LIKE '%s'",
		"DELETE FROM security_groups WHERE name LIKE '%s'",
		"DELETE FROM spaces_developers USING spaces WHERE spaces_developers.space_id = spaces.id AND spaces.name LIKE '%s'",
		"DELETE FROM spaces_managers USING spaces WHERE spaces_managers.space_id = spaces.id AND spaces.name LIKE '%s'",
		"DELETE FROM spaces_auditors USING spaces WHERE spaces_auditors.space_id = spaces.id AND spaces.name LIKE '%s'",
		"DELETE FROM space_labels USING spaces WHERE space_labels.resource_guid = spaces.guid AND spaces.name LIKE '%s'",
		"DELETE FROM spaces WHERE name LIKE '%s'",
		"DELETE FROM service_plan_visibilities USING organizations WHERE service_plan_visibilities.organization_id = organizations.id AND organizations.name LIKE '%s'",
		"DELETE FROM organizations_users USING organizations WHERE organizations_users.organization_id = organizations.id AND organizations.name LIKE '%s'",
		"DELETE FROM organizations_managers USING organizations WHERE organizations_managers.organization_id = organizations.id AND organizations.name LIKE '%s'",
		"DELETE FROM organizations_isolation_segments USING organizations WHERE organizations_isolation_segments.organization_guid = organizations.guid AND organizations.name LIKE '%s'",
		"DELETE FROM organizations WHERE name LIKE '%s'",
		"DELETE FROM quota_definitions WHERE name LIKE '%s'",
		"DELETE FROM isolation_segment_annotations USING isolation_segments WHERE isolation_segment_annotations.resource_guid = isolation_segments.guid AND isolation_segments.name LIKE '%s'",
		"DELETE FROM isolation_segments WHERE name LIKE '%s'",
		"DELETE FROM quota_definitions WHERE name LIKE '%s'",
		"DELETE FROM events WHERE actee_name LIKE '%s'",
		"DELETE FROM service_plan_visibilities USING service_plans WHERE service_plans.id = service_plan_visibilities.service_plan_id AND service_plans.name LIKE '%s'",
		"DELETE FROM service_plans WHERE name LIKE '%s'",
		"DELETE FROM services WHERE label LIKE '%s'",
		"DELETE FROM service_brokers WHERE name LIKE '%s'",
	}
	nameQuery := fmt.Sprintf("%s-%%", testConfig.GetNamePrefix())

	for _, statement := range deleteStatements {
		ExecuteStatement(ccdb, ctx, fmt.Sprintf(statement, nameQuery))
	}
	fmt.Printf("%v Running 'VACUUM FULL' on db...\n", time.Now().Format(time.RFC850))
	ExecuteStatement(ccdb, ctx, "VACUUM FULL;")

	if uaadb != nil {
		userGuids := ExecuteSelectStatement(uaadb, ctx, fmt.Sprintf("SELECT id FROM users WHERE username LIKE '%s'", nameQuery))

		for _, userGuid := range userGuids {
			ExecuteStatement(ccdb, ctx, fmt.Sprintf("DELETE FROM users WHERE guid = '%s'", userGuid))
		}

		ExecuteStatement(uaadb, ctx, fmt.Sprintf("DELETE FROM users WHERE username LIKE '%s'", nameQuery))
	}
}

func AnalyzeDB(ccdb *sql.DB, ctx context.Context) {
	fmt.Printf("%v Running 'ANALYZE' on db...\n", time.Now().Format(time.RFC850))
	ExecuteStatement(ccdb, ctx, "ANALYZE;")
}

func ExecuteStoredProcedure(testConfig Config, db *sql.DB, ctx context.Context, statement string) {
	sqlCmd := ""
	switch testConfig.DatabaseType {
	case psql_db:
		sqlCmd = "SELECT FROM "
	case mysql_db:
		sqlCmd = "CALL "
	default:
		log.Fatalf("Invalid 'database_type' parameter: %s", testConfig.DatabaseType)
	}
	ExecuteStatement(db, ctx, sqlCmd+statement)
}

func ExecuteStatement(db *sql.DB, ctx context.Context, statement string) {
	result, err := db.ExecContext(ctx, statement)
	checkError(err)
	_, err = result.RowsAffected()
	checkError(err)
}

func ExecutePreparedInsertStatement(db *sql.DB, ctx context.Context, statement string, args ...interface{}) int {
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

func ExecuteSelectStatement(db *sql.DB, ctx context.Context, statement string) []interface{} {
	rows, err := db.QueryContext(ctx, statement)
	checkError(err)
	defer rows.Close()
	results := make([]interface{}, 0)

	for rows.Next() {
		var result interface{}
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
