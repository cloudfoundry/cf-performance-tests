CREATE PROCEDURE share_private_domains_with_orgs(num_orgs_per_domain INT)
BEGIN
    DECLARE done INT DEFAULT 0;
    DECLARE private_domain_id INT;
    DECLARE owning_org_id INT;
    DECLARE domains_cursor CURSOR FOR SELECT id FROM domains WHERE name LIKE '{{.Prefix}}-private-domain-%';
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = 1;

    OPEN domains_cursor;
    domain_loop:
    LOOP
        FETCH domains_cursor INTO private_domain_id;
        IF done = 1 THEN
            LEAVE domain_loop;
        END IF;

        SELECT owning_organization_id INTO owning_org_id FROM domains WHERE id = private_domain_id;

        INSERT IGNORE INTO organizations_private_domains (organization_id, private_domain_id)
        SELECT o.id, private_domain_id
        FROM organizations o
        WHERE o.name LIKE '{{.Prefix}}-org-%'
          AND o.id <> owning_org_id
        ORDER BY RAND()
        LIMIT num_orgs_per_domain;
    END LOOP;
    CLOSE domains_cursor;
END;
