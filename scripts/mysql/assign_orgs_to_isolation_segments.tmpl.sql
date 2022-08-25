CREATE PROCEDURE assign_orgs_to_isolation_segments(num_orgs INT)
BEGIN
    DECLARE org_guid VARCHAR(255);
    DECLARE v_isolation_segment_guid VARCHAR(255);
    DECLARE finished BOOLEAN DEFAULT FALSE;
    DECLARE orgs_cursor CURSOR FOR SELECT guid
                                   FROM organizations
                                   WHERE name LIKE '{{.Prefix}}-org-%'
                                   ORDER BY RAND()
                                   LIMIT num_orgs;
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET finished = TRUE;

    OPEN orgs_cursor;
    org_loop:
    LOOP
        FETCH orgs_cursor INTO org_guid;
        IF finished THEN
            LEAVE org_loop;
        END IF;
        SELECT guid
        FROM isolation_segments
        WHERE name LIKE '{{.Prefix}}-isolation-segment-%'
        ORDER BY RAND()
        LIMIT 1
        INTO v_isolation_segment_guid;
        INSERT INTO organizations_isolation_segments (organization_guid, isolation_segment_guid)
        VALUES (org_guid, v_isolation_segment_guid);
    END LOOP;
    CLOSE orgs_cursor;
END;