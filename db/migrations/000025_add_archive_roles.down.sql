DROP TABLE IF EXISTS global_roles;
DROP TABLE IF EXISTS team_roles;

DELETE FROM permissions WHERE feature IN ('archive:read', 'archive:write');
