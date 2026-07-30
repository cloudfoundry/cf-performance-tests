CREATE PROCEDURE assign_user_disjoint_org_roles(
    IN user_guid VARCHAR(255)
)
BEGIN
    DECLARE v_user_id INT;
    SELECT id FROM users WHERE guid = user_guid INTO v_user_id;

    INSERT INTO organizations_managers (organization_id, user_id)
    SELECT id, v_user_id FROM (
        SELECT id, ROW_NUMBER() OVER (ORDER BY id) - 1 AS rn FROM selected_orgs
    ) s WHERE s.rn % 4 = 0;

    INSERT INTO organizations_billing_managers (organization_id, user_id)
    SELECT id, v_user_id FROM (
        SELECT id, ROW_NUMBER() OVER (ORDER BY id) - 1 AS rn FROM selected_orgs
    ) s WHERE s.rn % 4 = 1;

    INSERT INTO organizations_auditors (organization_id, user_id)
    SELECT id, v_user_id FROM (
        SELECT id, ROW_NUMBER() OVER (ORDER BY id) - 1 AS rn FROM selected_orgs
    ) s WHERE s.rn % 4 = 2;

    INSERT INTO organizations_users (organization_id, user_id)
    SELECT id, v_user_id FROM (
        SELECT id, ROW_NUMBER() OVER (ORDER BY id) - 1 AS rn FROM selected_orgs
    ) s WHERE s.rn % 4 = 3;
END;
