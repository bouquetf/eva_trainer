import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { transferCaptaincy } from '../api/roles'
import type { TeamMemberWithUser } from '../api/roles'
import { useAuthStore } from '../store/auth'

interface TransferCaptaincyDialogProps {
  isOpen: boolean
  teamId: number
  members: TeamMemberWithUser[]
  onClose: () => void
  onTransferred: () => void
}

function TransferCaptaincyDialog({ isOpen, teamId, members, onClose, onTransferred }: TransferCaptaincyDialogProps) {
  const queryClient = useQueryClient()
  const user = useAuthStore.getState().user
  const [selectedUserId, setSelectedUserId] = useState<number | null>(null)

  const mutation = useMutation({
    mutationFn: () => transferCaptaincy(teamId, { user_id: selectedUserId! }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['team-members', teamId] })
      queryClient.invalidateQueries({ queryKey: ['team', teamId] })
      queryClient.invalidateQueries({ queryKey: ['my-teams'] })
      onTransferred()
    },
    onError: () => {
      queryClient.invalidateQueries({ queryKey: ['team-members', teamId] })
    },
  })

  if (!isOpen) return null

  const candidateMembers = members.filter(m => {
    if (m.user_id === user?.id) return false
    return !(user?.is_admin && m.user_id === user.id);

  })

  const selectedMember = members.find(m => m.user_id === selectedUserId)

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={e => e.stopPropagation()}>
        <h3>Transférer la capitainerie</h3>
        <p>Sélectionnez le nouveau capitaine parmi les membres de l'équipe.</p>

        <select
          className="transfer-select"
          value={selectedUserId ?? ''}
          onChange={e => setSelectedUserId(e.target.value ? Number(e.target.value) : null)}
        >
          <option value="">-- Sélectionner un membre --</option>
          {candidateMembers.map(m => (
            <option key={m.user_id} value={m.user_id}>
              {m.display_name || `#${m.user_id}`} ({m.role})
            </option>
          ))}
        </select>

        {selectedMember && (
          <div className="transfer-warning">
            Vous allez transférer votre rôle de capitaine à {selectedMember.display_name || `#${selectedMember.user_id}`}.
            Vous deviendrez cocapitaine. Cette action est irréversible.
          </div>
        )}

        {mutation.isError && (
          <div className="error-message">
            {mutation.error && typeof mutation.error === 'object' && 'body' in mutation.error
              ? (mutation.error as { body?: { message?: string } }).body?.message ?? 'Erreur lors du transfert.'
              : 'Erreur lors du transfert.'}
          </div>
        )}

        <div className="modal-actions">
          <button className="btn-secondary" onClick={onClose} disabled={mutation.isPending}>
            Annuler
          </button>
          <button
            className="btn-danger"
            onClick={() => mutation.mutate()}
            disabled={selectedUserId === null || mutation.isPending}
          >
            {mutation.isPending ? 'Transfert...' : 'Confirmer le transfert'}
          </button>
        </div>
      </div>
    </div>
  )
}

export default TransferCaptaincyDialog
