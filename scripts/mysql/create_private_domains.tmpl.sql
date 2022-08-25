CREATE PROCEDURE create_private_domains(num_private_domains INT)
BEGIN
    DECLARE org_id INT;
    DECLARE num_created_private_domains INT;
    DECLARE private_domain_guid VARCHAR(255);
    DECLARE orgs_cursor CURSOR FOR SELECT id FROM organizations WHERE name LIKE '{{.Prefix}}-org-%' ORDER BY RAND();
    -- when we've iterated over all orgs, re-open the cursor so that we get a new batch of random org ids
    DECLARE CONTINUE HANDLER FOR NOT FOUND
        BEGIN
            CLOSE orgs_cursor;
            OPEN orgs_cursor;
        END;
    SET num_created_private_domains = 0;

    OPEN orgs_cursor;
    org_loop:
    LOOP
        SET num_created_private_domains = num_created_private_domains + 1;
        IF num_created_private_domains > num_private_domains THEN
            LEAVE org_loop;
        END IF;
        FETCH orgs_cursor INTO org_id;
        SET private_domain_guid = uuid();
        INSERT INTO domains (guid, name, owning_organization_id)
        VALUES (private_domain_guid, CONCAT('{{.Prefix}}-private-domain-', private_domain_guid), org_id);
    END LOOP;
    CLOSE orgs_cursor;
END;