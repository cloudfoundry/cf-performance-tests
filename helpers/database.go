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
	"strings"
	"text/template"
	"time"

	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"

	_ "github.com/go-sql-driver/mysql"
)

func OpenDbConnections(testConfig Config) (ccdb, uaadb *sql.DB, ctx context.Context) {
	log.Printf("Opening database connection to %s...", testConfig.DatabaseType)
	driverName := ""
	switch testConfig.DatabaseType {
	case PsqlDb:
		driverName = "pgx"
	case MysqlDb:
		driverName = "mysql"
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

func evaluateTemplate(templ string, testConfig Config) string {
	type StoredProceduresSQLTemplate struct {
		Prefix string
	}

	tmpl, err := template.New("sql_functions").Parse(templ)
	if err != nil {
		log.Fatal(err)
	}

	templateResult := new(bytes.Buffer)
	err = tmpl.Execute(templateResult, StoredProceduresSQLTemplate{testConfig.GetNamePrefix()})
	if err != nil {
		log.Fatal(err)
	}

	return templateResult.String()
}

func ImportStoredProcedures(ccdb *sql.DB, ctx context.Context, testConfig Config) {
	_, currentDir, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Failed to retrieve current file location")
	}

	if testConfig.DatabaseType == PsqlDb {
		sqlFunctionsTemplate, err := ioutil.ReadFile(path.Join(path.Dir(currentDir), "../scripts/pgsql_functions.tmpl.sql"))
		if err != nil {
			log.Fatal(err)
		}

		ExecuteStatement(ccdb, ctx, evaluateTemplate(string(sqlFunctionsTemplate), testConfig))
	}

	if testConfig.DatabaseType == MysqlDb {
		mysqlDir := path.Join(path.Dir(currentDir), "../scripts/mysql/")
		mysqlDirFiles, err := ioutil.ReadDir(mysqlDir)
		if err != nil {
			log.Fatal(err)
		}

		for _, mysqlFile := range mysqlDirFiles {
			log.Printf("Reading MySQL stored procedure from file '%s'...", mysqlFile.Name())
			sqlTemplate, err := ioutil.ReadFile(path.Join(mysqlDir, mysqlFile.Name()))
			if err != nil {
				log.Fatal(err)
			}
			procedureName := strings.Split(mysqlFile.Name(), ".")[0]
			ExecuteStatement(ccdb, ctx, fmt.Sprintf("DROP FUNCTION IF EXISTS %s", procedureName))
			ExecuteStatement(ccdb, ctx, fmt.Sprintf("DROP PROCEDURE IF EXISTS %s", procedureName))
			ExecuteStatement(ccdb, ctx, evaluateTemplate(string(sqlTemplate), testConfig))
		}
	}
}

// define "random()" function for MySQL to enable re-use of PostgreSQL statements
func DefineRandomFunction(ccdb *sql.DB, ctx context.Context) {
	ExecuteStatement(ccdb, ctx, "DROP FUNCTION IF EXISTS random")
	ExecuteStatement(ccdb, ctx, "CREATE FUNCTION random() RETURNS FLOAT RETURN RAND()")
}

func CleanupTestData(ccdb, uaadb *sql.DB, ctx context.Context, testConfig Config) {
	deleteStatementsPostgres := []string{
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
	deleteStatementsMySql := []string{
		"DELETE FROM d_a USING domain_annotations d_a, domains d WHERE d_a.resource_guid = d.guid AND d.name LIKE '%s'",
		"DELETE FROM domains WHERE name LIKE '%s'",
		"DELETE FROM s_b USING service_bindings s_b, service_instances s_i WHERE s_i.guid = s_b.service_instance_guid AND s_i.name LIKE '%s'",
		"DELETE FROM service_instances WHERE name LIKE '%s'",
		"DELETE FROM s_g_s USING security_groups_spaces s_g_s, security_groups s_g WHERE s_g_s.security_group_id = s_g.id AND s_g.name LIKE '%s'",
		"DELETE FROM s_g_s USING security_groups_spaces s_g_s, spaces s WHERE s_g_s.space_id = s.id AND s.name LIKE '%s'",
		"DELETE FROM security_groups WHERE name LIKE '%s'",
		"DELETE FROM s_d USING spaces_developers s_d, spaces s WHERE s_d.space_id = s.id AND s.name LIKE '%s'",
		"DELETE FROM s_m USING spaces_managers s_m, spaces s WHERE s_m.space_id = s.id AND s.name LIKE '%s'",
		"DELETE FROM s_a USING spaces_auditors s_a, spaces s WHERE s_a.space_id = s.id AND s.name LIKE '%s'",
		"DELETE FROM s_l USING space_labels s_l, spaces s WHERE s_l.resource_guid = s.guid AND s.name LIKE '%s'",
		"DELETE FROM spaces WHERE name LIKE '%s'",
		"DELETE FROM s_p_v USING service_plan_visibilities s_p_v, organizations o WHERE s_p_v.organization_id = o.id AND o.name LIKE '%s'",
		"DELETE FROM o_u USING organizations_users o_u, organizations o WHERE o_u.organization_id = o.id AND o.name LIKE '%s'",
		"DELETE FROM o_m USING organizations_managers o_m, organizations o WHERE o_m.organization_id = o.id AND o.name LIKE '%s'",
		"DELETE FROM o_i_s USING organizations_isolation_segments o_i_s, organizations o WHERE o_i_s.organization_guid = o.guid AND o.name LIKE '%s'",
		"DELETE FROM organizations WHERE name LIKE '%s'",
		"DELETE FROM i_s_a USING isolation_segment_annotations i_s_a, isolation_segments i_s WHERE i_s_a.resource_guid = i_s.guid AND i_s.name LIKE '%s'",
		"DELETE FROM isolation_segments WHERE name LIKE '%s'",
		"DELETE FROM s_p_v USING service_plan_visibilities s_p_v, service_plans s_p WHERE s_p.id = s_p_v.service_plan_id AND s_p.name LIKE '%s'",
		"DELETE FROM service_plans WHERE name LIKE '%s'",
		"DELETE FROM services WHERE label LIKE '%s'",
		"DELETE FROM service_brokers WHERE name LIKE '%s'",
	}
	nameQuery := fmt.Sprintf("%s-%%", testConfig.GetNamePrefix())

	if testConfig.DatabaseType == PsqlDb {
		for _, statement := range deleteStatementsPostgres {
			ExecuteStatement(ccdb, ctx, fmt.Sprintf(statement, nameQuery))
		}

		log.Printf("%v Running 'VACUUM FULL' on db...\n", time.Now().Format(time.RFC850))
		ExecuteStatement(ccdb, ctx, "VACUUM FULL;")
	}

	if testConfig.DatabaseType == MysqlDb {
		for _, statement := range deleteStatementsMySql {
			ExecuteStatement(ccdb, ctx, fmt.Sprintf(statement, nameQuery))
		}
	}

	if uaadb != nil {
		userGuids := ExecuteSelectStatement(uaadb, ctx, fmt.Sprintf("SELECT id FROM users WHERE username LIKE '%s'", nameQuery))

		for _, userGuid := range userGuids {
			ExecuteStatement(ccdb, ctx, fmt.Sprintf("DELETE FROM users WHERE guid = '%s'", userGuid))
		}

		ExecuteStatement(uaadb, ctx, fmt.Sprintf("DELETE FROM users WHERE username LIKE '%s'", nameQuery))
	}
}

func AnalyzeDB(ccdb *sql.DB, ctx context.Context, testConfig Config) {
	if testConfig.DatabaseType == PsqlDb {
		log.Printf("%v Running 'ANALYZE' on db...\n", time.Now().Format(time.RFC850))
		ExecuteStatement(ccdb, ctx, "ANALYZE;")
	}
	if testConfig.DatabaseType == PsqlDb {
		log.Printf("Skipping 'ANALYZE' for MySQL.")
	}
}

func ExecuteStoredProcedure(db *sql.DB, ctx context.Context, statement string, testConfig Config) {
	sqlCmd := ""
	switch testConfig.DatabaseType {
	case PsqlDb:
		sqlCmd = "SELECT FROM "
	case MysqlDb:
		sqlCmd = "CALL "
	}
	log.Printf("Executing stored procedure: %s", sqlCmd+statement)
	ExecuteStatement(db, ctx, sqlCmd+statement)
	log.Printf("Finished stored procedure: %s", sqlCmd+statement)
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

func ConvertToString(input interface{}) string {
	if result, ok := input.(string); ok {
		return result
	}
	if result, ok := input.([]uint8); ok {
		return string(result)
	}
	log.Fatalf("Cannot convert input '%v' to string (type is '%T')", input, input)
	return ""
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
