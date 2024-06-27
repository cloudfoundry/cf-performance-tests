CREATE PROCEDURE assign_user_as_org_role(
    IN user_guid VARCHAR(255),
    IN org_role VARCHAR(255),
    IN num_orgs INT
)
BEGIN
    DECLARE v_user_id INT;
    DECLARE v_org_id INT;
    DECLARE finished BOOLEAN DEFAULT FALSE;
    DECLARE orgs_cursor CURSOR FOR SELECT id
                                   FROM selected_orgs
                                   ORDER BY RAND()
                                   LIMIT num_orgs;
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET finished = TRUE;

    SET @insert_sql := CONCAT('INSERT INTO ', org_role, ' (organization_id, user_id) VALUES (?, ?)');
    PREPARE insert_statement FROM @insert_sql;

    SELECT id FROM users WHERE guid = user_guid INTO v_user_id;
    OPEN orgs_cursor;
    org_loop:
    LOOP
        FETCH orgs_cursor INTO v_org_id;
        IF finished THEN
            LEAVE org_loop;
        END IF;

        SET @org_id = v_org_id;
        SET @user_id = v_user_id;
        EXECUTE insert_statement USING @org_id, @user_id;
    END LOOP;
    CLOSE orgs_cursor;

    DEALLOCATE PREPARE insert_statement;
END;