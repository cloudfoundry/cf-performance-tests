CREATE FUNCTION create_service_broker() RETURNS INT
BEGIN
    DECLARE service_broker_guid VARCHAR(255);
    SET service_broker_guid = UUID();

    INSERT INTO service_brokers (guid, name, broker_url, auth_password)
    VALUES (service_broker_guid, CONCAT('{{.Prefix}}-service-broker-', service_broker_guid), '', '');

    RETURN LAST_INSERT_ID();
END;
