CREATE PROCEDURE create_users_with_org_and_space_roles(org_guid VARCHAR(255), space_guid VARCHAR(255), num_users INT)
BEGIN
    DECLARE default_quota_definition_id INT DEFAULT 1;
    DECLARE org_id INT;
    DECLARE space_id INT;
    DECLARE counter INT;
    DECLARE user_guid VARCHAR(255);
    DECLARE active BOOLEAN DEFAULT TRUE;
    DECLARE user_id INT;

    INSERT INTO organizations (guid, name, quota_definition_id) VALUES (org_guid, CONCAT('{{.Prefix}}-org-', org_guid), default_quota_definition_id);
    SET org_id = LAST_INSERT_ID();
    INSERT INTO spaces (guid, name, organization_id) VALUES (space_guid, CONCAT('{{.Prefix}}-space-', space_guid), org_id);
    SET space_id = LAST_INSERT_ID();

    SET counter = 0;
    WHILE counter < num_users
        DO
            SET counter = counter + 1;
            SET user_guid = uuid();
            INSERT INTO users (guid, default_space_id, active) VALUES (user_guid, space_id, active);
            SET user_id = LAST_INSERT_ID();

            INSERT INTO organizations_managers (organization_id, user_id) VALUES (org_id, user_id);
            INSERT INTO organizations_billing_managers (organization_id, user_id) VALUES (org_id, user_id);
            INSERT INTO organizations_auditors (organization_id, user_id) VALUES (org_id, user_id);
            INSERT INTO organizations_users (organization_id, user_id) VALUES (org_id, user_id);

            INSERT INTO spaces_managers (space_id, user_id) VALUES (space_id, user_id);
            INSERT INTO spaces_developers (space_id, user_id) VALUES (space_id, user_id);
            INSERT INTO spaces_supporters (space_id, user_id) VALUES (space_id, user_id);
            INSERT INTO spaces_auditors (space_id, user_id) VALUES (space_id, user_id);
        END WHILE;
END;