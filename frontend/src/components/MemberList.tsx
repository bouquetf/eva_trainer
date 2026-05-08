import {useState} from 'react'
import {useMutation, useQueryClient} from '@tanstack/react-query'
import type {TeamMemberWithUser} from '../api/roles'
import {updateMemberRole} from '../api/roles'
import {useAuthStore} from '../store/auth'
import ConfirmDialog from './ConfirmDialog'

interface MemberListProps {
  teamId: number
  members: TeamMemberWithUser[]
  currentUserId: number
  currentUserRole: string
  onRoleChanged: () => void
}

const roleLabels: Record<string, string> = {
  captain: 'Capitaine',
  cocaptain: 'Cocapitaine',
  editor: 'Éditeur',
  member: 'Membre',
}

function roleBadgeClass(role: string, isAdmin: boolean): string {
  if (isAdmin) return 'role-badge role-badge-admin'
  return `role-badge role-badge-${role}`
}

function MemberList({ teamId, members, currentUserId, currentUserRole, onRoleChanged }: MemberListProps) {
  const queryClient = useQueryClient()
  const user = useAuthStore.getState().user
  const [confirmTarget, setConfirmTarget] = useState<{ userId: number; username: string; role: string } | null>(null)

  const mutation = useMutation({
    mutationFn: ({ userId, role }: { userId: number; role: string }) =>
      updateMemberRole(teamId, userId, { role: role as 'member' | 'editor' | 'cocaptain' }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['team-members', teamId] })
      queryClient.invalidateQueries({ queryKey: ['team', teamId] })
      onRoleChanged()
    },
    onError: () => {
      queryClient.invalidateQueries({ queryKey: ['team-members', teamId] })
    },
  })

  const handleRoleAction = (userId: number, role: string) => {
    mutation.mutate({ userId, role })
    setConfirmTarget(null)
  }

  const canManage = currentUserRole === 'captain' || currentUserRole === 'cocaptain'
  const canPromoteCocaptain = currentUserRole === 'captain'

  return (
    <div>
      {members.map(m => {
        const isSelf = m.user_id === currentUserId
        const isGlobalAdmin = user?.is_admin === true && m.user_id === user.id

        return (
          <div key={m.user_id} className="member-item">
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <span className="member-user-id">
                {m.display_name || `#${m.user_id}`}
                {isSelf && <span style={{ color: '#888', fontSize: 12, marginLeft: 6 }}>(vous)</span>}
              </span>
              <span className={roleBadgeClass(m.role, isGlobalAdmin)}>
                {isGlobalAdmin ? '⚙️ Admin' : roleLabels[m.role] || m.role}
              </span>
              {isGlobalAdmin && (
                <span className="role-badge role-badge-admin" style={{ fontSize: 11 }}>Shield</span>
              )}
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
              <span className="member-joined" style={{ color: '#666', fontSize: 12 }}>
                {new Date(m.joined_at).toLocaleDateString()}
              </span>
              {canManage && !isSelf && !isGlobalAdmin && m.role !== 'captain' && (
                <div className="member-actions">
                  {(currentUserRole === 'captain' || m.role !== 'cocaptain') && (
                    <>
                      {m.role !== 'editor' && (
                        <button
                          className="member-action-btn"
                          onClick={() => setConfirmTarget({ userId: m.user_id, username: m.display_name || `#${m.user_id}`, role: 'editor' })}
                          disabled={mutation.isPending}
                        >
                          ✏️ Éditeur
                        </button>
                      )}
                      {m.role !== 'member' && (
                        <button
                          className="member-action-btn"
                          onClick={() => setConfirmTarget({ userId: m.user_id, username: m.display_name || `#${m.user_id}`, role: 'member' })}
                          disabled={mutation.isPending}
                        >
                          👤 Membre
                        </button>
                      )}
                      {m.role !== 'cocaptain' && canPromoteCocaptain && currentUserRole === 'captain' && (
                        <button
                          className="member-action-btn"
                          onClick={() => setConfirmTarget({ userId: m.user_id, username: m.display_name || `#${m.user_id}`, role: 'cocaptain' })}
                          disabled={mutation.isPending}
                        >
                          ⚓ Cocapitaine
                        </button>
                      )}
                    </>
                  )}
                </div>
              )}
            </div>
          </div>
        )
      })}

      <ConfirmDialog
        open={confirmTarget !== null}
        title="Confirmer le changement de rôle"
        message={
          confirmTarget
            ? `Confirmer le changement de rôle pour ${confirmTarget.username} vers "${roleLabels[confirmTarget.role] || confirmTarget.role}" ?`
            : ''
        }
        confirmLabel="Confirmer"
        onConfirm={() => {
          if (confirmTarget) {
            handleRoleAction(confirmTarget.userId, confirmTarget.role)
          }
        }}
        onCancel={() => setConfirmTarget(null)}
      />
    </div>
  )
}

export default MemberList
