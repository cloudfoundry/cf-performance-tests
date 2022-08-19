CREATE PROCEDURE create_isolation_segments(num_isolation_segments INT)
BEGIN
    DECLARE isolation_segment_guid VARCHAR(255);
    DECLARE counter INT;
    SET counter = 0;

    WHILE counter < num_isolation_segments
        DO
            SET counter = counter + 1;
            SET isolation_segment_guid = uuid();
            INSERT INTO isolation_segments (guid, name)
            VALUES (isolation_segment_guid, CONCAT('{{.Prefix}}-isolation-segment-', isolation_segment_guid));
        END WHILE;
END;