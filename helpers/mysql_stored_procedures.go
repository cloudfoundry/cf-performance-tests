package helpers

// we can't use a single .sql file for storing all MySQL procedures because the DELIMITER keyword cannot be used
// (DELIMITER is a feature of the "mysql" client, but is not supported in other APIs)

var StoredProceduresMySql = [][]string{
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
	{
		"create_private_domains", `
CREATE PROCEDURE create_private_domains(num_private_domains INT)
BEGIN
    DECLARE org_id INT;
    DECLARE num_created_private_domains INT;
    DECLARE private_domain_guid VARCHAR(255);
    DECLARE private_domain_name_prefix VARCHAR(255);
    DECLARE orgs_cursor CURSOR FOR SELECT id FROM organizations WHERE name LIKE '{{.Prefix}}-org-%' ORDER BY RAND();
    -- when we've iterated over all orgs, re-open the cursor so that we get a new batch of random org ids
    DECLARE CONTINUE HANDLER FOR NOT FOUND
    BEGIN
        CLOSE orgs_cursor;
        OPEN orgs_cursor;
    END;
    SET num_created_private_domains = 0;
    SET private_domain_name_prefix = '{{.Prefix}}-private-domain-';

    OPEN orgs_cursor;
    org_loop: LOOP
        SET num_created_private_domains = num_created_private_domains + 1;
        IF num_created_private_domains > num_private_domains THEN
            LEAVE org_loop;
        END IF;
        FETCH orgs_cursor INTO org_id;
        SET private_domain_guid = uuid();
        INSERT INTO domains (guid, name, owning_organization_id)
            VALUES (private_domain_guid, CONCAT(private_domain_name_prefix, private_domain_guid), org_id);
    END LOOP;
    CLOSE orgs_cursor;
END;
`,
	},
	{
		"assign_user_as_org_manager", `
CREATE PROCEDURE assign_user_as_org_manager(user_guid TEXT, num_orgs INT)
BEGIN
    DECLARE v_user_id INT;
    DECLARE num_assigned_orgs INT;
    DECLARE org_id INT;
    DECLARE org_name_query VARCHAR(255);
    DECLARE orgs_cursor CURSOR FOR SELECT id FROM organizations WHERE name LIKE org_name_query ORDER BY RAND() LIMIT num_orgs;
    SET num_assigned_orgs = 0;
    SET org_name_query = '{{.Prefix}}-org-%';

    SELECT id FROM users WHERE guid = user_guid INTO v_user_id;
    OPEN orgs_cursor;
    org_loop: LOOP
        SET num_assigned_orgs = num_assigned_orgs + 1;
        IF num_assigned_orgs > num_orgs THEN
            LEAVE org_loop;
        END IF;
        FETCH orgs_cursor INTO org_id;
        INSERT INTO organizations_managers (organization_id, user_id) VALUES (org_id, v_user_id);
    END LOOP;
    CLOSE orgs_cursor;
END;
`,
	},
	{
		"create_isolation_segments", `
CREATE PROCEDURE create_isolation_segments(num_isolation_segments INT)
BEGIN
    DECLARE isolation_segment_guid VARCHAR(255);
	DECLARE	isolation_segment_name_prefix VARCHAR(255);
    DECLARE counter INT;
    SET isolation_segment_name_prefix = '{{.Prefix}}-isolation-segment-';
    SET counter = 0;

    WHILE counter < num_isolation_segments DO
        SET counter = counter + 1;
        SET isolation_segment_guid = uuid();
        INSERT INTO isolation_segments (guid, name)
            VALUES (isolation_segment_guid, CONCAT(isolation_segment_name_prefix, isolation_segment_guid));
    END WHILE;
END;
`}, {
		"assign_orgs_to_isolation_segments", `
CREATE PROCEDURE assign_orgs_to_isolation_segments(num_orgs INT)
BEGIN
    DECLARE org_guid VARCHAR(255);
    DECLARE org_name_query VARCHAR(255);
    DECLARE isolation_segment_name_query VARCHAR(255);
    DECLARE v_isolation_segment_guid VARCHAR(255);
    DECLARE orgs_cursor CURSOR FOR SELECT guid FROM organizations WHERE name LIKE org_name_query ORDER BY RAND() LIMIT num_orgs;
    DECLARE finished INT;
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET finished = 1;
    SET org_name_query = '{{.Prefix}}-org-%';
    SET isolation_segment_name_query = '{{.Prefix}}-isolation-segment-%';
    SET finished = 0;

    OPEN orgs_cursor;
    org_loop: LOOP
        FETCH orgs_cursor INTO org_guid;
        IF finished = 1 THEN 
            LEAVE org_loop;
        SELECT guid FROM isolation_segments WHERE name LIKE isolation_segment_name_query ORDER BY RAND() LIMIT 1 INTO v_isolation_segment_guid;
        INSERT INTO organizations_isolation_segments (organization_guid, isolation_segment_guid)
            VALUES (org_guid, v_isolation_segment_guid);
    END LOOP;
    CLOSE orgs_cursor;
END;
`},
}
