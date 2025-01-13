CREATE PROCEDURE create_services_and_plans(num_services INT,
                                           service_broker_id INT,
                                           num_service_plans INT,
                                           service_plan_public BOOLEAN,
                                           num_visible_orgs INT)
BEGIN
    DECLARE service_guid VARCHAR(255);
    DECLARE service_bindable BOOLEAN DEFAULT TRUE;
    DECLARE service_plan_guid VARCHAR(255);
    DECLARE services_counter INT DEFAULT 0;
    DECLARE service_plan_free BOOLEAN DEFAULT TRUE;
    DECLARE service_plans_counter INT DEFAULT 0;
    DECLARE latest_service_id INT;
    DECLARE latest_service_plan_id INT;

    WHILE services_counter < num_services
        DO
            SET services_counter = services_counter + 1;
            SET service_guid = UUID();
            INSERT INTO services (guid, label, description, bindable, service_broker_id, extra)
            VALUES (service_guid,
                    CONCAT('{{.Prefix}}-service-', service_guid),
                    CONCAT('{{.Prefix}}-service-description-', service_guid),
                    service_bindable,
                    service_broker_id,
                    '{"shareable": true}');
            SET latest_service_id = LAST_INSERT_ID();
            SET service_plans_counter = 0;
            WHILE service_plans_counter < num_service_plans
                DO
                    SET service_plans_counter = service_plans_counter + 1;
                    SET service_plan_guid := UUID();
                    INSERT INTO service_plans (guid, name, description, free, service_id, unique_id, public, extra)
                    VALUES (service_plan_guid,
                            CONCAT('{{.Prefix}}-service-plan-', service_plan_guid),
                            CONCAT('{{.Prefix}}-service-plan-description-', service_plan_guid),
                            service_plan_free,
                            latest_service_id,
                            CONCAT('unique-', service_plan_guid),
                            service_plan_public,
                            '{"shareable": true}');
                    SET latest_service_plan_id = LAST_INSERT_ID();
                    INSERT INTO service_plan_visibilities (guid, service_plan_id, organization_id)
                    SELECT UUID(), latest_service_plan_id, id
                    FROM selected_orgs
                    ORDER BY RAND()
                    LIMIT num_visible_orgs;
                END WHILE;
        END WHILE;
END;