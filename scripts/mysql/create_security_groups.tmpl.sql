CREATE PROCEDURE create_security_groups(security_groups INT)
BEGIN
    DECLARE security_group_guid VARCHAR(255);
    DECLARE security_rule MEDIUMTEXT;
    DECLARE counter INT;
    SET security_rule = '[
        {
            "protocol": "icmp",
            "destination": "0.0.0.0/0",
            "type": 0,
            "code": 0
        },
        {
            "protocol": "tcp",
            "destination": "10.0.11.0/24",
            "ports": "80,443",
            "log": true,
            "description": "Allow http and https traffic to ZoneA"
        }
    ]';
    SET counter = 0;
    WHILE counter < security_groups
        DO
            SET counter = counter + 1;
            SET security_group_guid = uuid();
            INSERT INTO security_groups (guid, name, rules)
            VALUES (security_group_guid, CONCAT('{{.Prefix}}-security-group-', security_group_guid), security_rule);
        END WHILE;
END;