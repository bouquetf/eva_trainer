DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE feature IN ('line_of_fire:view', 'line_of_fire:create', 'line_of_fire:edit', 'line_of_fire:delete')
);

DELETE FROM permissions WHERE feature IN ('line_of_fire:view', 'line_of_fire:create', 'line_of_fire:edit', 'line_of_fire:delete');
