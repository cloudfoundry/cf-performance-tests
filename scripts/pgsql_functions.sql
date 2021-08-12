-- FUNC DEF:
CREATE OR REPLACE FUNCTION create_orgs(
    num_orgs INTEGER
) RETURNS void AS
$$
DECLARE
    org_guid text;
    org_name_prefix text := 'perf-org-';
    default_quota_definition_id int := 1;
BEGIN
    FOR i IN 1..num_orgs LOOP
        org_guid := gen_random_uuid();
        INSERT INTO organizations (guid, name, quota_definition_id) VALUES (org_guid, org_name_prefix || org_guid, default_quota_definition_id);
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- ============================================================= --

-- FUNC DEF:
CREATE OR REPLACE FUNCTION create_spaces(
    num_spaces_per_org INTEGER
) RETURNS void AS
$$
DECLARE
    org_id int;
    org_name_query text := 'perf-org-%';
    space_guid text;
    space_name_prefix text := 'perf-space-';
BEGIN
    FOR org_id IN (SELECT id FROM organizations WHERE name LIKE org_name_query) LOOP
        FOR i IN 1..num_spaces_per_org LOOP
            space_guid := gen_random_uuid();
            INSERT INTO spaces (guid, name, organization_id) VALUES (space_guid, space_name_prefix || space_guid, org_id);
        END LOOP;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- ============================================================= --

-- FUNC DEF:
CREATE OR REPLACE FUNCTION create_security_groups(
    security_groups INTEGER
) RETURNS void AS
$$
DECLARE
    security_group_guid text;
    security_group_name_prefix text := 'perf-security-group-';
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
    FOR i IN 1..security_groups LOOP
        security_group_guid := gen_random_uuid();
        INSERT INTO security_groups (guid, name, rules) VALUES (security_group_guid, security_group_name_prefix || security_group_guid, security_rule);
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- ============================================================= --

-- FUNC DEF:
CREATE OR REPLACE FUNCTION assign_security_groups_to_spaces(
    num_spaces INTEGER,
    num_security_groups_per_space INTEGER
) RETURNS void AS
$$
DECLARE
    space_id int;
    space_name_query text := 'perf-space-%';
    security_group_id int;
    security_group_name_query text := 'perf-security-group-%';
BEGIN
    FOR space_id IN (SELECT id FROM spaces WHERE name LIKE space_name_query ORDER BY random() LIMIT num_spaces) LOOP
        FOR security_group_id IN (SELECT id FROM security_groups WHERE name LIKE security_group_name_query ORDER BY random() LIMIT num_security_groups_per_space) LOOP
            INSERT INTO security_groups_spaces (security_group_id, space_id) VALUES (security_group_id, space_id);
        END LOOP;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- ============================================================= --

-- FUNC DEF:
CREATE OR REPLACE FUNCTION assign_user_as_space_developer(
    user_guid TEXT,
    num_spaces INTEGER
) RETURNS void AS
$$
DECLARE
    user_id int;
    space_id int;
    space_name_query text := 'perf-space-%';
BEGIN
    SELECT id FROM users WHERE guid = user_guid INTO user_id;
    FOR space_id IN (SELECT id FROM spaces WHERE name LIKE space_name_query ORDER BY random() LIMIT num_spaces) LOOP
        INSERT INTO spaces_developers (space_id, user_id) VALUES (space_id, user_id);
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- ============================================================= --

-- FUNC DEF:
CREATE OR REPLACE FUNCTION create_shared_domains(
    num_shared_domains INTEGER
) RETURNS void AS
$$
DECLARE
    shared_domain_guid text;
    shared_domain_name_prefix text := 'perf-shared-domain-';
BEGIN
    FOR i IN 1..num_shared_domains LOOP
        shared_domain_guid := gen_random_uuid();
        INSERT INTO domains (guid, name) VALUES (shared_domain_guid, shared_domain_name_prefix || shared_domain_guid);
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- ============================================================= --

-- FUNC DEF:
CREATE OR REPLACE FUNCTION create_private_domains(
    num_private_domains INTEGER
) RETURNS void AS
$$
DECLARE
    org_id int;
    org_name_query text := 'perf-org-%';
    num_created_private_domains int := 0;
    private_domain_guid text;
    private_domain_name_prefix text := 'perf-private-domain-';
BEGIN
    LOOP
        FOR org_id IN (SELECT id FROM organizations WHERE name LIKE org_name_query ORDER BY random()) LOOP
            IF num_created_private_domains = num_private_domains THEN
                RETURN;
            END IF;
            private_domain_guid := gen_random_uuid();
            INSERT INTO domains (guid, name, owning_organization_id) VALUES (private_domain_guid, private_domain_name_prefix || private_domain_guid, org_id);
            num_created_private_domains := num_created_private_domains + 1;
        END LOOP;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- ============================================================= --

-- FUNC DEF:
CREATE OR REPLACE FUNCTION assign_user_as_org_manager(
    user_guid TEXT,
    num_orgs INTEGER
) RETURNS void AS
$$
DECLARE
    user_id int;
    org_id int;
    org_name_query text := 'perf-org-%';
BEGIN
    SELECT id FROM users WHERE guid = user_guid INTO user_id;
    FOR org_id IN (SELECT id FROM organizations WHERE name LIKE org_name_query ORDER BY random() LIMIT num_orgs) LOOP
        INSERT INTO organizations_managers (organization_id, user_id) VALUES (org_id, user_id);
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- ============================================================= --

-- FUNC DEF:
CREATE OR REPLACE FUNCTION create_isolation_segments(
    num_isolation_segments INTEGER
) RETURNS void AS
$$
DECLARE
    isolation_segment_guid text;
    isolation_segment_name_prefix text := 'perf-isolation-segment-';
BEGIN
    FOR i IN 1..num_isolation_segments LOOP
        isolation_segment_guid := gen_random_uuid();
        INSERT INTO isolation_segments (guid, name) VALUES (isolation_segment_guid, isolation_segment_name_prefix || isolation_segment_guid);
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- ============================================================= --

-- FUNC DEF:
CREATE OR REPLACE FUNCTION assign_orgs_to_isolation_segments(
    num_orgs INTEGER
) RETURNS void AS
$$
DECLARE
    org_guid text;
    org_name_query text := 'perf-org-%';
    isolation_segment_name_query text := 'perf-isolation-segment-%';
    isolation_segment_guid text;
BEGIN
    FOR org_guid IN (SELECT guid FROM organizations WHERE name LIKE org_name_query ORDER BY random() LIMIT num_orgs) LOOP
        SELECT guid FROM isolation_segments WHERE name LIKE isolation_segment_name_query ORDER BY random() LIMIT 1 INTO isolation_segment_guid;
        INSERT INTO organizations_isolation_segments (organization_guid, isolation_segment_guid) VALUES (org_guid, isolation_segment_guid);
    END LOOP;
END;
$$ LANGUAGE plpgsql;
