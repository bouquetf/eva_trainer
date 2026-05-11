import { useState, useEffect, useRef } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { listTeamMaps } from '../api/maps'
import { fetchRevisionZones } from '../api/revision'
import { useRevisionStore } from '../store/revision'
import { RevisionMap } from '../components/RevisionMap'
import { RevisionScore } from '../components/RevisionScore'
import type { Map as TeamMap } from '../api/maps'
import '../styles/revision.css'

type Step = 'select' | 'active' | 'finished'

export default function Revision() {
  const { id, mapId } = useParams<{ id: string; mapId?: string }>()
  const navigate = useNavigate()
  const teamId = Number(id)

  const [step, setStep] = useState<Step>(mapId ? 'active' : 'select')
  const [maps, setMaps] = useState<TeamMap[]>([])
  const [selectedMapId, setSelectedMapId] = useState<number>(mapId ? Number(mapId) : 0)
  const [totalQuestions, setTotalQuestions] = useState(20)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const autoStarted = useRef(false)

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
    if (!mapId || autoStarted.current || loading || error || maps.length === 0) return
    autoStarted.current = true

    const start = async () => {
      setLoading(true)
      setError(null)
      try {
        const parsedMapId = Number(mapId)
        const zones = await fetchRevisionZones(teamId, parsedMapId)
        if (zones.length === 0) {
          setError('Aucune zone disponible pour la révision sur cette carte.')
          setLoading(false)
          return
        }
        const mapName = maps.find((m) => m.id === parsedMapId)?.name ?? `Carte #${parsedMapId}`
        startSession({ totalQuestions: 20, mapId: parsedMapId, mapName, teamId }, zones)
      } catch {
        setError('Erreur lors du chargement des zones de révision.')
      } finally {
        setLoading(false)
      }
    }
    start()
  }, [mapId, loading, error, maps, teamId, startSession])

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

      startSession(
        { totalQuestions, mapId: selectedMapId, mapName, teamId },
        zones,
      )
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

  const handleSubmitAnswer = () => {
    if (!userAnswer.trim() || isAnswerSubmitted) return
    submitAnswer(userAnswer)
    setShowFeedback(true)

    if (feedbackTimer.current) clearTimeout(feedbackTimer.current)
    feedbackTimer.current = setTimeout(() => {
      setShowFeedback(false)
      nextQuestion()
    }, 1000)
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      handleSubmitAnswer()
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
      setTotalQuestions(20)
    }
  }

  useEffect(() => {
    setShowFeedback(false)
    inputRef.current?.focus()
  }, [currentQuestion])

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
                  <label htmlFor="totalQuestions">Nombre de questions</label>
                  <input
                    id="totalQuestions"
                    type="number"
                    min={1}
                    value={totalQuestions}
                    onChange={(e) => setTotalQuestions(Math.max(1, parseInt(e.target.value, 10) || 1))}
                  />
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
                <RevisionMap mapId={config.mapId} polygons={currentQuestion.polygons} />
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
