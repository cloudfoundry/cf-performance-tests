CREATE PROCEDURE create_service_instances(
    space_id INT,
    service_plan_id INT,
    num_service_instances INT)
BEGIN
    DECLARE service_instance_guid VARCHAR(255);
    DECLARE service_instances_counter INT DEFAULT 0;

    WHILE service_instances_counter < num_service_instances
        DO
            SET service_instances_counter = service_instances_counter + 1;
            SET service_instance_guid = UUID();
            INSERT INTO service_instances (guid, name, space_id, service_plan_id)
            VALUES (service_instance_guid, CONCAT('{{.Prefix}}-service-instance-', service_instance_guid), space_id,
                    service_plan_id);
        END WHILE;
END;
