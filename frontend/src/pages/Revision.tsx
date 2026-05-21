import { useState, useEffect, useRef } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { listTeamMaps } from '../api/maps'
import { fetchRevisionZones } from '../api/revision'
import { fetchSettings } from '../api/auth'
import { useRevisionStore } from '../store/revision'
import { RevisionMap } from '../components/RevisionMap'
import { RevisionScore } from '../components/RevisionScore'
import type { Map as TeamMap } from '../api/maps'
import type { UserSettings } from '../api/auth'
import '../styles/revision.css'

type Step = 'select' | 'active' | 'finished'

export default function Revision() {
  const { id, mapId } = useParams<{ id: string; mapId?: string }>()
  const navigate = useNavigate()
  const teamId = Number(id)

  const [step, setStep] = useState<Step>(mapId ? 'active' : 'select')
  const [maps, setMaps] = useState<TeamMap[]>([])
  const [selectedMapId, setSelectedMapId] = useState<number>(mapId ? Number(mapId) : 0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const autoStarted = useRef(false)
  const [settings, setSettings] = useState<UserSettings | null>(null)
  const settingsLoaded = useRef(false)

  const {
    currentQuestion,
    questionIndex,
    isAnswerSubmitted,
    isAnswerCorrect,
    correctCount,
    totalAnswered,
    startSession,
    submitAnswer,
    nextQuestion,
    finishSession,
    resetSession,
    setUserAnswer,
    userAnswer,
    config,
    isFinished,
  } = useRevisionStore()

  useEffect(() => {
    if (isNaN(teamId)) return

    const init = async () => {
      setLoading(true)
      setError(null)
      try {
        const mapsData = await listTeamMaps(teamId)
        setMaps(mapsData.maps)
      } catch {
        setError('Erreur lors du chargement des cartes.')
        setLoading(false)
        return
      }
      setLoading(false)
    }
    init()
  }, [teamId])

  useEffect(() => {
    if (settingsLoaded.current) return
    settingsLoaded.current = true
    fetchSettings()
      .then(setSettings)
      .catch(() => {
        setSettings({
          revision_correct_timer: 3,
          revision_incorrect_timer: 5,
          revision_question_count: 20,
        })
      })
  }, [])

  useEffect(() => {
    if (!mapId || autoStarted.current || loading || error || maps.length === 0 || !settings) return
    autoStarted.current = true

    const start = async () => {
      setLoading(true)
      setError(null)
      try {
        const parsedMapId = Number(mapId)
        const zones = await fetchRevisionZones(teamId, parsedMapId)
        if (zones.length === 0) {
          setError('Aucune zone disponible pour la révision sur cette carte.')
          setStep('select')
          setLoading(false)
          return
        }
        const mapName = maps.find((m) => m.id === parsedMapId)?.name ?? `Carte #${parsedMapId}`
        startSession({
          totalQuestions: settings.revision_question_count,
          mapId: parsedMapId,
          mapName,
          teamId,
          correctTimerSeconds: settings.revision_correct_timer,
          incorrectTimerSeconds: settings.revision_incorrect_timer,
        }, zones)
      } catch {
        setError('Erreur lors du chargement des zones de révision.')
      } finally {
        setLoading(false)
      }
    }
    start()
  }, [mapId, loading, error, maps, teamId, startSession, settings])

  const handleMapClick = (mapId: number) => {
    setSelectedMapId((prev) => prev === mapId ? 0 : mapId)
  }

  const handleStartRevision = async () => {
    if (!selectedMapId) return

    setLoading(true)
    setError(null)
    try {
      const zones = await fetchRevisionZones(teamId, selectedMapId)
      if (zones.length === 0) {
        setError('Aucune zone disponible pour la révision sur cette carte.')
        setLoading(false)
        return
      }

      const mapName = maps.find((m) => m.id === selectedMapId)?.name ?? `Carte #${selectedMapId}`

      startSession({
        totalQuestions: settings?.revision_question_count ?? 20,
        mapId: selectedMapId,
        mapName,
        teamId,
        correctTimerSeconds: settings?.revision_correct_timer ?? 3,
        incorrectTimerSeconds: settings?.revision_incorrect_timer ?? 5,
      }, zones)
      setStep('active')
    } catch {
      setError('Erreur lors du chargement des zones de révision.')
    } finally {
      setLoading(false)
    }
  }

  const inputRef = useRef<HTMLInputElement>(null)
  const [showFeedback, setShowFeedback] = useState(false)
  const feedbackTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  const handleSkipFeedback = () => {
    if (!showFeedback) return
    if (feedbackTimer.current) clearTimeout(feedbackTimer.current)
    setShowFeedback(false)
    nextQuestion()
  }

  const handleSubmitAnswer = () => {
    if (!userAnswer.trim() || isAnswerSubmitted) return
    submitAnswer(userAnswer)
    setShowFeedback(true)

    const wasCorrect = useRevisionStore.getState().isAnswerCorrect

    if (feedbackTimer.current) clearTimeout(feedbackTimer.current)
    const delay = (wasCorrect
      ? (config?.correctTimerSeconds ?? 3)
      : (config?.incorrectTimerSeconds ?? 5)) * 1000
    feedbackTimer.current = setTimeout(() => {
      setShowFeedback(false)
      nextQuestion()
    }, delay)
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      if (!isAnswerSubmitted) {
        handleSubmitAnswer()
      } else if (showFeedback) {
        handleSkipFeedback()
      }
    }
    if (e.key === ' ' && showFeedback) {
      e.preventDefault()
      handleSkipFeedback()
    }
  }

  const handleFinishEarly = () => {
    finishSession()
    setStep('finished')
  }

  const handleNewRevision = () => {
    resetSession()
    if (mapId) {
      window.location.reload()
    } else {
      setStep('select')
      setSelectedMapId(0)
    }
  }

  useEffect(() => {
    setShowFeedback(false)
    inputRef.current?.focus()
  }, [currentQuestion])

  useEffect(() => {
    if (!isAnswerSubmitted && inputRef.current && step === 'active') {
      inputRef.current.focus()
    }
  }, [isAnswerSubmitted, step])

  useEffect(() => {
    if (!showFeedback) return
    const handler = () => handleSkipFeedback()
    document.addEventListener('mousedown', handler)
    return () => {
      document.removeEventListener('mousedown', handler)
    }
  }, [showFeedback])

  useEffect(() => {
    return () => {
      if (feedbackTimer.current) clearTimeout(feedbackTimer.current)
    }
  }, [])

  useEffect(() => {
    if (isFinished && step === 'active') {
      setStep('finished')
    }
  }, [isFinished, step])

  if (isNaN(teamId)) {
    return (
      <div style={{ padding: 50, textAlign: 'center', color: '#f44336' }}>
        Paramètres d'URL invalides.
      </div>
    )
  }

  return (
    <div className="revision-container">
      {step === 'select' && (
        <div className="revision-select-screen">
          <div className="revision-header">
            <h1>Mode Révision</h1>
          </div>

          {error && <div className="error-message">{error}</div>}

          {loading ? (
            <div className="spinner" />
          ) : (
            <>
              <div className="revision-map-list">
                {maps.map((m) => (
                  <div
                    key={m.id}
                    className={`revision-map-card${selectedMapId === m.id ? ' selected' : ''}`}
                    onClick={() => handleMapClick(m.id)}
                  >
                    <h3>{m.name}</h3>
                    <div className="meta">
                      {m.is_obsolete ? 'Obsolète' : 'Active'}
                    </div>
                  </div>
                ))}
              </div>

              {selectedMapId > 0 && (
                <div className="revision-config">
                  <button className="btn-primary" onClick={handleStartRevision}>
                    Lancer la révision
                  </button>
                </div>
              )}
            </>
          )}
        </div>
      )}

      {step === 'active' && currentQuestion && config && (
        <div className="revision-active-screen">
          <div className="toolbar">
            <h1 style={{ margin: 0, fontSize: 18, flex: 1, textAlign: 'center' }}>
              Révision - {config.mapName}
            </h1>
            <button className="btn-secondary" onClick={handleFinishEarly}>
              Terminer la révision
            </button>
          </div>

          <div className="main-content">
            <div className="canvas-area">
              <div className="revision-canvas-wrapper">
                <RevisionMap mapId={currentQuestion.map_id} polygons={currentQuestion.polygons} />
              </div>
              <div className="revision-input-area">
                <input
                  ref={inputRef}
                  type="text"
                  placeholder="Nom de la zone..."
                  value={userAnswer}
                  onChange={(e) => setUserAnswer(e.target.value)}
                  onKeyDown={handleKeyDown}
                  disabled={isAnswerSubmitted}
                  className={showFeedback ? (isAnswerCorrect ? 'feedback-correct' : 'feedback-incorrect') : ''}
                />
                {!isAnswerSubmitted && (
                  <button className="btn-primary" onClick={handleSubmitAnswer} disabled={!userAnswer.trim()}>
                    Valider
                  </button>
                )}
              </div>
            </div>

            <div className="side-panel">
              {currentQuestion && <span data-testid="current-zone" style={{ display: 'none' }}>{currentQuestion.name}</span>}
              <RevisionScore
                correctCount={correctCount}
                totalAnswered={totalAnswered}
                questionIndex={questionIndex + 1}
                totalQuestions={config.totalQuestions}
              />
              {showFeedback && (
                <div className={`revision-feedback ${isAnswerCorrect ? 'correct' : 'incorrect'}`}>
                  {isAnswerCorrect ? (
                    '✓ Correct !'
                  ) : (
                    <>
                      ✗ Incorrect
                      <div className="correct-answer">
                        La réponse était : {currentQuestion.name}
                      </div>
                    </>
                  )}
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {step === 'finished' && config && (
        <div className="revision-finished-screen">
          <h2>Révision terminée !</h2>

          <div className="revision-final-score">
            <div className="stat">
              <div className={`stat-value ${correctCount / Math.max(totalAnswered, 1) >= 0.5 ? 'good' : 'bad'}`}>
                {correctCount}
              </div>
              <div className="stat-label">Bonnes réponses</div>
            </div>
            <div className="stat">
              <div className={`stat-value ${totalAnswered - correctCount <= correctCount ? 'good' : 'bad'}`}>
                {totalAnswered - correctCount}
              </div>
              <div className="stat-label">Mauvaises réponses</div>
            </div>
            <div className="stat">
              <div className={`stat-value ${correctCount / Math.max(totalAnswered, 1) >= 0.5 ? 'good' : 'bad'}`}>
                {totalAnswered > 0 ? Math.round((correctCount / totalAnswered) * 100) : 0}%
              </div>
              <div className="stat-label">Précision</div>
            </div>
          </div>

          <div className="revision-actions">
            <button className="btn-primary" onClick={handleNewRevision}>
              Nouvelle révision
            </button>
            <button className="btn-secondary" onClick={() => navigate(`/teams/${teamId}`)}>
              Retour à l'équipe
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
