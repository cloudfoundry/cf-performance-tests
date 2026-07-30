CREATE PROCEDURE create_org_quotas_and_distribute_orgs(
    num_quotas INT
)
BEGIN
    DECLARE quota_name_prefix VARCHAR(255) DEFAULT '{{.Prefix}}-org-quota-';
    DECLARE org_name_query VARCHAR(255) DEFAULT '{{.Prefix}}-org-%';
    DECLARE i INT DEFAULT 0;

    -- create the quotas
    WHILE i < num_quotas DO
        INSERT INTO quota_definitions
            (guid, name, non_basic_services_allowed, total_services, memory_limit, total_routes)
        VALUES
            (uuid(), CONCAT(quota_name_prefix, uuid()), true, -1, -1, -1);
        SET i = i + 1;
    END WHILE;

    -- distribute prefixed orgs across the new quotas round-robin
    UPDATE organizations o
    JOIN (
        SELECT id, (ROW_NUMBER() OVER (ORDER BY id) - 1) % num_quotas AS quota_idx
        FROM organizations
        WHERE name LIKE org_name_query
    ) sub ON o.id = sub.id
    JOIN (
        SELECT id, ROW_NUMBER() OVER (ORDER BY id) - 1 AS rn
        FROM quota_definitions
        WHERE name LIKE CONCAT(quota_name_prefix, '%')
    ) q ON sub.quota_idx = q.rn
    SET o.quota_definition_id = q.id;
END;
