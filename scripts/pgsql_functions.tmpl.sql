-- FUNC DEF:
CREATE OR REPLACE FUNCTION create_orgs(
    num_orgs INTEGER
) RETURNS void AS
$$
DECLARE
    org_guid text;
    org_name_prefix text := '{{.Prefix}}-org-';
    default_quota_definition_id int := 1;
BEGIN
    FOR _ IN 1..num_orgs LOOP
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
    org_name_query text := '{{.Prefix}}-org-%';
    space_guid text;
    space_name_prefix text := '{{.Prefix}}-space-';
BEGIN
    FOR org_id IN (SELECT id FROM organizations WHERE name LIKE org_name_query) LOOP
        FOR _ IN 1..num_spaces_per_org LOOP
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
    security_group_name_prefix text := '{{.Prefix}}-security-group-';
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
    FOR _ IN 1..security_groups LOOP
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
    v_space_id int;
    space_name_query text := '{{.Prefix}}-space-%';
    v_security_group_id int;
    security_group_name_query text := '{{.Prefix}}-security-group-%';
BEGIN
    FOR v_space_id IN (SELECT id FROM spaces WHERE name LIKE space_name_query ORDER BY random() LIMIT num_spaces) LOOP
        FOR v_security_group_id IN (SELECT id FROM security_groups WHERE name LIKE security_group_name_query ORDER BY random() LIMIT num_security_groups_per_space) LOOP
            INSERT INTO security_groups_spaces (security_group_id, space_id) VALUES (v_security_group_id, v_space_id);
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
    v_user_id int;
    v_space_id int;
    space_name_query text := '{{.Prefix}}-space-%';
BEGIN
    SELECT id FROM users WHERE guid = user_guid INTO v_user_id;
    FOR v_space_id IN (SELECT id FROM spaces WHERE name LIKE space_name_query ORDER BY random() LIMIT num_spaces) LOOP
        INSERT INTO spaces_developers (space_id, user_id) VALUES (v_space_id, v_user_id);
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
    shared_domain_name_prefix text := '{{.Prefix}}-shared-domain-';
BEGIN
    FOR _ IN 1..num_shared_domains LOOP
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
    org_name_query text := '{{.Prefix}}-org-%';
    num_created_private_domains int := 0;
    private_domain_guid text;
    private_domain_name_prefix text := '{{.Prefix}}-private-domain-';
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
    v_user_id int;
    org_id int;
    org_name_query text := '{{.Prefix}}-org-%';
BEGIN
    SELECT id FROM users WHERE guid = user_guid INTO v_user_id;
    FOR org_id IN (SELECT id FROM organizations WHERE name LIKE org_name_query ORDER BY random() LIMIT num_orgs) LOOP
        INSERT INTO organizations_managers (organization_id, user_id) VALUES (org_id, v_user_id);
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
    isolation_segment_name_prefix text := '{{.Prefix}}-isolation-segment-';
BEGIN
    FOR _ IN 1..num_isolation_segments LOOP
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
    org_name_query text := '{{.Prefix}}-org-%';
    isolation_segment_name_query text := '{{.Prefix}}-isolation-segment-%';
    v_isolation_segment_guid text;
BEGIN
    FOR org_guid IN (SELECT guid FROM organizations WHERE name LIKE org_name_query ORDER BY random() LIMIT num_orgs) LOOP
        SELECT guid FROM isolation_segments WHERE name LIKE isolation_segment_name_query ORDER BY random() LIMIT 1 INTO v_isolation_segment_guid;
        INSERT INTO organizations_isolation_segments (organization_guid, isolation_segment_guid) VALUES (org_guid, v_isolation_segment_guid);
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- ============================================================= --

-- FUNC DEF:
CREATE OR REPLACE FUNCTION create_service_instances(
    p_space_id INTEGER,
    p_service_plan_id INTEGER,
    num_service_instances INTEGER
) RETURNS void AS
$$
DECLARE
    service_instance_guid text;
    service_instance_name_prefix text := '{{.Prefix}}-service-instance-';
BEGIN
    FOR _ IN 1..num_service_instances LOOP
        service_instance_guid := gen_random_uuid();
        INSERT INTO service_instances (guid, name, space_id, service_plan_id) VALUES (service_instance_guid, service_instance_name_prefix || service_instance_guid, p_space_id, p_service_plan_id);
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- ============================================================= --

-- FUNC DEF:
CREATE OR REPLACE FUNCTION create_service_keys_for_service_instances(
    p_space_id INTEGER,
    num_service_keys_per_service_instance INTEGER
) RETURNS void AS
$$
DECLARE
    v_service_instance_id int;
    service_key_guid text;
    service_key_name_prefix text := '{{.Prefix}}-service-key-';
BEGIN
    FOR v_service_instance_id IN (SELECT id FROM service_instances WHERE space_id = p_space_id) LOOP
        FOR _ IN 1..num_service_keys_per_service_instance LOOP
            service_key_guid := gen_random_uuid();
            INSERT INTO service_keys (guid, name, credentials, service_instance_id) VALUES (service_key_guid, service_key_name_prefix || service_key_guid, '', v_service_instance_id);
        END LOOP;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- ============================================================= --

-- FUNC DEF:
CREATE OR REPLACE FUNCTION create_services_and_plans(
    num_services INTEGER,
    service_broker_id INTEGER,
    num_service_plans INTEGER,
    service_plan_public BOOLEAN,
    num_visible_orgs INTEGER
) RETURNS void AS
$$
DECLARE
    service_guid TEXT;
    service_label_prefix TEXT := '{{.Prefix}}-service-';
    service_description_prefix TEXT := '{{.Prefix}}-service-description-';
    service_bindable BOOLEAN := true;
    service_plan_guid TEXT;
    service_plan_name_prefix TEXT := '{{.Prefix}}-service-plan';
    service_plan_description_prefix TEXT := '{{.Prefix}}-service-plan-description-';
    service_plan_free BOOLEAN := true;
    latest_service_id INTEGER;
    latest_service_plan_id INTEGER;

BEGIN
    FOR _ IN 1..num_services LOOP
        service_guid := gen_random_uuid();
        INSERT INTO services (guid, label, description, bindable, service_broker_id)
            VALUES (
                service_guid,
                service_label_prefix || service_guid,
                service_description_prefix || service_guid,
                service_bindable,
                service_broker_id
                ) RETURNING id INTO latest_service_id;
        FOR _ IN 1..num_service_plans LOOP
            service_plan_guid := gen_random_uuid();
            INSERT INTO service_plans (guid, name, description, free, service_id, unique_id, public)
                VALUES (
                       service_plan_guid,
                       service_plan_name_prefix || service_plan_guid,
                       service_plan_description_prefix || service_plan_guid,
                       service_plan_free,
                       latest_service_id,
                       'unique-' || service_plan_guid,
                       service_plan_public
                   ) RETURNING id INTO latest_service_plan_id;
            INSERT INTO service_plan_visibilities (guid, service_plan_id, organization_id)
                SELECT gen_random_uuid(), latest_service_plan_id, id
                FROM organizations ORDER BY random() LIMIT num_visible_orgs;
        END LOOP;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

