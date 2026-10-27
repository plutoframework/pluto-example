-- Synthetic, local-only fixtures. Do not run on an existing application database.
BEGIN;
INSERT INTO organizations (id, name) VALUES
    ('northstar', 'Northstar Research'),
    ('harbor', 'Harbor Studio');
INSERT INTO memberships (tenant_id, actor_id, display_name, role) VALUES
    ('northstar', 'alice', 'Alice Chen', 'member'),
    ('northstar', 'bob', 'Bob Rivera', 'approver'),
    ('northstar', 'carol', 'Carol Okafor', 'administrator'),
    ('harbor', 'dana', 'Dana Patel', 'administrator'),
    ('harbor', 'alice', 'Alice Chen', 'member');
COMMIT;
