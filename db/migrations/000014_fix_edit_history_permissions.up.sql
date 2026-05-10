-- Remove edit_history:view from member role
DELETE FROM role_permissions
WHERE role = 'member'
  AND permission_id = (SELECT id FROM permissions WHERE feature = 'edit_history:view');

-- Add edit_history:view to editor role
INSERT INTO role_permissions (role, permission_id)
SELECT 'editor', id FROM permissions WHERE feature = 'edit_history:view'
ON CONFLICT (role, permission_id) DO NOTHING;
