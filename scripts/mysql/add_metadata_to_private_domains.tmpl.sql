CREATE PROCEDURE add_metadata_to_private_domains(num_labels_per_domain INT, num_annotations_per_domain INT)
BEGIN
    DECLARE done INT DEFAULT 0;
    DECLARE domain_guid VARCHAR(255);
    DECLARE i INT;
    DECLARE domains_cursor CURSOR FOR SELECT guid FROM domains WHERE name LIKE '{{.Prefix}}-private-domain-%';
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = 1;

    OPEN domains_cursor;
    domain_loop:
    LOOP
        FETCH domains_cursor INTO domain_guid;
        IF done = 1 THEN
            LEAVE domain_loop;
        END IF;

        SET i = 0;
        WHILE i < num_labels_per_domain DO
            SET i = i + 1;
            INSERT INTO domain_labels (guid, resource_guid, key_name, value)
            VALUES (uuid(), domain_guid, CONCAT('{{.Prefix}}-key-', i), CONCAT('{{.Prefix}}-value-', i));
        END WHILE;

        SET i = 0;
        WHILE i < num_annotations_per_domain DO
            SET i = i + 1;
            INSERT INTO domain_annotations (guid, resource_guid, key_name, value)
            VALUES (uuid(), domain_guid, CONCAT('{{.Prefix}}-key-', i), CONCAT('{{.Prefix}}-value-', i));
        END WHILE;
    END LOOP;
    CLOSE domains_cursor;
END;
