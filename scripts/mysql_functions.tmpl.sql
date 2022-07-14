DELIMITER $$

DROP PROCEDURE IF EXISTS create_orgs$$
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
END $$

DROP PROCEDURE IF EXISTS create_shared_domains$$
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
END $$

DROP PROCEDURE IF EXISTS create_private_domains$$
CREATE PROCEDURE create_private_domains(num_private_domains INT)
BEGIN
    DECLARE org_id INT;
    DECLARE num_created_private_domains INT;
    DECLARE private_domain_guid TEXT;
    DECLARE private_domain_name_prefix TEXT;
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
END $$

DROP PROCEDURE IF EXISTS assign_user_as_org_manager$$
CREATE PROCEDURE assign_user_as_org_manager(user_guid TEXT, num_orgs INT)
BEGIN
    DECLARE v_user_id INT;
    DECLARE num_assigned_orgs INT;
    DECLARE org_id INT;
    DECLARE org_name_query TEXT;
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
END $$

DELIMITER ;
