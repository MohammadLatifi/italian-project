import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { getLesson, startAttempt, submitAnswer, completeAttempt } from '../api'
import type { Lesson, LessonAttempt, AnswerFeedback, Exercise } from '../types'
import { BlockRenderer } from '../components/blocks/BlockRenderer'
import { ExerciseRenderer } from '../components/exercises/ExerciseRenderer'

export default function LearnLessonPage() {
  const { courseId, lessonId } = useParams<{ courseId: string; lessonId: string }>()
  const cId = Number(courseId)
  const lId = Number(lessonId)

  const [lesson, setLesson] = useState<Lesson | null>(null)
  const [attempt, setAttempt] = useState<LessonAttempt | null>(null)
  const [feedbacks, setFeedbacks] = useState<Record<number, AnswerFeedback>>({})
  const [completing, setCompleting] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    getLesson(cId, lId).then(data => {
      const { current_attempt, ...lessonData } = data as Lesson & { current_attempt?: LessonAttempt }
      setLesson(lessonData)
      if (current_attempt) setAttempt(current_attempt)
    })
  }, [cId, lId])

  async function handleStartAttempt() {
    const a = await startAttempt(cId, lId)
    setAttempt(a)
  }

  async function handleAnswer(exerciseId: number, answer: string): Promise<AnswerFeedback> {
    if (!attempt) throw new Error('No active attempt')
    try {
      const fb = await submitAnswer(cId, lId, attempt.ID, exerciseId, answer)
      setFeedbacks(prev => ({ ...prev, [exerciseId]: fb }))
      return fb
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to submit answer')
      throw err
    }
  }

  async function handleComplete() {
    if (!attempt) return
    setCompleting(true)
    try {
      const updated = await completeAttempt(cId, lId, attempt.ID)
      setAttempt(updated)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to complete lesson')
    } finally {
      setCompleting(false)
    }
  }

  if (!lesson) return <p style={{ padding: '2rem', color: 'var(--muted)' }}>Loading…</p>

  const exercises: Exercise[] = lesson.Exercises ?? []
  const answeredAll = exercises.length > 0 && exercises.every(ex => feedbacks[ex.ID] !== undefined)

  return (
    <div className="lesson-page">
      <Link to={`/courses/${cId}`} className="lesson-back">← Back to course</Link>
      <h1 className="lesson-title">{lesson.Title}</h1>

      {(lesson.ContentBlocks ?? []).map(block => (
        <BlockRenderer key={block.ID} block={block} />
      ))}

      {exercises.length > 0 && (
        <div className="exercises-section">
          <h2>Exercises</h2>

          {!attempt ? (
            <div style={{ textAlign: 'center', padding: '1.5rem 0' }}>
              <p style={{ color: 'var(--muted)', marginBottom: '1rem', fontSize: '0.9rem' }}>
                Ready to test your knowledge? Start the exercises below.
              </p>
              <button className="btn btn-primary" onClick={handleStartAttempt}>
                Start Exercises
              </button>
            </div>
          ) : attempt.Completed ? (
            <div className="score-card">
              <div className="score-number">{attempt.Score ?? 0}%</div>
              <div className="score-label">Lesson complete · {exercises.length} exercises</div>
              <Link to={`/courses/${cId}`} className="btn btn-outline" style={{ marginTop: '1rem', display: 'inline-flex' }}>
                Back to Course
              </Link>
            </div>
          ) : (
            <>
              {exercises.map((ex, i) => (
                <ExerciseRenderer
                  key={ex.ID}
                  exercise={ex}
                  index={i}
                  onAnswer={handleAnswer}
                  feedback={feedbacks[ex.ID]}
                />
              ))}
              {answeredAll && (
                <div style={{ textAlign: 'center', paddingTop: '0.5rem' }}>
                  <button className="btn btn-primary" onClick={handleComplete} disabled={completing}>
                    {completing ? 'Saving…' : 'Complete Lesson'}
                  </button>
                </div>
              )}
            </>
          )}
        </div>
      )}

      {error && <p className="admin-error" style={{ marginTop: '1rem' }}>{error}</p>}
    </div>
  )
}
