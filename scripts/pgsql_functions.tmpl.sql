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
CREATE OR REPLACE FUNCTION create_selected_orgs_table(
    num_orgs INTEGER
) RETURNS void AS
$$
DECLARE
BEGIN
    DROP TABLE IF EXISTS selected_orgs;

    CREATE TABLE selected_orgs(id INT NOT NULL PRIMARY KEY);

    INSERT INTO selected_orgs (id)
    SELECT id FROM organizations
    WHERE name LIKE '{{.Prefix}}-org-%'
    ORDER BY random()
    LIMIT num_orgs;
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
    space_name_query text := '{{.Prefix}}-space-%';
BEGIN
    FOR _ IN 1..num_spaces_per_org LOOP
        INSERT INTO spaces (guid, name, organization_id) SELECT md5(random()::text || clock_timestamp()::text)::uuid AS guid, space_name_prefix || md5(random()::text) AS name, id AS organization_id FROM organizations WHERE name LIKE org_name_query;
    END LOOP;

    INSERT INTO space_labels (guid, key_name, resource_guid) SELECT guid, '{{.Prefix}}' AS key_name, guid AS resource_guid FROM spaces WHERE name LIKE space_name_query;
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
    FOR v_security_group_id IN (SELECT id FROM security_groups WHERE name LIKE security_group_name_query ORDER BY random() LIMIT num_security_groups_per_space) LOOP
        FOR v_space_id IN (SELECT id FROM spaces WHERE name LIKE space_name_query ORDER BY random() LIMIT num_spaces) LOOP
            INSERT INTO security_groups_spaces (security_group_id, space_id) VALUES (v_security_group_id, v_space_id);
        END LOOP;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- ============================================================= --

-- FUNC DEF:
CREATE OR REPLACE FUNCTION assign_user_as_space_role(
    user_guid TEXT,
    space_role TEXT,
    num_spaces INTEGER
) RETURNS void AS
$$
DECLARE
    v_user_id int;
    v_space_id int;
    space_name_query text := '{{.Prefix}}-space-%';
BEGIN
    SELECT id FROM users WHERE guid = user_guid INTO v_user_id;

    FOR v_space_id IN (SELECT spaces.id FROM spaces JOIN selected_orgs ON spaces.organization_id = selected_orgs.id WHERE spaces.name LIKE space_name_query ORDER BY random() LIMIT num_spaces) LOOP
        EXECUTE FORMAT('INSERT INTO %s (space_id, user_id) VALUES (%s, %s)', space_role, v_space_id, v_user_id);
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
CREATE OR REPLACE FUNCTION assign_user_as_org_role(
    user_guid TEXT,
    org_role TEXT,
    num_orgs INTEGER
) RETURNS void AS
$$
DECLARE
    v_user_id int;
    v_org_id int;
    org_name_query text := '{{.Prefix}}-org-%';
BEGIN
    SELECT id FROM users WHERE guid = user_guid INTO v_user_id;

    FOR v_org_id IN (SELECT id FROM selected_orgs ORDER BY random() LIMIT num_orgs) LOOP
        EXECUTE FORMAT('INSERT INTO %s (organization_id, user_id) VALUES (%s, %s)', org_role, v_org_id, v_user_id);
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
    visible_orgs_per_plan INTEGER,
    with_boilerplate BOOLEAN
) RETURNS void AS
$$
DECLARE
    service_guid TEXT;
    service_label_prefix TEXT := '{{.Prefix}}-service-';
    service_description_prefix TEXT := '{{.Prefix}}-service-description-';
    service_bindable BOOLEAN := true;
    service_plan_guid TEXT;
    service_plan_name_prefix TEXT := '{{.Prefix}}-service-plan-';
    service_plan_description_prefix TEXT := '{{.Prefix}}-service-plan-description-';
    boilerplate TEXT;
    service_plan_free BOOLEAN := true;
    latest_service_id INTEGER;
    latest_service_plan_id INTEGER;

BEGIN
    IF with_boilerplate = true THEN
        boilerplate := 'Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.' ||
                       'Duis autem vel eum iriure dolor in hendrerit in vulputate velit esse molestie consequat, vel illum dolore eu feugiat nulla facilisis at vero eros et accumsan et iusto odio dignissim qui blandit praesent luptatum zzril delenit augue duis dolore te feugait nulla facilisi. Lorem ipsum dolor sit amet, consectetuer adipiscing elit, sed diam nonummy nibh euismod tincidunt ut laoreet dolore magna aliquam erat volutpat.' ||
                       'Ut wisi enim ad minim veniam, quis nostrud exerci tation ullamcorper suscipit lobortis nisl ut aliquip ex ea commodo consequat. Duis autem vel eum iriure dolor in hendrerit in vulputate velit esse molestie consequat, vel illum dolore eu feugiat nulla facilisis at vero eros et accumsan et iusto odio dignissim qui blandit praesent luptatum zzril delenit augue duis dolore te feugait nulla facilisi.' ||
                       'Nam liber tempor cum soluta nobis eleifend option congue nihil imperdiet doming id quod mazim placerat facer possim assum. Lorem ipsum dolor sit amet, consectetuer adipiscing elit, sed diam nonummy nibh euismod tincidunt ut laoreet dolore magna aliquam erat volutpat. Ut wisi enim ad minim veniam, quis nostrud exerci tation ullamcorper suscipit lobortis nisl ut aliquip ex ea commodo consequat.' ||
                       'Duis autem vel eum iriure dolor in hendrerit in vulputate velit esse molestie consequat, vel illum dolore eu feugiat nulla facilisis.' ||
                       'At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, At accusam aliquyam diam diam dolore dolores duo eirmod eos erat, et nonumy sed tempor et et invidunt justo labore Stet clita ea et gubergren, kasd magna no rebum. sanctus sea sed takimata ut vero voluptua. est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat.' ||
                       'Consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus.';
    ELSE
        boilerplate := '';
    END IF;

    FOR _ IN 1..num_services LOOP
        service_guid := gen_random_uuid();
        INSERT INTO services (guid, label, description, bindable, service_broker_id, extra)
            VALUES (
                service_guid,
                service_label_prefix || service_guid,
                service_description_prefix || service_guid,
                service_bindable,
                service_broker_id,
                '{"shareable": true}'
                ) RETURNING id INTO latest_service_id;
        FOR _ IN 1..num_service_plans LOOP
            service_plan_guid := gen_random_uuid();
            INSERT INTO service_plans (guid, name, description, free, service_id, unique_id, public, extra, create_instance_schema, update_instance_schema, create_binding_schema)
                VALUES (
                       service_plan_guid,
                       service_plan_name_prefix || service_plan_guid,
                       service_plan_description_prefix || service_plan_guid || boilerplate,
                       service_plan_free,
                       latest_service_id,
                       'unique-' || service_plan_guid,
                       service_plan_public,
                       '{"shareable": true}',
                       boilerplate,
                       boilerplate,
                       boilerplate
                   ) RETURNING id INTO latest_service_plan_id;
            INSERT INTO service_plan_visibilities (guid, service_plan_id, organization_id)
                SELECT gen_random_uuid(), latest_service_plan_id, id
                FROM selected_orgs ORDER BY random() LIMIT visible_orgs_per_plan;
        END LOOP;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- ============================================================= --

-- FUNC DEF:
CREATE OR REPLACE FUNCTION create_event_types_table(
) RETURNS void AS
$$

BEGIN
    DROP TABLE IF EXISTS event_types;
    CREATE table event_types (id serial primary key , audit_event_type VARCHAR(128), count_events INT);
    INSERT INTO event_types (audit_event_type, count_events) VALUES ('audit.user.space_developer_add',100000),
                                                          ('audit.app.environment_variables.show',100000),
                                                          ('audit.service_binding.delete',100000),
                                                          ('audit.user.organization_manager_remove',50000),
                                                          ('audit.user.organization_billing_manager_remove',50000),
                                                          ('audit.service_binding.create',50000),
                                                          ('audit.service_instance.start_delete',50000),
                                                          ('audit.service_plan.update',50000),
                                                          ('audit.app.environment.show',50000),
                                                          ('audit.app.map-route',50000),
                                                          ('audit.app.unmap-route',50000),
                                                          ('audit.user.space_supporter_add',10000),
                                                          ('audit.app.process.crash',10000),
                                                          ('app.crash',10000),
                                                          ('audit.user.space_auditor_remove',10000),
                                                          ('audit.user.organization_manager_add',10000),
                                                          ('audit.user.space_manager_add',10000),
                                                          ('audit.user.space_supporter_remove',10000),
                                                          ('audit.app.build.create',10000),
                                                          ('audit.app.droplet.create',10000),
                                                          ('audit.app.process.update',10000),
                                                          ('audit.app.process.scale',10000),
                                                          ('audit.app.revision.create',10000),
                                                          ('audit.app.stop',10000),
                                                          ('audit.app.start',10000),
                                                          ('audit.service.update',10000),
                                                          ('audit.app.droplet.mapped',10000),
                                                          ('audit.app.package.create',10000),
                                                          ('audit.app.package.upload',10000),
                                                          ('audit.app.process.rescheduling',10000),
                                                          ('audit.route.create',10000),
                                                          ('audit.app.update',10000),
                                                          ('audit.app.package.delete',10000),
                                                          ('audit.app.droplet.delete',10000),
                                                          ('audit.service_broker.update',10000),
                                                          ('audit.app.process.delete',10000),
                                                          ('audit.user.space_manager_remove',10000),
                                                          ('audit.route.delete-request',10000),
                                                          ('audit.app.create',10000),
                                                          ('audit.app.delete-request',10000),
                                                          ('audit.service_instance.delete',10000),
                                                          ('audit.service_instance.create',10000),
                                                          ('audit.service_instance.update',10000),
                                                          ('audit.app.process.create',10000),
                                                          ('audit.app.ssh-authorized',5000),
                                                          ('audit.user.organization_auditor_add',5000),
                                                          ('audit.service_plan_visibility.update',5000),
                                                          ('audit.service_key.create',5000),
                                                          ('audit.user.organization_user_add',5000),
                                                          ('audit.service_key.delete',5000),
                                                          ('audit.service_instance.start_create',5000),
                                                          ('audit.app.apply_manifest',5000),
                                                          ('audit.user.space_auditor_add',5000),
                                                          ('audit.user.space_developer_remove',5000),
                                                          ('audit.user.organization_auditor_remove',5000),
                                                          ('audit.user.organization_user_remove',5000),
                                                          ('audit.app.restart',1000),
                                                          ('audit.service_plan.create',1000),
                                                          ('audit.service_plan.delete',1000),
                                                          ('audit.service_plan_visibility.delete',1000),
                                                          ('audit.app.upload-bits',1000),
                                                          ('audit.app.task.create',1000),
                                                          ('audit.user_provided_service_instance.update',1000),
                                                          ('audit.service_instance.start_update',1000),
                                                          ('audit.app.deployment.create',1000),
                                                          ('audit.space.create',1000),
                                                          ('audit.space.delete-request',1000),
                                                          ('audit.organization.update',500),
                                                          ('audit.user_provided_service_instance.create',500),
                                                          ('audit.user_provided_service_instance.delete',500),
                                                          ('audit.service_instance.unbind_route',500),
                                                          ('audit.service_instance.bind_route',500),
                                                          ('audit.app.restage',500),
                                                          ('audit.route.update',500),
                                                          ('audit.service.create',100),
                                                          ('audit.service.delete',100),
                                                          ('audit.service_broker.create',100),
                                                          ('audit.service_broker.delete',100),
                                                          ('audit.service_instance.purge',100),
                                                          ('audit.organization.delete-request',100),
                                                          ('audit.app.package.download',100),
                                                          ('audit.organization.create',100),
                                                          ('audit.app.copy-bits',100),
                                                          ('audit.service_key.update',100),
                                                          ('audit.service_instance.share',100),
                                                          ('audit.service_instance.unshare',100),
                                                          ('audit.service_binding.start_delete',10),
                                                          ('audit.service_binding.start_create',10),
                                                          ('audit.app.deployment.cancel',10),
                                                          ('audit.app.task.cancel',10),
                                                          ('audit.service_key.start_delete',10),
                                                          ('audit.space.update',10),
                                                          ('audit.app.ssh-unauthorized',10),
                                                          ('audit.service_key.start_create',10),
                                                          ('audit.app.process.terminate_instance',1),
                                                          ('audit.app.droplet.download',1),
                                                          ('audit.service_dashboard_client.create',1),
                                                          ('audit.service_dashboard_client.delete',1),
                                                          ('audit.service_route_binding.delete',1),
                                                          ('audit.service_route_binding.create',1),
                                                          ('blob.remove_orphan',1);


END;
$$ LANGUAGE plpgsql;

-- ============================================================= --

-- FUNC DEF:
CREATE OR REPLACE FUNCTION create_events(
) RETURNS void AS
$$
DECLARE
    event_type text;
    amount int;
    num_events int;
    org_guid text;
    space_guid text;
    org_name_query text := '{{.Prefix}}-%';
    space_name_query text := '{{.Prefix}}-space-%';
    events_guid text;
    events_actor_prefix text := '{{.Prefix}}-events-actor-';
    events_actor_type_prefix text := '{{.Prefix}}-events-actor-type-';
    events_actee_prefix text := '{{.Prefix}}-events-actee-';
    events_actee_type_prefix text := '{{.Prefix}}-events-actee-type-';
BEGIN
    FOR event_type, num_events IN (SELECT audit_event_type, count_events FROM event_types) LOOP
        FOR amount IN 1..num_events LOOP
            events_guid := gen_random_uuid();
            SELECT guid FROM organizations WHERE name LIKE org_name_query ORDER BY random() LIMIT 1 INTO org_guid;
            SELECT guid FROM spaces WHERE name LIKE space_name_query ORDER BY random() LIMIT 1 INTO space_guid;
            INSERT INTO events (guid, "timestamp", "type", actor, actor_type, actee, actee_type, organization_guid, space_guid)
            VALUES (org_name_query || events_guid, current_timestamp, event_type, events_actor_prefix || events_guid, events_actor_type_prefix || events_guid,
                    events_actee_prefix || events_guid, events_actee_type_prefix || events_guid, org_guid, space_guid);
        END LOOP;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- ============================================================= --

-- FUNC DEF:
CREATE OR REPLACE FUNCTION create_users_with_org_and_space_roles(
    org_guid TEXT,
    space_guid TEXT,
    num_users INTEGER
) RETURNS void AS
$$
DECLARE
    org_name_prefix text := '{{.Prefix}}-org-';
    default_quota_definition_id int := 1;
    org_id int;
    space_name_prefix text := '{{.Prefix}}-space-';
    space_id int;
    user_guid text;
    active BOOLEAN := true;
    user_id int;
BEGIN
    INSERT INTO organizations (guid, name, quota_definition_id) VALUES (org_guid, org_name_prefix || org_guid, default_quota_definition_id) RETURNING id INTO org_id;
    INSERT INTO spaces (guid, name, organization_id) VALUES (space_guid, space_name_prefix || space_guid, org_id) RETURNING id INTO space_id;

    FOR _ IN 1..num_users LOOP
        user_guid := gen_random_uuid();
        INSERT INTO users (guid, default_space_id, active) VALUES (user_guid, space_id, active) RETURNING id INTO user_id;

        INSERT INTO organizations_managers (organization_id, user_id) VALUES (org_id, user_id);
        INSERT INTO organizations_billing_managers (organization_id, user_id) VALUES (org_id, user_id);
        INSERT INTO organizations_auditors (organization_id, user_id) VALUES (org_id, user_id);
        INSERT INTO organizations_users (organization_id, user_id) VALUES (org_id, user_id);

        INSERT INTO spaces_managers (space_id, user_id) VALUES (space_id, user_id);
        INSERT INTO spaces_developers (space_id, user_id) VALUES (space_id, user_id);
        INSERT INTO spaces_supporters (space_id, user_id) VALUES (space_id, user_id);
        INSERT INTO spaces_auditors (space_id, user_id) VALUES (space_id, user_id);
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- ============================================================= --

-- FUNC DEF:
CREATE OR REPLACE FUNCTION create_routes_and_route_mappings_for_app(
    app_guid TEXT,
    org_name TEXT,
    space_guid TEXT,
    num_route_mappings INTEGER
) RETURNS void AS
$$
DECLARE
    default_domain_id int := 1;
    quota_id int;
    space_id int;
    route_guid text;
    route_mapping_guid text;
    process_type text := 'web';
    host_prefix text := '{{.Prefix}}-';
    shortened_route_guid text;
BEGIN
    SELECT quota_definition_id INTO quota_id FROM organizations WHERE name = org_name;
    UPDATE quota_definitions SET total_routes = -1 WHERE id = quota_id;

    SELECT id INTO space_id FROM spaces WHERE guid = space_guid;

    FOR _ IN 1..num_route_mappings LOOP
        route_guid := gen_random_uuid();
--      shorten guid to be able to map more routes to the app (diego limitation)
        shortened_route_guid := substring(route_guid FROM 1 FOR 13);
        INSERT INTO routes (guid, domain_id, space_id, host) VALUES (route_guid, default_domain_id, space_id, host_prefix || shortened_route_guid);

        route_mapping_guid := gen_random_uuid();
        INSERT INTO route_mappings (guid, app_guid, route_guid, process_type) VALUES (route_mapping_guid, app_guid, route_guid, process_type);
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- ============================================================= --

-- FUNC DEF:
CREATE OR REPLACE FUNCTION create_service_instances_for_orgs_spaces_plans(
    orgs INTEGER,
    spacesPerOrg INTEGER,
    servicePlans INTEGER,
    instancesPerPlanPerSpace INTEGER,
    namePrefix VARCHAR
) RETURNS void AS $$
DECLARE
    i INT := 0;
    j INT := 0;
    k INT := 0;
    spaceOffset INT;
    servicePlanOffset INT;
    spaceId INT;
    servicePlanId INT;
BEGIN
    -- Loop through organizations
    WHILE i < orgs LOOP
        j := 0;
        
        -- Loop through spaces per organization
        WHILE j < spacesPerOrg LOOP
            -- Calculate spaceOffset
            spaceOffset := i * spacesPerOrg + j;

            -- Get spaceId
            SELECT id INTO spaceId
            FROM spaces
            WHERE name LIKE CONCAT(namePrefix, '-space-%')
            LIMIT 1 OFFSET spaceOffset;

            k := 0;
            -- Loop through service plans
            WHILE k < servicePlans LOOP
                -- Calculate servicePlanOffset
                servicePlanOffset := k;

                -- Get servicePlanId
                SELECT id INTO servicePlanId
                FROM service_plans
                WHERE name LIKE CONCAT(namePrefix, '-service-plan-%')
                LIMIT 1 OFFSET servicePlanOffset;

                -- Call the stored procedure to create service instances
                PERFORM create_service_instances(spaceId, servicePlanId, instancesPerPlanPerSpace);

                k := k + 1;
            END LOOP;

            j := j + 1;
        END LOOP;

        i := i + 1;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- ============================================================= --

-- FUNC DEF:
CREATE OR REPLACE FUNCTION create_service_instance_shares(
    orgs INT, 
    spacesPerOrg INT, 
    serviceInstanceSharesPerSpace INT, 
    namePrefix VARCHAR
) RETURNS void AS $$
DECLARE
    i INT DEFAULT 0;
    j INT DEFAULT 0;
    k INT DEFAULT 0;
    spaceOffset INT;
    spaceId INT;
    shareSpaceGuid VARCHAR(255);
    serviceInstanceGuid VARCHAR(255);
BEGIN
    -- Loop through organizations
    WHILE i < orgs LOOP
        j := 0;
        
        -- Loop through spaces per organization
        WHILE j < spacesPerOrg LOOP
            -- Calculate spaceOffset and get spaceId
            spaceOffset := i * spacesPerOrg + j;

            -- Get spaceId
            SELECT id INTO spaceId
            FROM spaces
            WHERE name LIKE CONCAT(namePrefix, '-space-%')
            LIMIT 1 OFFSET spaceOffset;

            k := 0;
            -- Loop through the service instance shares
            WHILE k < serviceInstanceSharesPerSpace LOOP
                -- Find a random space to share that isn't our space
                SELECT guid INTO shareSpaceGuid
                FROM spaces
                WHERE name LIKE CONCAT(namePrefix, '-space-%')
                AND id != spaceId
                ORDER BY RANDOM()
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

                k := k + 1;
            END LOOP;

            j := j + 1;
        END LOOP;

        i := i + 1;
    END LOOP;
END;
$$ LANGUAGE plpgsql;
