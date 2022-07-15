package helpers

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"log"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
)

var StoredProcedures = [][]string{
	{
		"create_orgs", `
CREATE PROCEDURE create_orgs (num_orgs INT)
BEGIN
    DECLARE org_guid VARCHAR(255);
    DECLARE org_name_prefix VARCHAR(255);
    DECLARE default_quota_definition_id INT;
    DECLARE counter INT;
    SET org_name_prefix = '{{.Prefix}}-org-';
    SET default_quota_definition_id = 1;
    SET counter = 0;
    WHILE counter < num_orgs DO
        SET counter = counter + 1;
        SET org_guid = uuid();
        INSERT INTO organizations (guid, name, quota_definition_id)
            VALUES (org_guid, CONCAT(org_name_prefix, org_guid), default_quota_definition_id);
    END WHILE;
END;
`},
	{
		"create_shared_domains", `
CREATE PROCEDURE create_shared_domains(num_shared_domains INT)
BEGIN
    DECLARE shared_domain_guid VARCHAR(255);
    DECLARE shared_domain_name_prefix VARCHAR(255);
    DECLARE counter INT;
    SET shared_domain_name_prefix = '{{.Prefix}}-shared-domain-';
    SET counter = 0;
    WHILE counter < num_shared_domains DO
        SET counter = counter + 1;
        SET shared_domain_guid = uuid();
        INSERT INTO domains (guid, name)
           VALUES (shared_domain_guid, CONCAT(shared_domain_name_prefix, shared_domain_guid));
    END WHILE;
END;
`,
	},
}

func InitializeMySql(ccdb *sql.DB, ctx context.Context, testConfig Config) {
	type StoredProceduresSQLTemplate struct {
		Prefix string
	}

	for _, storedProcedure := range StoredProcedures {
		log.Printf("Initialising stored procedure %s ...", storedProcedure[0])

		tmpl, err := template.New("sql_functions").Parse(storedProcedure[1])
		if err != nil {
			log.Fatal(err)
		}

		sqlFunctionsTemplateResult := new(bytes.Buffer)
		err = tmpl.Execute(sqlFunctionsTemplateResult, StoredProceduresSQLTemplate{testConfig.GetNamePrefix()})
		if err != nil {
			log.Fatal(err)
		}

		ExecuteStatement(ccdb, ctx, fmt.Sprintf("DROP PROCEDURE IF EXISTS %s;", storedProcedure[0]))
		ExecuteStatement(ccdb, ctx, sqlFunctionsTemplateResult.String())
	}
}
