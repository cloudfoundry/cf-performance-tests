package helpers

// we can't use a single .sql file for storing all MySQL procedures because the DELIMITER keyword cannot be used
// (DELIMITER is a feature of the "mysql" client, but is not supported in other APIs)

var StoredProceduresMySql = [][]string{
	{
		"create_orgs", `
CREATE PROCEDURE create_orgs(num_orgs INT)
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
		"create_spaces", `
CREATE PROCEDURE create_spaces(num_spaces_per_org INT)
BEGIN
    DECLARE org_id INT;
    DECLARE org_name_query VARCHAR(255);
    DECLARE space_name_prefix VARCHAR(255);
    DECLARE space_guid VARCHAR(255);
    DECLARE counter INT;
    DECLARE finished BOOLEAN DEFAULT FALSE;
    DECLARE orgs_cursor CURSOR FOR SELECT id FROM organizations WHERE name LIKE org_name_query;
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET finished = TRUE;
    SET org_name_query = '{{.Prefix}}-org-%';
    SET space_name_prefix = '{{.Prefix}}-space-';
    SET counter = 0;

    OPEN orgs_cursor;
    org_loop: LOOP
        FETCH orgs_cursor INTO org_id;
        IF finished = TRUE THEN
            LEAVE org_loop;
        END IF;
        SET counter = 0;
        WHILE counter < num_spaces_per_org DO
            SET counter = counter + 1;
            SET space_guid = uuid();
            INSERT INTO spaces (guid, name, organization_id)
                VALUES (space_guid, CONCAT(space_name_prefix, space_guid), org_id);
            INSERT INTO space_labels (guid, key_name, resource_guid)
                VALUES (space_guid, '{{.Prefix}}', space_guid);
        END WHILE;
    END LOOP;
    CLOSE orgs_cursor;
END;
`,
	},
	{
		"create_security_groups", `
CREATE PROCEDURE create_security_groups(security_groups INT)
BEGIN
    DECLARE security_group_guid VARCHAR(255);
    DECLARE security_group_name_prefix VARCHAR(255);
    DECLARE security_rule MEDIUMTEXT;
    DECLARE counter INT;
    SET security_group_name_prefix = '{{.Prefix}}-security-group-';
    SET security_rule = '[
        {
            "protocol": "icmp",
            "destination": "0.0.0.0/0",
            "type": 0,
            "code": 0
        },
        {
            "protocol": "tcp",
            "destination": "10.0.11.0/24",
            "ports": "80,443",
            "log": true,
            "description": "Allow http and https traffic to ZoneA"
        }
    ]';
    SET counter = 0;
    WHILE counter < security_groups DO
        SET counter = counter + 1;
        SET security_group_guid = uuid();
        INSERT INTO security_groups (guid, name, rules)
            VALUES (security_group_guid, CONCAT(security_group_name_prefix, security_group_guid), security_rule);
    END WHILE;
END;
`,
	},
	{
		"assign_security_groups_to_spaces", `
CREATE PROCEDURE assign_security_groups_to_spaces(num_spaces INT, num_security_groups_per_space INT)
BEGIN
    DECLARE v_space_id INT;
    DECLARE space_name_query VARCHAR(255);
    DECLARE v_security_group_id int;
    DECLARE security_group_name_query VARCHAR(255);
    DECLARE spaces_finished, security_groups_finished BOOLEAN DEFAULT FALSE;
    DECLARE spaces_cursor CURSOR FOR SELECT id FROM spaces WHERE name LIKE space_name_query ORDER BY RAND() LIMIT num_spaces;
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET spaces_finished = TRUE;

    SET space_name_query = '{{.Prefix}}-space-%';
    SET security_group_name_query = '{{.Prefix}}-security-group-%';

    OPEN spaces_cursor;
    spaces_loop: LOOP
        FETCH FROM spaces_cursor INTO v_space_id;
        IF spaces_finished = TRUE THEN
            LEAVE spaces_loop;
        END IF;

        innerblock: BEGIN
        DECLARE security_groups_cursor CURSOR FOR SELECT id FROM security_groups
            WHERE name LIKE security_group_name_query ORDER BY RAND() LIMIT num_security_groups_per_space;
        DECLARE CONTINUE HANDLER FOR NOT FOUND SET security_groups_finished = TRUE;

        OPEN security_groups_cursor;
        security_groups_loop: LOOP
            FETCH FROM security_groups_cursor INTO v_security_group_id;
            IF spaces_finished = TRUE THEN
                LEAVE security_groups_loop;
            END IF;
            INSERT INTO security_groups_spaces (security_group_id, space_id) VALUES (v_security_group_id, v_space_id);
        END LOOP;
        CLOSE security_groups_cursor;
        END innerblock;
    END LOOP;
    CLOSE spaces_cursor;
END;
`,
	},
	{
		"assign_user_as_space_developer", `
CREATE PROCEDURE assign_user_as_space_developer(user_guid VARCHAR(255), num_spaces INT)
BEGIN
    DECLARE v_user_id INT;
    DECLARE v_space_id INT;
    DECLARE space_name_query VARCHAR(255);
    DECLARE finished BOOLEAN DEFAULT FALSE;
    DECLARE spaces_cursor CURSOR FOR SELECT id FROM spaces WHERE name LIKE space_name_query ORDER BY RAND() LIMIT num_spaces;
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET finished = TRUE;
    SET space_name_query = '{{.Prefix}}-space-%';

    SELECT id FROM users WHERE guid = user_guid INTO v_user_id;
    OPEN spaces_cursor;
    spaces_loop: LOOP
        FETCH FROM spaces_cursor INTO v_space_id;
        IF finished = TRUE THEN
            LEAVE spaces_loop;
        END IF;
        INSERT INTO spaces_developers (space_id, user_id) VALUES (v_space_id, v_user_id);
    END LOOP;
    CLOSE spaces_cursor;
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
    DECLARE finished INT;
    DECLARE orgs_cursor CURSOR FOR SELECT guid FROM organizations WHERE name LIKE org_name_query ORDER BY RAND() LIMIT num_orgs;
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET finished = 1;
    SET org_name_query = '{{.Prefix}}-org-%';
    SET isolation_segment_name_query = '{{.Prefix}}-isolation-segment-%';
    SET finished = 0;

    OPEN orgs_cursor;
    org_loop: LOOP
        FETCH orgs_cursor INTO org_guid;
        IF finished = 1 THEN
            LEAVE org_loop;
        END IF;
        SELECT guid FROM isolation_segments WHERE name LIKE isolation_segment_name_query ORDER BY RAND() LIMIT 1 INTO v_isolation_segment_guid;
        INSERT INTO organizations_isolation_segments (organization_guid, isolation_segment_guid)
            VALUES (org_guid, v_isolation_segment_guid);
    END LOOP;
    CLOSE orgs_cursor;
END;
`},
}
