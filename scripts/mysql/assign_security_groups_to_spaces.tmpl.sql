CREATE PROCEDURE assign_security_groups_to_spaces(num_spaces INT, num_security_groups_per_space INT)
BEGIN
    DECLARE v_space_id INT;
    DECLARE v_security_group_id INT;
    DECLARE spaces_finished, security_groups_finished BOOLEAN DEFAULT FALSE;
    DECLARE spaces_cursor CURSOR FOR SELECT id
                                     FROM spaces
                                     WHERE name LIKE '{{.Prefix}}-space-%'
                                     ORDER BY RAND()
                                     LIMIT num_spaces;
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET spaces_finished = TRUE;

    OPEN spaces_cursor;
    spaces_loop:
    LOOP
        FETCH FROM spaces_cursor INTO v_space_id;
        IF spaces_finished THEN
            LEAVE spaces_loop;
        END IF;

        innerblock:
        BEGIN
            DECLARE security_groups_cursor CURSOR FOR SELECT id
                                                      FROM security_groups
                                                      WHERE name LIKE '{{.Prefix}}-security-group-%'
                                                      ORDER BY RAND()
                                                      LIMIT num_security_groups_per_space;
            DECLARE CONTINUE HANDLER FOR NOT FOUND SET security_groups_finished = TRUE;

            SET security_groups_finished = FALSE;
            OPEN security_groups_cursor;
            security_groups_loop:
            LOOP
                FETCH FROM security_groups_cursor INTO v_security_group_id;
                IF security_groups_finished THEN
                    LEAVE security_groups_loop;
                END IF;
                INSERT INTO security_groups_spaces (security_group_id, space_id)
                VALUES (v_security_group_id, v_space_id);
            END LOOP security_groups_loop;
            CLOSE security_groups_cursor;
        END innerblock;
    END LOOP spaces_loop;
    CLOSE spaces_cursor;
END;