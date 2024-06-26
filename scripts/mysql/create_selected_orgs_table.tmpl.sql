CREATE PROCEDURE create_selected_orgs_table(
    num_orgs INT
)
BEGIN
    DROP TABLE IF EXISTS selected_orgs;

    CREATE TABLE selected_orgs(id INT NOT NULL PRIMARY KEY);

    INSERT INTO selected_orgs
    SELECT id FROM organizations
    WHERE name LIKE '{{.Prefix}}-org-%'
    ORDER BY RAND()
    LIMIT num_orgs;
END;