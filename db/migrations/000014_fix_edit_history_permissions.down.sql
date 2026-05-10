-- Remove edit_history:view from editor role
DELETE FROM role_permissions
WHERE role = 'editor'
  AND permission_id = (SELECT id FROM permissions WHERE feature = 'edit_history:view');

-- Restore edit_history:view to member role
INSERT INTO role_permissions (role, permission_id)
SELECT 'member', id FROM permissions WHERE feature = 'edit_history:view'
ON CONFLICT (role, permission_id) DO NOTHING;
