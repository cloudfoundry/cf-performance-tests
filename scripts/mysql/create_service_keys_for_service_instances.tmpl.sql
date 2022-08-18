CREATE PROCEDURE create_service_keys_for_service_instances(
    p_space_id INT,
    num_service_keys_per_service_instance INT)
BEGIN
    DECLARE v_service_instance_id INT;
    DECLARE service_key_guid VARCHAR(255);
    DECLARE service_keys_counter INT;
    DECLARE finished BOOLEAN DEFAULT FALSE;
    DECLARE service_instances_cursor CURSOR FOR SELECT id FROM service_instances WHERE space_id = p_space_id;
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET finished = TRUE;

    OPEN service_instances_cursor;
    service_instances_loop:
    LOOP
        FETCH service_instances_cursor INTO v_service_instance_id;
        IF finished THEN
            LEAVE service_instances_loop;
        END IF;
        SET service_keys_counter = 0;
        WHILE service_keys_counter < num_service_keys_per_service_instance
            DO
                SET service_keys_counter = service_keys_counter + 1;
                SET service_key_guid:= UUID();
                INSERT INTO service_keys (guid, name, credentials, service_instance_id)
                    VALUES (service_key_guid, CONCAT('{{.Prefix}}-service-key-', service_key_guid), '', v_service_instance_id);
            END WHILE;
    END LOOP;
    CLOSE service_instances_cursor;
END;
