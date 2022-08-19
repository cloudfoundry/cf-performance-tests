CREATE PROCEDURE create_spaces(num_spaces_per_org INT)
BEGIN
    DECLARE org_id INT;
    DECLARE space_guid VARCHAR(255);
    DECLARE counter INT;
    DECLARE finished BOOLEAN DEFAULT FALSE;
    DECLARE orgs_cursor CURSOR FOR SELECT id FROM organizations WHERE name LIKE '{{.Prefix}}-org-%';
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET finished = TRUE;

    OPEN orgs_cursor;
    org_loop:
    LOOP
        FETCH orgs_cursor INTO org_id;
        IF finished = TRUE THEN
            LEAVE org_loop;
        END IF;
        SET counter = 0;
        WHILE counter < num_spaces_per_org
            DO
                SET counter = counter + 1;
                SET space_guid = uuid();
                INSERT INTO spaces (guid, name, organization_id)
                VALUES (space_guid, CONCAT('{{.Prefix}}-space-', space_guid), org_id);
                INSERT INTO space_labels (guid, key_name, resource_guid)
                VALUES (space_guid, '{{.Prefix}}', space_guid);
            END WHILE;
    END LOOP;
    CLOSE orgs_cursor;
END;