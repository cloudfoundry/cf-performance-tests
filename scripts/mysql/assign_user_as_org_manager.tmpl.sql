CREATE PROCEDURE assign_user_as_org_manager(user_guid TEXT, num_orgs INT)
BEGIN
    DECLARE v_user_id INT;
    DECLARE org_id INT;
    DECLARE finished BOOLEAN DEFAULT FALSE;
    DECLARE orgs_cursor CURSOR FOR SELECT id
                                   FROM organizations
                                   WHERE name LIKE '{{.Prefix}}-org-%'
                                   ORDER BY RAND()
                                   LIMIT num_orgs;
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET finished = TRUE;

    SELECT id FROM users WHERE guid = user_guid INTO v_user_id;
    OPEN orgs_cursor;
    org_loop:
    LOOP
        FETCH orgs_cursor INTO org_id;
        IF finished THEN
            LEAVE org_loop;
        END IF;
        INSERT INTO organizations_managers (organization_id, user_id) VALUES (org_id, v_user_id);
    END LOOP;
    CLOSE orgs_cursor;
END;