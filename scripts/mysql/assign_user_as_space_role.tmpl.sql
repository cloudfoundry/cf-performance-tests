CREATE PROCEDURE assign_user_as_space_role(
    user_guid VARCHAR(255),
    space_role VARCHAR(255),
    num_spaces INT)
BEGIN
    DECLARE v_user_id INT;
    DECLARE v_space_id INT;
    DECLARE finished BOOLEAN DEFAULT FALSE;
    DECLARE spaces_cursor CURSOR FOR SELECT spaces.id
                                     FROM spaces
                                     JOIN selected_orgs
                                     ON spaces.organization_id = selected_orgs.id
                                     WHERE name LIKE '{{.Prefix}}-space-%'
                                     ORDER BY RAND()
                                     LIMIT num_spaces;
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET finished = TRUE;

    SET @insert_sql := CONCAT('INSERT INTO ', space_role, ' (space_id, user_id) VALUES (?, ?)');
    PREPARE insert_statement FROM @insert_sql;

    SELECT id FROM users WHERE guid = user_guid INTO v_user_id;
    OPEN spaces_cursor;
    spaces_loop:
    LOOP
        FETCH spaces_cursor INTO v_space_id;
        IF finished THEN
            LEAVE spaces_loop;
        END IF;

        SET @space_id = v_space_id;
        SET @user_id = v_user_id;
        EXECUTE insert_statement USING @space_id, @user_id;
    END LOOP;
    CLOSE spaces_cursor;

    DEALLOCATE PREPARE insert_statement;
END;