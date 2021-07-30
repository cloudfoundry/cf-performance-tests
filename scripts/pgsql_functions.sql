-- FUNC DEF:
CREATE FUNCTION create_test_orgs(
    orgs INTEGER
) RETURNS void AS
$$
DECLARE
org_guid text;
BEGIN
FOR i IN 1..orgs
        LOOP
            org_guid := gen_random_uuid();
INSERT INTO organizations (guid, name, quota_definition_id)
VALUES (org_guid, 'perf-test-org-' || org_guid, 1);
END LOOP;
END;
$$ LANGUAGE plpgsql;

-- ============================================================= --

-- FUNC DEF:
CREATE FUNCTION create_test_shared_domains(
    shared_domains INTEGER
) RETURNS void AS
$$
DECLARE
shared_domain_guid text;
BEGIN
FOR i IN 1..shared_domains
            LOOP
                shared_domain_guid := gen_random_uuid();
INSERT INTO domains (guid, name)
VALUES (shared_domain_guid, 'perf-test-shared-domain-' || shared_domain_guid);
END LOOP;
END;
$$ LANGUAGE plpgsql;

-- ============================================================= --

-- FUNC DEF:
CREATE FUNCTION create_test_private_domains(
    private_domains INTEGER
) RETURNS void AS
$$
DECLARE
private_domain_guid text;
BEGIN
FOR i IN 1..private_domains
            LOOP
                private_domain_guid := gen_random_uuid();
INSERT INTO domains (guid, name, owning_organization_id)
SELECT private_domain_guid, 'perf-test-private-domain-' || private_domain_guid, id
FROM organizations WHERE name != 'default' ORDER BY random() LIMIT 1;
END LOOP;
END;
$$ LANGUAGE plpgsql;

-- ============================================================= --

-- FUNC DEF:
CREATE FUNCTION create_test_spaces(
    spaces INTEGER
) RETURNS void AS
    $$
DECLARE
space_guid text;
    org_id int;
BEGIN
FOR org_id IN (SELECT id FROM organizations) LOOP
            FOR i IN 1..spaces
                LOOP
                    space_guid := gen_random_uuid();
INSERT INTO spaces (guid, name, organization_id)
SELECT space_guid, 'perf-test-space-' || space_guid, org_id;
END LOOP;
END LOOP;
END;
$$ LANGUAGE plpgsql;

-- ============================================================= --

-- FUNC DEF:
CREATE FUNCTION create_test_security_groups(
    security_groups INTEGER
) RETURNS void AS
    $$
DECLARE
security_group_guid text;
    security_rule text := '[
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
BEGIN
FOR i IN 1..security_groups
        LOOP
            security_group_guid := gen_random_uuid();
INSERT INTO security_groups (guid, name, rules)
SELECT security_group_guid, 'perf-test-security-group-' || security_group_guid, security_rule;
END LOOP;
END;
$$ LANGUAGE plpgsql;

-- ============================================================= --

-- FUNC DEF:
CREATE FUNCTION create_test_security_group_spaces(
) RETURNS void AS
    $$
DECLARE
security_group_id int;
    space_id int;
BEGIN
FOR space_id IN (SELECT id FROM spaces WHERE name LIKE 'perf-test-space-%') LOOP
            FOR i IN 1..5
                LOOP
SELECT id FROM security_groups WHERE name LIKE 'perf-%' ORDER BY random() LIMIT 1 INTO security_group_id;
INSERT INTO security_groups_spaces (security_group_id, space_id)
SELECT security_group_id, space_id;
END LOOP;
END LOOP;
END;
$$ LANGUAGE plpgsql;

