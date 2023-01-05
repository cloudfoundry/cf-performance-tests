CREATE PROCEDURE create_events()
BEGIN
    DECLARE counter INT;
    DECLARE event_type text;
    DECLARE amount INT;
    DECLARE num_events INT;
    DECLARE org_guid VARCHAR(255);
    DECLARE space_guid VARCHAR(255);
    DECLARE events_guid VARCHAR(255);
    DECLARE finished BOOLEAN DEFAULT FALSE;
    DECLARE events_cursor CURSOR FOR SELECT audit_event_type, count_events FROM event_types;
    DECLARE import_cursor CURSOR FOR SELECT ean, sku, mpn, manufacturerName, manufacturerUniqueId, images FROM importjob;
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET finished = TRUE;
    -- create a temporary table where the events will be stored "temporarily", is faster than inserting directly to events
    CREATE TEMPORARY TABLE IF NOT EXISTS temp_events as select * from events;
    OPEN events_cursor;
    events_loop:
    LOOP
        FETCH events_cursor INTO event_type, num_events;
        IF finished = TRUE THEN
            LEAVE events_loop;
        END IF;
        SET counter = 0;
        WHILE counter < num_events
            DO
                SET events_guid = uuid();
                SELECT guid FROM organizations WHERE name LIKE '{{.Prefix}}-%' ORDER BY RAND() LIMIT 1 INTO org_guid;
                SELECT guid FROM spaces WHERE name LIKE '{{.Prefix}}-space-%' ORDER BY RAND() LIMIT 1 INTO space_guid;
                INSERT INTO temp_events (guid, timestamp, type, actor, actor_type, actee, actee_type, organization_guid, space_guid)
                VALUES (events_guid, current_timestamp, event_type, CONCAT('{{.Prefix}}-events-actor-', events_guid), CONCAT('{{.Prefix}}-events-actor-type-', events_guid),
                        CONCAT('{{.Prefix}}-events-actee-', events_guid), CONCAT('{{.Prefix}}-events-actee-type-', events_guid), org_guid, space_guid);
                SET counter = counter + 1;
            END WHILE;
    END LOOP;
    CLOSE events_cursor;
    -- delete entries in events table, because in the temp_events id 1 exists and that values will be copied into
    -- events where id 1 already exists..
    TRUNCATE events;
    INSERT INTO events SELECT * FROM temp_events;
    DROP TABLE IF EXISTS temp_events;
END;
