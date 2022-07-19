CREATE PROCEDURE create_orgs(num_orgs INT)
BEGIN
    DECLARE org_guid VARCHAR(255);
    DECLARE default_quota_definition_id INT;
    DECLARE counter INT;
    SET default_quota_definition_id = 1;
    SET counter = 0;
    WHILE counter < num_orgs
        DO
            SET counter = counter + 1;
            SET org_guid = uuid();
            INSERT INTO organizations (guid, name, quota_definition_id)
            VALUES (org_guid, CONCAT('{{.Prefix}}-org-', org_guid), default_quota_definition_id);
        END WHILE;
END;