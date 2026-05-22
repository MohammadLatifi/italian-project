import type { Exercise, AnswerFeedback } from '../../types'
import { MultipleChoice } from './MultipleChoice'
import { FillInTheBlank } from './FillInTheBlank'
import { ConversationReconstruction } from './ConversationReconstruction'

interface Props {
  exercise: Exercise
  index: number
  onAnswer: (exerciseId: number, answer: string) => Promise<AnswerFeedback>
  feedback?: AnswerFeedback
}

export function ExerciseRenderer({ exercise, index, onAnswer, feedback }: Props) {
  switch (exercise.Type) {
    case 'multiple-choice':
      return <MultipleChoice exercise={exercise} index={index} onAnswer={onAnswer} feedback={feedback} />
    case 'fill-in-the-blank':
      return <FillInTheBlank exercise={exercise} index={index} onAnswer={onAnswer} feedback={feedback} />
    case 'conversation-reconstruction':
      return <ConversationReconstruction exercise={exercise} index={index} onAnswer={onAnswer} feedback={feedback} />
    default:
      return <div style={{ color: 'var(--muted)', fontSize: '0.875rem' }}>Unknown exercise type: {exercise.Type}</div>
  }
}
