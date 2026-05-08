import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { createTeam } from '../api/teams'

function CreateTeam() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [name, setName] = useState('')
  const [error, setError] = useState<string | null>(null)

  const mutation = useMutation({
    mutationFn: () => createTeam({ name: name.trim() }),
    onSuccess: (team) => {
      queryClient.invalidateQueries({ queryKey: ['my-teams'] })
      navigate(`/teams/${team.id}`)
    },
    onError: (err: unknown) => {
      const apiErr = err as { status?: number; body?: { message?: string } }
      setError(apiErr?.body?.message ?? 'Erreur lors de la création de l\'équipe.')
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    mutation.mutate()
  }

  return (
    <div>
      <h1>Créer une équipe</h1>
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="team-name">Nom de l'équipe</label>
          <input
            id="team-name"
            type="text"
            value={name}
            onChange={e => setName(e.target.value)}
            placeholder="Ex: Équipe A"
            disabled={mutation.isPending}
            autoFocus
          />
          <div className="hint">Lettres, chiffres, espaces, tirets, underscores, apostrophes et points.</div>
          {error && <div className="error">{error}</div>}
        </div>
        <button type="submit" className="btn-primary" disabled={!name.trim() || mutation.isPending}>
          {mutation.isPending ? 'Création...' : 'Créer l\'équipe'}
        </button>
      </form>
    </div>
  )
}

export default CreateTeam
