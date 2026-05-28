INSERT INTO permissions (feature, description) VALUES
    ('line_of_fire:view', 'View lines of fire of a team'),
    ('line_of_fire:create', 'Create new lines of fire'),
    ('line_of_fire:edit', 'Modify existing lines of fire (cones, name)'),
    ('line_of_fire:delete', 'Delete lines of fire')
ON CONFLICT (feature) DO NOTHING;

INSERT INTO role_permissions (role, permission_id)
SELECT 'member', id FROM permissions WHERE feature = 'line_of_fire:view'
ON CONFLICT (role, permission_id) DO NOTHING;

INSERT INTO role_permissions (role, permission_id)
SELECT 'editor', id FROM permissions WHERE feature IN ('line_of_fire:view', 'line_of_fire:create', 'line_of_fire:edit', 'line_of_fire:delete')
ON CONFLICT (role, permission_id) DO NOTHING;

INSERT INTO role_permissions (role, permission_id)
SELECT 'cocaptain', id FROM permissions WHERE feature IN ('line_of_fire:view', 'line_of_fire:create', 'line_of_fire:edit', 'line_of_fire:delete')
ON CONFLICT (role, permission_id) DO NOTHING;

INSERT INTO role_permissions (role, permission_id)
SELECT 'captain', id FROM permissions WHERE feature IN ('line_of_fire:view', 'line_of_fire:create', 'line_of_fire:edit', 'line_of_fire:delete')
ON CONFLICT (role, permission_id) DO NOTHING;
