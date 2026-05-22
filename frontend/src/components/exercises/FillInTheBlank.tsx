import { useState } from 'react'
import type { Exercise, AnswerFeedback } from '../../types'

interface Props {
  exercise: Exercise
  index: number
  onAnswer: (exerciseId: number, answer: string) => Promise<AnswerFeedback>
  feedback?: AnswerFeedback
}

export function FillInTheBlank({ exercise, index, onAnswer, feedback }: Props) {
  const [value, setValue] = useState('')

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!value.trim() || feedback) return
    await onAnswer(exercise.ID, value.trim())
  }

  return (
    <div className="exercise-card">
      <p className="exercise-num">Exercise {index + 1} · Fill in the Blank</p>
      <p className="exercise-question">{exercise.Question}</p>
      <form onSubmit={handleSubmit}>
        <div className="fitb-row">
          <input
            className={`fitb-input${feedback ? (feedback.correct ? ' fitb-input--correct' : ' fitb-input--wrong') : ''}`}
            value={value}
            onChange={e => setValue(e.target.value)}
            disabled={!!feedback}
            placeholder="Type your answer…"
          />
          <button type="submit" className="btn btn-primary" disabled={!!feedback || !value.trim()}>
            Check
          </button>
        </div>
      </form>
      {feedback && (
        <div className={`exercise-feedback exercise-feedback--${feedback.correct ? 'correct' : 'wrong'}`}>
          <strong>{feedback.correct ? '✓ Correct!' : `✗ Correct answer: ${feedback.correct_answer}`}</strong>
          {feedback.explanation && <span>{feedback.explanation}</span>}
        </div>
      )}
    </div>
  )
}
