CREATE PROCEDURE create_shared_domains(num_shared_domains INT)
BEGIN
    DECLARE shared_domain_guid VARCHAR(255);
    DECLARE counter INT;
    SET counter = 0;
    WHILE counter < num_shared_domains
        DO
            SET counter = counter + 1;
            SET shared_domain_guid = uuid();
            INSERT INTO domains (guid, name)
            VALUES (shared_domain_guid, CONCAT('{{.Prefix}}-shared-domain-', shared_domain_guid));
        END WHILE;
END;