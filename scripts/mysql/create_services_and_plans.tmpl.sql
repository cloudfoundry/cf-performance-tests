CREATE PROCEDURE create_services_and_plans(num_services INT,
                                           service_broker_id INT,
                                           num_service_plans INT,
                                           service_plan_public BOOLEAN,
                                           num_visible_orgs INT,
                                           with_boilerplate BOOLEAN)
BEGIN
    DECLARE service_guid VARCHAR(255);
    DECLARE service_bindable BOOLEAN DEFAULT TRUE;
    DECLARE service_plan_guid VARCHAR(255);
    DECLARE services_counter INT DEFAULT 0;
    DECLARE boilerplate TEXT;
    DECLARE service_plan_free BOOLEAN DEFAULT TRUE;
    DECLARE service_plans_counter INT DEFAULT 0;
    DECLARE latest_service_id INT;
    DECLARE latest_service_plan_id INT;

    IF with_boilerplate = true THEN
        SET boilerplate = CONCAT('Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.',
                                 'Duis autem vel eum iriure dolor in hendrerit in vulputate velit esse molestie consequat, vel illum dolore eu feugiat nulla facilisis at vero eros et accumsan et iusto odio dignissim qui blandit praesent luptatum zzril delenit augue duis dolore te feugait nulla facilisi. Lorem ipsum dolor sit amet, consectetuer adipiscing elit, sed diam nonummy nibh euismod tincidunt ut laoreet dolore magna aliquam erat volutpat.',
                                 'Ut wisi enim ad minim veniam, quis nostrud exerci tation ullamcorper suscipit lobortis nisl ut aliquip ex ea commodo consequat. Duis autem vel eum iriure dolor in hendrerit in vulputate velit esse molestie consequat, vel illum dolore eu feugiat nulla facilisis at vero eros et accumsan et iusto odio dignissim qui blandit praesent luptatum zzril delenit augue duis dolore te feugait nulla facilisi.',
                                 'Nam liber tempor cum soluta nobis eleifend option congue nihil imperdiet doming id quod mazim placerat facer possim assum. Lorem ipsum dolor sit amet, consectetuer adipiscing elit, sed diam nonummy nibh euismod tincidunt ut laoreet dolore magna aliquam erat volutpat. Ut wisi enim ad minim veniam, quis nostrud exerci tation ullamcorper suscipit lobortis nisl ut aliquip ex ea commodo consequat.',
                                 'Duis autem vel eum iriure dolor in hendrerit in vulputate velit esse molestie consequat, vel illum dolore eu feugiat nulla facilisis.',
                                 'At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, At accusam aliquyam diam diam dolore dolores duo eirmod eos erat, et nonumy sed tempor et et invidunt justo labore Stet clita ea et gubergren, kasd magna no rebum. sanctus sea sed takimata ut vero voluptua. est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat.',
                                 'Consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus.');
    ELSE
        SET boilerplate = '';
    END IF;

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
                    INSERT INTO service_plans (guid, name, description, free, service_id, unique_id, public, extra, create_instance_schema, update_instance_schema, create_binding_schema)
                    VALUES (service_plan_guid,
                            CONCAT('{{.Prefix}}-service-plan-', service_plan_guid),
                            CONCAT('{{.Prefix}}-service-plan-description-', service_plan_guid, boilerplate),
                            service_plan_free,
                            latest_service_id,
                            CONCAT('unique-', service_plan_guid),
                            service_plan_public,
                            '{"shareable": true}',
                            boilerplate,
                            boilerplate,
                            boilerplate);
                    SET latest_service_plan_id = LAST_INSERT_ID();
                    INSERT INTO service_plan_visibilities (guid, service_plan_id, organization_id)
                    SELECT UUID(), latest_service_plan_id, id
                    FROM selected_orgs
                    ORDER BY RAND()
                    LIMIT num_visible_orgs;
                END WHILE;
        END WHILE;
END;