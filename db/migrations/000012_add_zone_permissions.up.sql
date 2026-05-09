INSERT INTO permissions (feature, description) VALUES
    ('zone:view', 'View zones of a team'),
    ('zone:create', 'Create new zones'),
    ('zone:edit', 'Modify existing zones (polygons, points, name)'),
    ('zone:delete', 'Delete zones'),
    ('zone:lock', 'Lock/unlock a map for editing')
ON CONFLICT (feature) DO NOTHING;

INSERT INTO role_permissions (role, permission_id)
SELECT 'member', id FROM permissions WHERE feature = 'zone:view'
ON CONFLICT (role, permission_id) DO NOTHING;

INSERT INTO role_permissions (role, permission_id)
SELECT 'editor', id FROM permissions WHERE feature IN ('zone:view', 'zone:create', 'zone:edit', 'zone:delete', 'zone:lock')
ON CONFLICT (role, permission_id) DO NOTHING;

INSERT INTO role_permissions (role, permission_id)
SELECT 'cocaptain', id FROM permissions WHERE feature IN ('zone:view', 'zone:create', 'zone:edit', 'zone:delete', 'zone:lock')
ON CONFLICT (role, permission_id) DO NOTHING;

INSERT INTO role_permissions (role, permission_id)
SELECT 'captain', id FROM permissions WHERE feature IN ('zone:view', 'zone:create', 'zone:edit', 'zone:delete', 'zone:lock')
ON CONFLICT (role, permission_id) DO NOTHING;
