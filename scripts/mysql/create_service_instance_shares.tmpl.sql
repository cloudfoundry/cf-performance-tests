DROP PROCEDURE IF EXISTS create_service_instance_shares;

CREATE PROCEDURE create_service_instance_shares(
    IN orgs INT, 
    IN spacesPerOrg INT, 
    IN serviceInstanceSharesPerSpace INT, 
    IN namePrefix VARCHAR(255)
)
BEGIN
    DECLARE i INT DEFAULT 0;
    DECLARE j INT DEFAULT 0;
    DECLARE k INT DEFAULT 0;
    DECLARE spaceOffset INT;
    DECLARE spaceId INT;
    DECLARE shareSpaceGuid VARCHAR(255);
    DECLARE serviceInstanceGuid VARCHAR(255);

    -- Loop through organizations
    WHILE i < orgs DO
        SET j = 0;
        
        -- Loop through spaces per organization
        WHILE j < spacesPerOrg DO
            -- Calculate spaceOffset and get spaceId
            SET spaceOffset = i * spacesPerOrg + j;

            -- Get spaceId
            SELECT id INTO spaceId
            FROM spaces
            WHERE name LIKE CONCAT(namePrefix, '-space-%')
            LIMIT 1 OFFSET spaceOffset;

            SET k = 0;
            -- Loop through the service instance shares
            WHILE k < serviceInstanceSharesPerSpace DO
                -- Find a random space to share that isn't our space
                SELECT guid INTO shareSpaceGuid
                FROM spaces
                WHERE name LIKE CONCAT(namePrefix, '-space-%')
                AND id != spaceId
                ORDER BY RAND()
                LIMIT 1;

                -- Find service instance for the current space
                SELECT service_instances.guid INTO serviceInstanceGuid
                FROM service_instances
                JOIN spaces ON service_instances.space_id = spaces.id
                WHERE spaces.id = spaceId
                AND service_instances.name LIKE CONCAT(namePrefix, '-service-instance-%')
                LIMIT 1 OFFSET k;

                -- Create the share for the service instance
                INSERT INTO service_instance_shares (service_instance_guid, target_space_guid)
                VALUES (serviceInstanceGuid, shareSpaceGuid);

                SET k = k + 1;
            END WHILE;

            SET j = j + 1;
        END WHILE;

        SET i = i + 1;
    END WHILE;
END;
