import { useState } from 'react'
import type { Exercise, AnswerFeedback } from '../../types'

interface Props {
  exercise: Exercise
  index: number
  onAnswer: (exerciseId: number, answer: string) => Promise<AnswerFeedback>
  feedback?: AnswerFeedback
}

export function ConversationReconstruction({ exercise, index, onAnswer, feedback }: Props) {
  const [selected, setSelected] = useState<string[]>([])
  const options = exercise.Options ?? []

  function toggle(opt: string) {
    if (feedback) return
    setSelected(prev => prev.includes(opt) ? prev.filter(o => o !== opt) : [...prev, opt])
  }

  async function handleSubmit() {
    if (feedback || selected.length === 0) return
    await onAnswer(exercise.ID, selected.join(' | '))
  }

  return (
    <div className="exercise-card">
      <p className="exercise-num">Exercise {index + 1} · Conversation Order</p>
      <p className="exercise-question">{exercise.Question}</p>
      <div className="cr-chips">
        {options.map((opt, i) => (
          <button
            key={i}
            className={`cr-chip${selected.includes(opt) ? ' cr-chip--selected' : ''}`}
            onClick={() => toggle(opt)}
            disabled={!!feedback}
          >
            {opt}
          </button>
        ))}
      </div>
      {selected.length > 0 && !feedback && (
        <div className="cr-order">
          <strong>Your order:</strong> {selected.join(' → ')}
        </div>
      )}
      <button className="btn btn-primary btn-sm" onClick={handleSubmit} disabled={!!feedback || selected.length === 0}>
        Check Order
      </button>
      {feedback && (
        <div className={`exercise-feedback exercise-feedback--${feedback.correct ? 'correct' : 'wrong'}`} style={{ marginTop: '0.875rem' }}>
          <strong>{feedback.correct ? '✓ Correct!' : `✗ Correct order: ${feedback.correct_answer}`}</strong>
          {feedback.explanation && <span>{feedback.explanation}</span>}
        </div>
      )}
    </div>
  )
}
