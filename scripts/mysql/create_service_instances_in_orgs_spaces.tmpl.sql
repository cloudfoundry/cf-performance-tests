DROP PROCEDURE IF EXISTS create_service_instances_for_orgs_spaces_plans;

CREATE PROCEDURE create_service_instances_for_orgs_spaces_plans(
    IN orgs INT, 
    IN spacesPerOrg INT, 
    IN servicePlans INT, 
    IN instancesPerPlanPerSpace INT, 
    IN namePrefix VARCHAR(255)
)
BEGIN
    DECLARE i INT DEFAULT 0;
    DECLARE j INT DEFAULT 0;
    DECLARE k INT DEFAULT 0;
    DECLARE spaceOffset INT;
    DECLARE servicePlanOffset INT;
    DECLARE spaceId INT;
    DECLARE servicePlanId INT;

    -- Loop through organizations
    WHILE i < orgs DO
        SET j = 0;
        
        -- Loop through spaces per organization
        WHILE j < spacesPerOrg DO
            -- Calculate spaceOffset
            SET spaceOffset = i * spacesPerOrg + j;

            -- Get spaceId
            SELECT id INTO spaceId
            FROM spaces
            WHERE name LIKE CONCAT(namePrefix, '-space-%')
            LIMIT 1 OFFSET spaceOffset;

            SET k = 0;
            -- Loop through service plans
            WHILE k < servicePlans DO
                -- Calculate servicePlanOffset
                SET servicePlanOffset = k;

                -- Get servicePlanId
                SELECT id INTO servicePlanId
                FROM service_plans
                WHERE name LIKE CONCAT(namePrefix, '-service-plan-%')
                LIMIT 1 OFFSET servicePlanOffset;

                -- Call the stored procedure to create service instances
                CALL create_service_instances(spaceId, servicePlanId, instancesPerPlanPerSpace);

                SET k = k + 1;
            END WHILE;

            SET j = j + 1;
        END WHILE;

        SET i = i + 1;
    END WHILE;
END;
