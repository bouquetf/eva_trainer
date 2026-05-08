DELETE FROM role_permissions WHERE permission_id IN (SELECT id FROM permissions);
DELETE FROM permissions;
