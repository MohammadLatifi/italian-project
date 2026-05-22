import type { Exercise, AnswerFeedback } from '../../types'

interface Props {
  exercise: Exercise
  index: number
  onAnswer: (exerciseId: number, answer: string) => Promise<AnswerFeedback>
  feedback?: AnswerFeedback
}

export function MultipleChoice({ exercise, index, onAnswer, feedback }: Props) {
  const options = exercise.Options ?? []

  // correct_answer is stored as 0-based index string ("0","1","2","3")
  const correctIndex = feedback ? parseInt(feedback.correct_answer) : -1
  const correctText = correctIndex >= 0 ? (options[correctIndex] ?? feedback!.correct_answer) : ''

  async function handleClick(optionIndex: number) {
    if (feedback) return
    await onAnswer(exercise.ID, String(optionIndex))
  }

  return (
    <div className="exercise-card">
      <p className="exercise-num">Exercise {index + 1} · Multiple Choice</p>
      <p className="exercise-question">{exercise.Question}</p>
      <div className="mc-options">
        {options.map((opt, i) => {
          let cls = 'mc-option'
          if (feedback) {
            if (i === correctIndex) cls += ' mc-option--correct'
          }
          return (
            <button key={i} className={cls} onClick={() => handleClick(i)} disabled={!!feedback}>
              <span style={{ color: 'var(--muted)', marginRight: '0.5rem', fontSize: '0.8rem' }}>
                {String.fromCharCode(65 + i)}.
              </span>
              {opt}
            </button>
          )
        })}
      </div>
      {feedback && (
        <div className={`exercise-feedback exercise-feedback--${feedback.correct ? 'correct' : 'wrong'}`}>
          <strong>{feedback.correct ? '✓ Correct!' : `✗ Correct answer: ${correctText}`}</strong>
          {feedback.explanation && <span>{feedback.explanation}</span>}
        </div>
      )}
    </div>
  )
}
