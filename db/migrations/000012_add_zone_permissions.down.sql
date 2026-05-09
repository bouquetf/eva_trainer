DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE feature IN ('zone:view', 'zone:create', 'zone:edit', 'zone:delete')
);

DELETE FROM permissions WHERE feature IN ('zone:view', 'zone:create', 'zone:edit', 'zone:delete');
