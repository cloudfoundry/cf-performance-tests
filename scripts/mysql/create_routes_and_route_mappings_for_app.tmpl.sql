CREATE PROCEDURE create_routes_and_route_mappings_for_app(
    IN app_guid VARCHAR(255),
    IN org_name VARCHAR(255),
    IN space_guid VARCHAR(255),
    IN num_route_mappings INT)
BEGIN
    DECLARE default_domain_id INT;
    DECLARE space_id INT;
    DECLARE quota_id INT;
    DECLARE route_guid VARCHAR(255);
    DECLARE route_mapping_guid VARCHAR(255);
    DECLARE process_type VARCHAR(20);
    DECLARE i INT;

    SET default_domain_id = 1;
    SET process_type = 'web';
    SET i = 1;

    SELECT quota_definition_id INTO quota_id FROM organizations WHERE name = org_name;
    UPDATE quota_definitions SET total_routes = -1 WHERE id = quota_id;

    SELECT id INTO space_id FROM spaces WHERE guid = space_guid;

    START TRANSACTION;
        WHILE i <= num_route_mappings DO
--          shorten guid to be able to map more routes to the app (diego limitation)
            SET route_guid = (SELECT LEFT(UUID(), 13));
            INSERT INTO routes (guid, domain_id, space_id, host) VALUES (route_guid, default_domain_id, space_id, CONCAT('{{.Prefix}}-', route_guid));

            SET route_mapping_guid = (SELECT UUID());
            INSERT INTO route_mappings (guid, app_guid, route_guid, process_type) VALUES (route_mapping_guid, app_guid, route_guid, process_type);

            SET i = i + 1;
        END WHILE;
    COMMIT;
END;
