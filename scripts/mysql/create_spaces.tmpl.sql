CREATE PROCEDURE create_spaces(num_spaces_per_org INT)
BEGIN
    DECLARE _counter INT DEFAULT 0;

    WHILE _counter < num_spaces_per_org DO
        INSERT INTO spaces (guid, name, organization_id)
        SELECT UUID(), CONCAT('{{.Prefix}}-space-', UUID()), id
        FROM organizations WHERE name LIKE '{{.Prefix}}-org-%';
        SET _counter = _counter + 1;
    END WHILE;

    INSERT INTO space_labels (guid, key_name, resource_guid)
    SELECT guid, '{{.Prefix}}', guid FROM spaces WHERE name LIKE CONCAT('{{.Prefix}}-space-%');
END;
