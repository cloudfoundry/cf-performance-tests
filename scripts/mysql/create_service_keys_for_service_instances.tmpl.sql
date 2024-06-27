CREATE PROCEDURE create_service_keys_for_service_instances(
    p_space_id INT,
    num_service_keys_per_service_instance INT)
BEGIN
    DECLARE _counter INT DEFAULT 0;

    WHILE _counter < num_service_keys_per_service_instance DO
        INSERT INTO service_keys (guid, name, credentials, service_instance_id)
        SELECT UUID(), CONCAT('{{.Prefix}}-service-key-', UUID()), '', id
        FROM service_instances WHERE name LIKE '{{.Prefix}}-service-instance-%'
        AND space_id = p_space_id;
        SET _counter = _counter + 1;
    END WHILE;
END;
