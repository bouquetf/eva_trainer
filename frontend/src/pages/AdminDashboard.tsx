import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  listAdminUsers,
  listAdminPermissions,
  getUserDetail,
  getRolePermissions,
  listAdminTeams,
  toggleAdmin,
  updateUserRole,
  addUserToTeam,
  removeUserFromTeam,
} from '../api/admin'
import type { Team } from '../api/admin'
import { useAuthStore } from '../store/auth'
import '../styles/admin.css'

type Tab = 'users' | 'permissions'

const VALID_ROLES = ['member', 'editor', 'cocaptain', 'captain']

function AdminDashboard() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const currentUser = useAuthStore((s) => s.user)
  const [tab, setTab] = useState<Tab>('users')
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  const [expandedUserId, setExpandedUserId] = useState<number | null>(null)
  const [selectedRole, setSelectedRole] = useState<string | null>(null)

  const [addTeamId, setAddTeamId] = useState<number>(0)
  const [addRole, setAddRole] = useState('member')
  const [showAddForm, setShowAddForm] = useState(false)

  const { data: usersData, isLoading: usersLoading } = useQuery({
    queryKey: ['admin-users'],
    queryFn: listAdminUsers,
  })

  const { data: permsData, isLoading: permsLoading } = useQuery({
    queryKey: ['admin-permissions'],
    queryFn: listAdminPermissions,
  })

  const { data: userDetail, isLoading: detailLoading } = useQuery({
    queryKey: ['admin-user-detail', expandedUserId],
    queryFn: () => getUserDetail(expandedUserId!),
    enabled: expandedUserId !== null,
  })

  const { data: allTeams } = useQuery({
    queryKey: ['admin-teams'],
    queryFn: listAdminTeams,
  })

  const { data: rolePerms } = useQuery({
    queryKey: ['admin-role-permissions', selectedRole],
    queryFn: () => getRolePermissions(selectedRole!),
    enabled: selectedRole !== null,
  })

  const toggleMutation = useMutation({
    mutationFn: ({ userId, isAdmin }: { userId: number; isAdmin: boolean }) =>
      toggleAdmin(userId, isAdmin),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] })
      setError(null)
    },
    onError: () => setError('Impossible de modifier le statut administrateur'),
  })

  const updateRoleMutation = useMutation({
    mutationFn: ({ userId, teamId, role }: { userId: number; teamId: number; role: string }) =>
      updateUserRole(userId, teamId, role),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-user-detail', expandedUserId] })
      setSuccess('Rôle mis à jour')
      setError(null)
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? 'Erreur lors de la mise à jour')
    },
  })

  const addMutation = useMutation({
    mutationFn: ({ userId, teamId, role }: { userId: number; teamId: number; role: string }) =>
      addUserToTeam(userId, teamId, role),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-user-detail', expandedUserId] })
      setAddTeamId(0)
      setAddRole('member')
      setShowAddForm(false)
      setSuccess('Membre ajouté')
      setError(null)
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? "Erreur lors de l'ajout")
    },
  })

  const removeMutation = useMutation({
    mutationFn: ({ userId, teamId }: { userId: number; teamId: number }) =>
      removeUserFromTeam(userId, teamId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-user-detail', expandedUserId] })
      setSuccess('Membre retiré')
      setError(null)
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? 'Erreur lors du retrait')
    },
  })

  const users = usersData?.users ?? []
  const permissions = permsData?.permissions ?? []
  const roles = permsData?.roles ?? []
  const memberships = userDetail?.memberships ?? []
  const teams = allTeams?.teams ?? []

  const availableTeams = teams.filter(
    (t) => !memberships.some((m) => m.team_id === t.id)
  )

  const handleToggleExpand = (userId: number) => {
    setExpandedUserId(expandedUserId === userId ? null : userId)
    setShowAddForm(false)
    setError(null)
    setSuccess(null)
  }

  return (
    <div className="admin-page">
      <div className="admin-header">
        <h1>Administration</h1>
        <button className="btn-secondary" onClick={() => navigate('/')}>
          Retour à l'accueil
        </button>
      </div>

      <div className="admin-tabs">
        <button className={tab === 'users' ? 'active' : ''} onClick={() => { setTab('users'); setSelectedRole(null) }}>
          Utilisateurs ({users.length})
        </button>
        <button className={tab === 'permissions' ? 'active' : ''} onClick={() => { setTab('permissions'); setExpandedUserId(null) }}>
          Permissions
        </button>
      </div>

      {error && <div className="error-message">{error}</div>}
      {success && <div className="success-message">{success}</div>}

      {tab === 'users' && (
        <div className="section">
          <h2 className="section-title">Tous les utilisateurs</h2>
          {usersLoading ? (
            <div className="empty-state"><p>Chargement...</p></div>
          ) : users.length === 0 ? (
            <div className="empty-state"><p>Aucun utilisateur</p></div>
          ) : (
            <table className="admin-table">
              <thead>
                <tr>
                  <th></th>
                  <th>ID</th>
                  <th>Discord</th>
                  <th>Nom d'affichage</th>
                  <th>Statut</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {users.map((u) => {
                  const isSelf = u.id === currentUser?.id
                  const isExpanded = expandedUserId === u.id
                  return (
                    <>
                      <tr
                        key={u.id}
                        className={isExpanded ? 'row-expanded' : ''}
                        onClick={() => handleToggleExpand(u.id)}
                      >
                        <td className="expand-cell">{isExpanded ? '▾' : '▸'}</td>
                        <td>{u.id}</td>
                        <td>{u.discord_username || '-'}</td>
                        <td>{u.display_name || '-'}</td>
                        <td>
                          {u.is_admin ? (
                            <span className="role-badge admin">Admin</span>
                          ) : (
                            <span className="role-badge member">Utilisateur</span>
                          )}
                        </td>
                        <td onClick={(e) => e.stopPropagation()}>
                          {isSelf ? (
                            <span className="self-label">Vous</span>
                          ) : (
                            <button
                              className="btn-secondary btn-small"
                              onClick={() => toggleMutation.mutate({ userId: u.id, isAdmin: !u.is_admin })}
                              disabled={toggleMutation.isPending}
                            >
                              {u.is_admin ? 'Retirer admin' : 'Promouvoir admin'}
                            </button>
                          )}
                        </td>
                      </tr>
                      {isExpanded && (
                        <tr key={`${u.id}-detail`}>
                          <td colSpan={6} className="detail-cell">
                            <div className="user-detail">
                              {detailLoading ? (
                                <div className="empty-state"><p>Chargement...</p></div>
                              ) : (
                                <>
                                  <h4>Équipes et rôles</h4>
                                  {memberships.length === 0 ? (
                                    <p className="text-muted">Aucune équipe</p>
                                  ) : (
                                    <table className="admin-table memberships-table">
                                      <thead>
                                        <tr>
                                          <th>Équipe</th>
                                          <th>Rôle</th>
                                          <th>Actions</th>
                                        </tr>
                                      </thead>
                                      <tbody>
                                        {memberships.map((m) => (
                                          <tr key={`${m.team_id}`}>
                                            <td>{m.team_name}</td>
                                            <td>
                                              <select
                                                className="role-select"
                                                value={m.role}
                                                onChange={(e) =>
                                                  updateRoleMutation.mutate({
                                                    userId: u.id,
                                                    teamId: m.team_id,
                                                    role: e.target.value,
                                                  })
                                                }
                                              >
                                                {VALID_ROLES.map((r) => (
                                                  <option key={r} value={r}>{r}</option>
                                                ))}
                                              </select>
                                            </td>
                                            <td>
                                              {m.role !== 'captain' && (
                                                <button
                                                  className="btn-small btn-danger"
                                                  onClick={() =>
                                                    removeMutation.mutate({
                                                      userId: u.id,
                                                      teamId: m.team_id,
                                                    })
                                                  }
                                                  disabled={removeMutation.isPending}
                                                >
                                                  Retirer
                                                </button>
                                              )}
                                              {m.role === 'captain' && (
                                                <span className="self-label">Capitaine</span>
                                              )}
                                            </td>
                                          </tr>
                                        ))}
                                      </tbody>
                                    </table>
                                  )}

                                  {showAddForm ? (
                                    <div className="add-membership-form">
                                      <h4>Ajouter à une équipe</h4>
                                      <div className="add-membership-row">
                                        <select
                                          value={addTeamId}
                                          onChange={(e) => setAddTeamId(Number(e.target.value))}
                                        >
                                          <option value={0}>Choisir une équipe</option>
                                          {availableTeams.map((t: Team) => (
                                            <option key={t.id} value={t.id}>{t.name}</option>
                                          ))}
                                        </select>
                                        <select value={addRole} onChange={(e) => setAddRole(e.target.value)}>
                                          {VALID_ROLES.map((r) => (
                                            <option key={r} value={r}>{r}</option>
                                          ))}
                                        </select>
                                        <button
                                          className="btn-primary btn-small"
                                          onClick={() =>
                                            addMutation.mutate({
                                              userId: u.id,
                                              teamId: addTeamId,
                                              role: addRole,
                                            })
                                          }
                                          disabled={addTeamId === 0 || addMutation.isPending}
                                        >
                                          Ajouter
                                        </button>
                                        <button className="btn-secondary btn-small" onClick={() => setShowAddForm(false)}>
                                          Annuler
                                        </button>
                                      </div>
                                    </div>
                                  ) : (
                                    availableTeams.length > 0 && (
                                      <button className="btn-secondary btn-small" onClick={() => setShowAddForm(true)} style={{ marginTop: 12 }}>
                                        + Ajouter à une équipe
                                      </button>
                                    )
                                  )}
                                </>
                              )}
                            </div>
                          </td>
                        </tr>
                      )}
                    </>
                  )
                })}
              </tbody>
            </table>
          )}
        </div>
      )}

      {tab === 'permissions' && (
        <div className="section">
          {selectedRole ? (
            <>
              <div className="perm-header">
                <button className="btn-secondary btn-small" onClick={() => setSelectedRole(null)}>
                  ← Retour
                </button>
                <h2 className="section-title" style={{ margin: 0 }}>Rôle : {selectedRole}</h2>
              </div>
              {rolePerms ? (
                <table className="admin-table" style={{ marginTop: 16 }}>
                  <thead>
                    <tr>
                      <th>Permission</th>
                      <th>Description</th>
                    </tr>
                  </thead>
                  <tbody>
                    {rolePerms.permissions.length === 0 ? (
                      <tr><td colSpan={2} className="empty-state"><p>Aucune permission pour ce rôle</p></td></tr>
                    ) : (
                      rolePerms.permissions.map((p) => (
                        <tr key={p.id}>
                          <td><code>{p.feature}</code></td>
                          <td>{p.description}</td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              ) : (
                <div className="empty-state"><p>Chargement...</p></div>
              )}
            </>
          ) : (
            <>
              <h2 className="section-title">Permissions par rôle</h2>
              {permsLoading ? (
                <div className="empty-state"><p>Chargement...</p></div>
              ) : permissions.length === 0 ? (
                <div className="empty-state"><p>Aucune permission</p></div>
              ) : (
                <>
                  <div className="perm-role-list">
                    {roles.map((r) => (
                      <button key={r} className="perm-role-btn" onClick={() => setSelectedRole(r)}>
                        {r}
                      </button>
                    ))}
                  </div>
                  <div className="perm-matrix" style={{ marginTop: 16 }}>
                    <table className="admin-table">
                      <thead>
                        <tr>
                          <th>Permission</th>
                          <th>Description</th>
                          {roles.map((r) => (
                            <th key={r}>{r}</th>
                          ))}
                        </tr>
                      </thead>
                      <tbody>
                        {permissions.map((p) => (
                          <tr key={p.id}>
                            <td><code>{p.feature}</code></td>
                            <td>{p.description}</td>
                            {roles.map((r) => (
                              <td key={r} className="perm-cell">
                                {p.roles.includes(r) ? (
                                  <span className="perm-check">✓</span>
                                ) : (
                                  <span className="perm-cross">-</span>
                                )}
                              </td>
                            ))}
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </>
              )}
            </>
          )}
        </div>
      )}
    </div>
  )
}

export default AdminDashboard
