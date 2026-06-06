import { useState, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  listAdminUsers,
  listAdminPermissions,
  getUserDetail,
  getRolePermissions,
  listAdminTeams,
  promoteUser,
  updateUserRole,
  addUserToTeam,
  removeUserFromTeam,
  exportTeamMap,
  exportAllTeamMaps,
  importTeamZones,
  listArchivists,
  addArchivist,
  removeArchivist,
  listReaderTeams,
  addReaderTeam,
  removeReaderTeam,
} from '../api/admin'
import type { ImportMapPayload } from '../api/admin'
import type { Team } from '../api/admin'
import { uploadMap, listAdminMaps, purgeObsoleteMaps, getMapImageUrl } from '../api/maps'

import {
  createSplit,
  listSplits,
  getSplitMaps,
  updateSplit,
  deleteSplit,
} from '../api/splits'
import type { Split } from '../api/splits'
import type { Map as EvaMap } from '../api/maps'
import {
  listCatalog,
  createCatalogEntry,
  updateCatalogEntry,
  deleteCatalogEntry,
} from '../api/map-catalog'
import type { MapCatalogEntry } from '../api/map-catalog'
import { searchUsers } from '../api/invitations'
import { useAuthStore } from '../store/auth'
import AdminMenu from '../components/AdminMenu'
import AdminLiguesTab from '../components/AdminLiguesTab'
import type { AdminTab } from '../components/AdminMenu'
import '../styles/admin.css'

const VALID_ROLES = ['member', 'editor', 'cocaptain', 'captain']

function AdminDashboard() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const currentUser = useAuthStore((s) => s.user)
  const [tab, setTab] = useState<AdminTab>('users')
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  const [expandedUserId, setExpandedUserId] = useState<number | null>(null)
  const [selectedRole, setSelectedRole] = useState<string | null>(null)

  const [addTeamId, setAddTeamId] = useState<number>(0)
  const [addRole, setAddRole] = useState('member')
  const [showAddForm, setShowAddForm] = useState(false)

  const [mapName, setMapName] = useState('')
  const [mapFile, setMapFile] = useState<File | null>(null)
  const [showPurgeModal, setShowPurgeModal] = useState(false)

  const [splitName, setSplitName] = useState('')
  const [selectedMapNames, setSelectedMapNames] = useState<string[]>([])
  const [editingSplit, setEditingSplit] = useState<Split | null>(null)
  const [expandedSplitId, setExpandedSplitId] = useState<number | null>(null)
  const [splitMaps, setSplitMaps] = useState<Record<number, EvaMap[]>>({})

  const [newCatalogName, setNewCatalogName] = useState('')
  const [editingCatalogEntry, setEditingCatalogEntry] = useState<MapCatalogEntry | null>(null)

  const [exportTeamId, setExportTeamId] = useState<number>(0)
  const [exportMapId, setExportMapId] = useState<number>(0)
  const [importTeamId, setImportTeamId] = useState<number>(0)
  const [importFile, setImportFile] = useState<File | null>(null)

  const [archivistSearch, setArchivistSearch] = useState('')
  const [archivistSearchResults, setArchivistSearchResults] = useState<{ id: number; display_name: string; discord_username: string }[]>([])
  const [showArchivistSuggestions, setShowArchivistSuggestions] = useState(false)
  const [selectedArchivistUserId, setSelectedArchivistUserId] = useState<number | null>(null)
  const [readerTeamId, setReaderTeamId] = useState<number>(0)
  const archivistSearchTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const { data: archivistsData, isLoading: archivistsLoading } = useQuery({
    queryKey: ['admin-archivists'],
    queryFn: listArchivists,
  })

  const { data: readerTeamsData, isLoading: readerTeamsLoading } = useQuery({
    queryKey: ['admin-reader-teams'],
    queryFn: listReaderTeams,
  })

  const archivists = archivistsData?.archivists ?? []
  const readerTeams = readerTeamsData?.teams ?? []

  const addArchivistMutation = useMutation({
    mutationFn: (userId: number) => addArchivist(userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-archivists'] })
      setArchivistSearch('')
      setArchivistSearchResults([])
      setSelectedArchivistUserId(null)
      setError(null)
      setSuccess('Archiviste ajouté')
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? "Erreur lors de l'ajout de l'archiviste")
    },
  })

  const removeArchivistMutation = useMutation({
    mutationFn: (userId: number) => removeArchivist(userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-archivists'] })
      setError(null)
      setSuccess('Archiviste retiré')
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? 'Erreur lors du retrait de l\'archiviste')
    },
  })

  const addReaderTeamMutation = useMutation({
    mutationFn: (teamId: number) => addReaderTeam(teamId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-reader-teams'] })
      setReaderTeamId(0)
      setError(null)
      setSuccess('Équipe lectrice ajoutée')
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? "Erreur lors de l'ajout de l'équipe lectrice")
    },
  })

  const removeReaderTeamMutation = useMutation({
    mutationFn: (teamId: number) => removeReaderTeam(teamId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-reader-teams'] })
      setError(null)
      setSuccess('Équipe lectrice retirée')
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? 'Erreur lors du retrait de l\'équipe lectrice')
    },
  })

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

  const promoteMutation = useMutation({
    mutationFn: (userId: number) => promoteUser(userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] })
      setSuccess('Utilisateur promu administrateur')
      setError(null)
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? "Impossible de promouvoir l'utilisateur")
    },
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

  const { data: mapsData, isLoading: mapsLoading } = useQuery({
    queryKey: ['admin-maps'],
    queryFn: listAdminMaps,
  })

  const uploadMutation = useMutation({
    mutationFn: ({ name, file }: { name: string; file: File }) => uploadMap(name, file),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['admin-maps'] })
      setMapName('')
      setMapFile(null)
      const msg = `Carte ${data.map.name} uploadée avec succès`
      const extra = data.previous_version_obsoleted
        ? ` L'ancienne version de ${data.previous_version_obsoleted.name} a été marquée obsolète`
        : ''
      setSuccess(msg + extra)
      setError(null)
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? "Erreur lors de l'upload")
    },
  })

  const purgeMutation = useMutation({
    mutationFn: purgeObsoleteMaps,
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['admin-maps'] })
      setShowPurgeModal(false)
      const count = data.purged_ids.length
      setSuccess(count > 0 ? `${count} carte(s) purgée(s) avec succès` : 'Aucune carte obsolète à purger')
      setError(null)
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? 'Erreur lors de la purge')
      setShowPurgeModal(false)
    },
  })

  const createSplitMutation = useMutation({
    mutationFn: (data: { name: string; map_names: string[] }) =>
      createSplit(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-splits'] })
      setSplitName('')
      setSelectedMapNames([])
      setSuccess('Split créé avec succès')
      setError(null)
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? 'Erreur lors de la création du split')
    },
  })

  const updateSplitMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: { name: string; map_names: string[] } }) =>
      updateSplit(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-splits'] })
      setEditingSplit(null)
      setSplitName('')
      setSelectedMapNames([])
      setSuccess('Split mis à jour avec succès')
      setError(null)
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? 'Erreur lors de la mise à jour du split')
    },
  })

  const deleteSplitMutation = useMutation({
    mutationFn: (id: number) => deleteSplit(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-splits'] })
      setSuccess('Split supprimé avec succès')
      setError(null)
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? 'Erreur lors de la suppression du split')
    },
  })

  const { data: catalogData, isLoading: catalogLoading } = useQuery({
    queryKey: ['admin-map-catalog'],
    queryFn: listCatalog,
  })

  const catalogEntries = catalogData?.entries ?? []

  const createCatalogMutation = useMutation({
    mutationFn: (name: string) => createCatalogEntry(name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-map-catalog'] })
      setNewCatalogName('')
      setSuccess('Carte ajoutée au catalogue')
      setError(null)
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? "Erreur lors de l'ajout")
    },
  })

  const updateCatalogMutation = useMutation({
    mutationFn: ({ id, name }: { id: number; name: string }) => updateCatalogEntry(id, name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-map-catalog'] })
      setEditingCatalogEntry(null)
      setSuccess('Nom de carte mis à jour')
      setError(null)
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? 'Erreur lors de la mise à jour')
    },
  })

  const importMutation = useMutation({
    mutationFn: ({ teamId, data }: { teamId: number; data: ImportMapPayload[] }) =>
      importTeamZones(teamId, data),
    onSuccess: (result) => {
      let msg = `${result.imported_count} zone(s) importée(s) sur ${result.maps_processed} carte(s)`
      if (result.errors && result.errors.length > 0) {
        msg += ` (${result.errors.length} erreur(s))`
      }
      setSuccess(msg)
      setError(null)
      setImportFile(null)
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? "Erreur lors de l'import")
    },
  })

  const deleteCatalogMutation = useMutation({
    mutationFn: (id: number) => deleteCatalogEntry(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-map-catalog'] })
      setSuccess('Carte retirée du catalogue')
      setError(null)
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? 'Erreur lors de la suppression')
    },
  })

  const loadSplitMaps = async (splitId: number) => {
    if (splitMaps[splitId]) return
    try {
      const resp = await getSplitMaps(splitId)
      setSplitMaps((prev) => ({ ...prev, [splitId]: resp.maps }))
    } catch {
      setError('Impossible de charger les cartes du split')
    }
  }

  const maps = mapsData?.maps ?? []
  const obsoleteMapsWithoutZones = maps.filter((m) => m.is_obsolete)

  const { data: splitsData, isLoading: splitsLoading } = useQuery({
    queryKey: ['admin-splits'],
    queryFn: listSplits,
  })

  const splits = splitsData?.splits ?? []

  const users = usersData?.users ?? []
  const permissions = permsData?.permissions ?? []
  const roles = permsData?.roles ?? []
  const memberships = userDetail?.memberships ?? []
  const teams = allTeams?.teams ?? []

  const availableTeams = teams.filter(
    (t) => !memberships.some((m) => m.team_id === t.id)
  )

  function handleTabChange(newTab: AdminTab) {
    setTab(newTab)
    setSelectedRole(null)
    setExpandedUserId(null)
    setEditingSplit(null)
    setError(null)
    setSuccess(null)
  }

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

      <AdminMenu
        activeTab={tab}
        onTabChange={handleTabChange}
        usersCount={users.length}
        mapsCount={maps.length}
        splitsCount={splits.length}
        catalogCount={catalogEntries.length}
      />

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
                            <>
                              {u.id === currentUser?.id ? (
                                <span className="self-label">Vous</span>
                              ) : (
                                <span className="role-badge admin">Admin</span>
                              )}
                            </>
                          ) : (
                            <button
                              className="btn-small btn-primary"
                              onClick={() => {
                                if (window.confirm(`Promouvoir ${u.display_name || u.discord_username || `#${u.id}`} au rôle admin ? Cette action est irréversible.`)) {
                                  promoteMutation.mutate(u.id)
                                }
                              }}
                              disabled={promoteMutation.isPending}
                            >
                              Promouvoir admin
                            </button>
                          )}
                        </td>
                        <td onClick={(e) => e.stopPropagation()}>
                          {isSelf ? (
                            <span className="self-label">Vous</span>
                          ) : u.is_admin ? (
                            <span className="role-badge admin" style={{ fontSize: 13 }}>Admin</span>
                          ) : (
                            <button
                              className="btn-secondary btn-small"
                              onClick={() => promoteMutation.mutate(u.id)}
                              disabled={promoteMutation.isPending}
                            >
                              Promouvoir admin
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
                                                  onClick={() => {
                                                    if (window.confirm(`Retirer ${u.display_name || u.discord_username || `#${u.id}`} de l'équipe "${m.team_name}" ?`)) {
                                                      removeMutation.mutate({
                                                        userId: u.id,
                                                        teamId: m.team_id,
                                                      })
                                                    }
                                                  }}
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

      {tab === 'maps' && (
        <div className="section">
          <h2 className="section-title">Uploader une carte</h2>
          <div className="map-upload-form">
            <div className="map-form-row">
              <div className="map-form-group">
                <label htmlFor="map-name">Nom de la carte</label>
                <select
                  id="map-name"
                  value={mapName}
                  onChange={(e) => setMapName(e.target.value)}
                >
                  <option value="">-- Sélectionner une carte --</option>
                  {catalogEntries.map((e) => (
                    <option key={e.id} value={e.name}>{e.name}</option>
                  ))}
                </select>
              </div>
              <div className="map-form-group">
                <label htmlFor="map-file">Fichier WebP</label>
                <input
                  id="map-file"
                  type="file"
                  accept=".webp,image/webp"
                  onChange={(e) => setMapFile(e.target.files?.[0] ?? null)}
                />
                <span className="map-form-hint">Format accepté : WebP. Taille max : 50 MB</span>
              </div>
              <button
                className="btn-primary"
                onClick={() => {
                  setError(null)
                  setSuccess(null)
                  if (!mapName) { setError('Veuillez sélectionner une carte'); return }
                  if (!mapFile) { setError('Veuillez sélectionner un fichier'); return }
                  if (!mapFile.name.toLowerCase().endsWith('.webp') && mapFile.type !== 'image/webp') {
                    setError('Seul le format WebP est accepté')
                    return
                  }
                  uploadMutation.mutate({ name: mapName, file: mapFile })
                }}
                disabled={uploadMutation.isPending}
              >
                {uploadMutation.isPending ? 'Upload en cours...' : 'Uploader la carte'}
              </button>
            </div>
          </div>
        </div>
      )}

      {tab === 'maps' && (
        <div className="section">
          <h2 className="section-title">Toutes les cartes</h2>
          {mapsLoading ? (
            <div className="empty-state"><p>Chargement...</p></div>
          ) : maps.length === 0 ? (
            <div className="empty-state"><p>Aucune carte pour le moment</p></div>
          ) : (
            <table className="admin-table">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Nom</th>
                  <th>Date d'ajout</th>
                  <th>Statut</th>
                  <th>Remplacée par</th>
                  <th>Image</th>
                </tr>
              </thead>
              <tbody>
                {maps.map((m) => (
                  <tr key={m.id}>
                    <td>{m.id}</td>
                    <td>{m.name}</td>
                    <td>{formatDate(m.added_at)}</td>
                    <td>
                      {m.is_obsolete ? (
                        <span className="map-badge-obsolete" title="Cette carte a été remplacée par une version plus récente">Obsolète</span>
                      ) : (
                        <span className="map-badge-active" title="Carte active utilisée par les zones">Active</span>
                      )}
                    </td>
                    <td>
                      {m.obsolete_of ? (
                        <span>→ ID {m.obsolete_of} ({maps.find(o => o.id === m.obsolete_of)?.name ?? '?'})</span>
                      ) : '-'}
                    </td>
                    <td>
                      <img
                        className="map-thumbnail"
                        src={getMapImageUrl(m.id)}
                        alt={m.name}
                        loading="lazy"
                      />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {tab === 'maps' && obsoleteMapsWithoutZones.length > 0 && (
        <div className="purge-section">
          <p>{obsoleteMapsWithoutZones.length} carte(s) obsolète(s) sans zone associée peuvent être purgées</p>
          <button
            className="btn-danger"
            onClick={() => setShowPurgeModal(true)}
            disabled={purgeMutation.isPending}
          >
            {purgeMutation.isPending ? 'Purge en cours...' : 'Purger les cartes obsolètes'}
          </button>
        </div>
      )}

      {tab === 'splits' && (
        <div className="section">
          <h2 className="section-title">{editingSplit ? 'Modifier le split' : 'Créer un split'}</h2>
          <div className="map-upload-form">
            <div className="map-form-row" style={{ flexDirection: 'column', gap: 12 }}>
              <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap' }}>
                <div className="map-form-group" style={{ flex: 1, minWidth: 200 }}>
                  <label htmlFor="split-name">Nom</label>
                  <input
                    id="split-name"
                    type="text"
                    placeholder="Ex: Season 1"
                    value={splitName}
                    onChange={(e) => setSplitName(e.target.value)}
                  />
                </div>
              </div>
              <div className="map-form-group">
                <label>Cartes incluses dans le split</label>
                <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8, marginTop: 4 }}>
                  {catalogEntries.map((e) => {
                    const checked = selectedMapNames.includes(e.name)
                    return (
                      <label
                        key={e.id}
                        style={{
                          display: 'flex', alignItems: 'center', gap: 4,
                          padding: '4px 8px', borderRadius: 4, cursor: 'pointer',
                          background: checked ? '#e8f5e9' : '#f5f5f5',
                          border: checked ? '1px solid #4caf50' : '1px solid #ddd',
                        }}
                      >
                        <input
                          type="checkbox"
                          checked={checked}
                          onChange={() => {
                            setSelectedMapNames((prev) =>
                              checked ? prev.filter((n) => n !== e.name) : [...prev, e.name]
                            )
                          }}
                        />
                        {e.name}
                      </label>
                    )
                  })}
                  {catalogEntries.length === 0 && (
                    <p className="text-muted">
                      Aucune carte dans le catalogue. Ajoutez des cartes depuis l'onglet "Catalogue".
                    </p>
                  )}
                </div>
              </div>
              <div style={{ display: 'flex', gap: 8 }}>
                <button
                  className="btn-primary"
                  onClick={() => {
                    setError(null)
                    setSuccess(null)
                    if (!splitName.trim()) { setError('Le nom du split est requis'); return }
                    if (selectedMapNames.length === 0) { setError('Au moins une carte doit être sélectionnée'); return }
                    const data: { name: string; map_names: string[] } = {
                      name: splitName.trim(),
                      map_names: selectedMapNames,
                    }
                    if (editingSplit) {
                      updateSplitMutation.mutate({ id: editingSplit.id, data })
                    } else {
                      createSplitMutation.mutate(data)
                    }
                  }}
                  disabled={createSplitMutation.isPending || updateSplitMutation.isPending}
                >
                  {createSplitMutation.isPending || updateSplitMutation.isPending
                    ? 'Enregistrement...'
                    : editingSplit ? 'Enregistrer les modifications' : 'Créer le split'
                  }
                </button>
                {editingSplit && (
                  <button className="btn-secondary" onClick={() => {
                    setEditingSplit(null)
                    setSplitName('')
                    setSelectedMapNames([])
                  }}>
                    Annuler
                  </button>
                )}
              </div>
            </div>
          </div>
        </div>
      )}

      {tab === 'splits' && (
        <div className="section">
          <h2 className="section-title">Tous les splits</h2>
          {splitsLoading ? (
            <div className="empty-state"><p>Chargement...</p></div>
          ) : splits.length === 0 ? (
            <div className="empty-state"><p>Aucun split pour le moment</p></div>
          ) : (
            <table className="admin-table">
              <thead>
                <tr>
                  <th></th>
                  <th>ID</th>
                  <th>Nom</th>
                  <th>Cartes</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {splits.map((s) => {
                  const isExpanded = expandedSplitId === s.id
                  return (
                    <>
                      <tr
                        key={s.id}
                        className={isExpanded ? 'row-expanded' : ''}
                        onClick={() => {
                          if (isExpanded) {
                            setExpandedSplitId(null)
                          } else {
                            setExpandedSplitId(s.id)
                            loadSplitMaps(s.id)
                          }
                        }}
                      >
                        <td className="expand-cell">{isExpanded ? '▾' : '▸'}</td>
                        <td>{s.id}</td>
                        <td>{s.name}</td>
                        <td>{s.map_names.join(', ')}</td>
                        <td onClick={(e) => e.stopPropagation()}>
                          <div style={{ display: 'flex', gap: 8 }}>
                            <button
                              className="btn-secondary btn-small"
                              onClick={() => {
                                setEditingSplit(s)
                                setSplitName(s.name)
                                setSelectedMapNames(s.map_names)
                                setTab('splits')
                              }}
                            >
                              Modifier
                            </button>
                            <button
                              className="btn-small btn-danger"
                              onClick={() => {
                                if (window.confirm(`Supprimer le split "${s.name}" ?`)) {
                                  deleteSplitMutation.mutate(s.id)
                                }
                              }}
                              disabled={deleteSplitMutation.isPending}
                            >
                              Supprimer
                            </button>
                          </div>
                        </td>
                      </tr>
                      {isExpanded && (
                        <tr key={`${s.id}-maps`}>
                          <td colSpan={5} className="detail-cell">
                            <div className="user-detail">
                              <h4>Cartes résolues pour ce split</h4>
                              {splitMaps[s.id] ? (
                                splitMaps[s.id].length === 0 ? (
                                  <p className="text-muted">Aucune carte trouvée pour ce split</p>
                                ) : (
                                  <table className="admin-table" style={{ marginTop: 8 }}>
                                    <thead>
                                      <tr>
                                        <th>ID</th>
                                        <th>Nom</th>
                                        <th>Statut</th>
                                      </tr>
                                    </thead>
                                    <tbody>
                                      {splitMaps[s.id].map((m) => (
                                        <tr key={m.id}>
                                          <td>{m.id}</td>
                                          <td>{m.name}</td>
                                          <td>
                                            {m.is_obsolete ? (
                                              <span className="map-badge-obsolete">Obsolète</span>
                                            ) : (
                                              <span className="map-badge-active">Active</span>
                                            )}
                                          </td>
                                        </tr>
                                      ))}
                                    </tbody>
                                  </table>
                                )
                              ) : (
                                <div className="empty-state"><p>Chargement...</p></div>
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

      {tab === 'catalog' && (
        <div className="section">
          <h2 className="section-title">Gestion du catalogue de cartes</h2>
          <div className="map-upload-form">
            <div className="map-form-row" style={{ gap: 12, alignItems: 'flex-end' }}>
              <div className="map-form-group" style={{ flex: 1 }}>
                <label htmlFor="new-catalog-name">Ajouter une carte</label>
                <input
                  id="new-catalog-name"
                  type="text"
                  placeholder="Ex: Atlantis"
                  value={newCatalogName}
                  onChange={(e) => setNewCatalogName(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' && newCatalogName.trim()) {
                      createCatalogMutation.mutate(newCatalogName.trim())
                    }
                  }}
                />
              </div>
              <button
                className="btn-primary"
                onClick={() => {
                  if (!newCatalogName.trim()) { setError('Le nom est requis'); return }
                  createCatalogMutation.mutate(newCatalogName.trim())
                }}
                disabled={createCatalogMutation.isPending}
              >
                Ajouter
              </button>
            </div>
          </div>
        </div>
      )}

      {tab === 'catalog' && (
        <div className="section">
          <h2 className="section-title">Cartes disponibles ({catalogEntries.length})</h2>
          {catalogLoading ? (
            <div className="empty-state"><p>Chargement...</p></div>
          ) : catalogEntries.length === 0 ? (
            <div className="empty-state"><p>Aucune carte dans le catalogue. Ajoutez des cartes avec le formulaire ci-dessus.</p></div>
          ) : (
            <table className="admin-table">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Nom</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {catalogEntries.map((e) => (
                  <tr key={e.id}>
                    <td>{e.id}</td>
                    <td>
                      {editingCatalogEntry?.id === e.id ? (
                        <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
                          <input
                            type="text"
                            value={editingCatalogEntry.name}
                            onChange={(ev) => setEditingCatalogEntry({ ...editingCatalogEntry, name: ev.target.value })}
                            onKeyDown={(ev) => {
                              if (ev.key === 'Enter' && editingCatalogEntry.name.trim()) {
                                updateCatalogMutation.mutate({ id: e.id, name: editingCatalogEntry.name.trim() })
                              }
                            }}
                          />
                          <button
                            className="btn-primary btn-small"
                            onClick={() => {
                              if (editingCatalogEntry.name.trim()) {
                                updateCatalogMutation.mutate({ id: e.id, name: editingCatalogEntry.name.trim() })
                              }
                            }}
                            disabled={updateCatalogMutation.isPending}
                          >
                            OK
                          </button>
                          <button className="btn-secondary btn-small" onClick={() => setEditingCatalogEntry(null)}>
                            Annuler
                          </button>
                        </div>
                      ) : (
                        e.name
                      )}
                    </td>
                    <td>
                      <div style={{ display: 'flex', gap: 8 }}>
                        <button
                          className="btn-secondary btn-small"
                          onClick={() => setEditingCatalogEntry({ ...e })}
                        >
                          Renommer
                        </button>
                        <button
                          className="btn-small btn-danger"
                          onClick={() => {
                            if (window.confirm(`Supprimer "${e.name}" du catalogue ?`)) {
                              deleteCatalogMutation.mutate(e.id)
                            }
                          }}
                          disabled={deleteCatalogMutation.isPending}
                        >
                          Supprimer
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {tab === 'import-export' && (
        <>
          <div className="section">
            <h2 className="section-title">Exporter les zones</h2>
            <div className="map-upload-form">
              <div className="map-form-row" style={{ gap: 12, alignItems: 'flex-end' }}>
                <div className="map-form-group">
                  <label htmlFor="export-team">Équipe</label>
                  <select
                    id="export-team"
                    value={exportTeamId}
                    onChange={(e) => setExportTeamId(Number(e.target.value))}
                  >
                    <option value={0}>Choisir une équipe</option>
                    {teams.map((t) => (
                      <option key={t.id} value={t.id}>{t.name}</option>
                    ))}
                  </select>
                </div>
                <div className="map-form-group">
                  <label htmlFor="export-map">Carte</label>
                  <select
                    id="export-map"
                    value={exportMapId}
                    onChange={(e) => setExportMapId(Number(e.target.value))}
                  >
                    <option value={0}>Toutes les cartes</option>
                    {maps.map((m) => (
                      <option key={m.id} value={m.id}>{m.name}</option>
                    ))}
                  </select>
                </div>
                <button
                  className="btn-primary"
                  onClick={async () => {
                    setError(null)
                    setSuccess(null)
                    if (!exportTeamId) { setError('Veuillez sélectionner une équipe'); return }
                    try {
                      const blob = exportMapId
                        ? await exportTeamMap(exportTeamId, exportMapId)
                        : await exportAllTeamMaps(exportTeamId)
                      const url = URL.createObjectURL(blob)
                      const a = document.createElement('a')
                      a.href = url
                      const teamName = teams.find(t => t.id === exportTeamId)?.name ?? `team-${exportTeamId}`
                      a.download = exportMapId
                        ? `${teamName}-${maps.find(m => m.id === exportMapId)?.name ?? `map-${exportMapId}`}.json`
                        : `${teamName}-all-zones.json`
                      document.body.appendChild(a)
                      a.click()
                      document.body.removeChild(a)
                      setSuccess('Export téléchargé')
                    } catch (e) {
                      const apiErr = e as { body?: { message?: string } }
                      setError(apiErr?.body?.message ?? "Erreur lors de l'export")
                    }
                  }}
                  disabled={!exportTeamId}
                >
                  Exporter
                </button>
              </div>
            </div>
          </div>

          <div className="section">
            <h2 className="section-title">Importer des zones</h2>
            <div className="map-upload-form">
              <div className="map-form-row" style={{ gap: 12, alignItems: 'flex-end' }}>
                <div className="map-form-group">
                  <label htmlFor="import-team">Équipe</label>
                  <select
                    id="import-team"
                    value={importTeamId}
                    onChange={(e) => setImportTeamId(Number(e.target.value))}
                  >
                    <option value={0}>Choisir une équipe</option>
                    {teams.map((t) => (
                      <option key={t.id} value={t.id}>{t.name}</option>
                    ))}
                  </select>
                </div>
                <div className="map-form-group">
                  <label htmlFor="import-file">Fichier JSON</label>
                  <input
                    id="import-file"
                    type="file"
                    accept=".json"
                    onChange={(e) => setImportFile(e.target.files?.[0] ?? null)}
                  />
                </div>
                <button
                  className="btn-primary"
                  onClick={() => {
                    setError(null)
                    setSuccess(null)
                    if (!importTeamId) { setError('Veuillez sélectionner une équipe'); return }
                    if (!importFile) { setError('Veuillez sélectionner un fichier'); return }

                    const reader = new FileReader()
                    reader.onload = (e) => {
                      try {
                        const data = JSON.parse(e.target?.result as string)
                        if (!Array.isArray(data)) {
                          setError('Le fichier doit contenir un tableau JSON')
                          return
                        }
                        importMutation.mutate({ teamId: importTeamId, data })
                      } catch {
                        setError('Fichier JSON invalide')
                      }
                    }
                    reader.readAsText(importFile)
                  }}
                  disabled={!importTeamId || !importFile || importMutation.isPending}
                >
                  {importMutation.isPending ? 'Import en cours...' : 'Importer'}
                </button>
              </div>
            </div>
          </div>

          {importMutation.data && importMutation.data.errors && importMutation.data.errors.length > 0 && (
            <div className="section">
              <h2 className="section-title">Erreurs d'import</h2>
              <table className="admin-table">
                <thead>
                  <tr>
                    <th>Carte</th>
                    <th>Erreur</th>
                  </tr>
                </thead>
                <tbody>
                  {importMutation.data.errors.map((e, i) => (
                    <tr key={i}>
                      <td>{e.map_name}</td>
                      <td>{e.error}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </>
      )}

      {tab === 'ligues' && (
        <AdminLiguesTab
          onSuccess={(msg) => setSuccess(msg)}
          onError={(msg) => setError(msg)}
        />
      )}

      {tab === 'archives' && (
        <>
          <div className="section">
            <h2 className="section-title">Archivistes</h2>
            <p className="text-muted" style={{ marginBottom: 12 }}>
              Les archivistes ont un accès complet en lecture et écriture aux archives.
            </p>
            {archivistsLoading ? (
              <div className="empty-state"><p>Chargement...</p></div>
            ) : (
              <>
                <table className="admin-table">
                  <thead>
                    <tr>
                      <th>Utilisateur</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {archivists.length === 0 ? (
                      <tr><td colSpan={2}><p className="text-muted">Aucun archiviste</p></td></tr>
                    ) : (
                      archivists.map((a) => (
                        <tr key={a.user_id}>
                          <td>{a.display_name || `#${a.user_id}`}</td>
                          <td>
                            <button
                              className="btn-small btn-danger"
                              onClick={() => {
                                if (window.confirm(`Retirer ${a.display_name || `#${a.user_id}`} des archivistes ?`)) {
                                  removeArchivistMutation.mutate(a.user_id)
                                }
                              }}
                              disabled={removeArchivistMutation.isPending}
                            >
                              Retirer
                            </button>
                          </td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
                <div className="add-membership-form" style={{ marginTop: 12 }}>
                  <h4>Ajouter un archiviste</h4>
                  <div className="add-membership-row">
                    <div className="invite-input-wrapper" style={{ flex: 1, position: 'relative' }}>
                      <input
                        type="text"
                        placeholder="Rechercher un utilisateur..."
                        value={archivistSearch}
                        onChange={(e) => {
                          const value = e.target.value
                          setArchivistSearch(value)
                          setSelectedArchivistUserId(null)
                          if (archivistSearchTimeoutRef.current) clearTimeout(archivistSearchTimeoutRef.current)
                          if (value.trim().length < 2) {
                            setArchivistSearchResults([])
                            return
                          }
                          archivistSearchTimeoutRef.current = setTimeout(async () => {
                            try {
                              const res = await searchUsers(value.trim())
                              setArchivistSearchResults(res.users)
                              setShowArchivistSuggestions(true)
                            } catch {
                              setArchivistSearchResults([])
                            }
                          }, 300)
                        }}
                        onFocus={() => {
                          if (archivistSearchResults.length > 0) setShowArchivistSuggestions(true)
                        }}
                        onBlur={() => setTimeout(() => setShowArchivistSuggestions(false), 200)}
                      />
                      {showArchivistSuggestions && archivistSearchResults.length > 0 && (
                        <ul className="suggestions-dropdown">
                          {archivistSearchResults.map((u) => (
                            <li
                              key={u.id}
                              onMouseDown={() => {
                                setArchivistSearch(u.display_name || u.discord_username || `#${u.id}`)
                                setSelectedArchivistUserId(u.id)
                                setArchivistSearchResults([])
                                setShowArchivistSuggestions(false)
                              }}
                            >
                              {u.display_name || 'Sans nom'} <span className="text-muted">{u.discord_username ? `(${u.discord_username})` : ''}</span>
                            </li>
                          ))}
                        </ul>
                      )}
                    </div>
                    <button
                      className="btn-primary btn-small"
                      onClick={() => {
                        if (!selectedArchivistUserId) { setError('Veuillez sélectionner un utilisateur'); return }
                        addArchivistMutation.mutate(selectedArchivistUserId)
                      }}
                      disabled={!selectedArchivistUserId || addArchivistMutation.isPending}
                    >
                      {addArchivistMutation.isPending ? 'Ajout...' : 'Ajouter'}
                    </button>
                  </div>
                </div>
              </>
            )}
          </div>

          <div className="section">
            <h2 className="section-title">Équipes lectrices</h2>
            <p className="text-muted" style={{ marginBottom: 12 }}>
              Les membres des équipes lectrices ont un accès en lecture aux archives.
            </p>
            {readerTeamsLoading ? (
              <div className="empty-state"><p>Chargement...</p></div>
            ) : (
              <>
                <table className="admin-table">
                  <thead>
                    <tr>
                      <th>Équipe</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {readerTeams.length === 0 ? (
                      <tr><td colSpan={2}><p className="text-muted">Aucune équipe lectrice</p></td></tr>
                    ) : (
                      readerTeams.map((t) => (
                        <tr key={t.team_id}>
                          <td>{t.team_name || `#${t.team_id}`}</td>
                          <td>
                            <button
                              className="btn-small btn-danger"
                              onClick={() => {
                                if (window.confirm(`Retirer le rôle lecteur de l'équipe "${t.team_name || `#${t.team_id}`}" ?`)) {
                                  removeReaderTeamMutation.mutate(t.team_id)
                                }
                              }}
                              disabled={removeReaderTeamMutation.isPending}
                            >
                              Retirer
                            </button>
                          </td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
                <div className="add-membership-form" style={{ marginTop: 12 }}>
                  <h4>Ajouter une équipe lectrice</h4>
                  <div className="add-membership-row">
                    <select
                      value={readerTeamId}
                      onChange={(e) => setReaderTeamId(Number(e.target.value))}
                    >
                      <option value={0}>Choisir une équipe</option>
                      {teams.map((t) => (
                        <option key={t.id} value={t.id}>{t.name}</option>
                      ))}
                    </select>
                    <button
                      className="btn-primary btn-small"
                      onClick={() => {
                        if (!readerTeamId) { setError('Veuillez sélectionner une équipe'); return }
                        addReaderTeamMutation.mutate(readerTeamId)
                      }}
                      disabled={!readerTeamId || addReaderTeamMutation.isPending}
                    >
                      {addReaderTeamMutation.isPending ? 'Ajout...' : 'Ajouter'}
                    </button>
                  </div>
                </div>
              </>
            )}
          </div>
        </>
      )}

      {showPurgeModal && (
        <div className="modal-overlay" onClick={() => setShowPurgeModal(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <h3>Confirmer la purge</h3>
            <p>Êtes-vous sûr de vouloir purger {obsoleteMapsWithoutZones.length} carte(s) obsolète(s) sans zone ? Cette action est irréversible.</p>
            <div className="modal-actions">
              <button className="btn-secondary" onClick={() => setShowPurgeModal(false)}>Annuler</button>
              <button className="btn-danger" onClick={() => purgeMutation.mutate()} disabled={purgeMutation.isPending}>
                {purgeMutation.isPending ? 'Purge en cours...' : 'Purger'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${pad(d.getDate())}/${pad(d.getMonth() + 1)}/${d.getFullYear()} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

export default AdminDashboard
