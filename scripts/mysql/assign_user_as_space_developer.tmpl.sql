CREATE PROCEDURE assign_user_as_space_developer(user_guid VARCHAR(255), num_spaces INT)
BEGIN
    DECLARE v_user_id INT;
    DECLARE v_space_id INT;
    DECLARE finished BOOLEAN DEFAULT FALSE;
    DECLARE spaces_cursor CURSOR FOR SELECT id
                                     FROM spaces
                                     WHERE name LIKE '{{.Prefix}}-space-%'
                                     ORDER BY RAND()
                                     LIMIT num_spaces;
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET finished = TRUE;

    SELECT id FROM users WHERE guid = user_guid INTO v_user_id;
    OPEN spaces_cursor;
    spaces_loop:
    LOOP
        FETCH FROM spaces_cursor INTO v_space_id;
        IF finished THEN
            LEAVE spaces_loop;
        END IF;
        INSERT INTO spaces_developers (space_id, user_id) VALUES (v_space_id, v_user_id);
    END LOOP;
    CLOSE spaces_cursor;
END;