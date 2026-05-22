export interface Word {
  ID: number
  CreatedAt: string
  UpdatedAt: string
  DeletedAt: string | null
  LanguageID: number | null
  Word: string
  TranslationEN: string
  ExampleSentence: string
  PracticeCount: number
  HowManyFalse: number
  LastPracticeFalse: boolean
}

export interface Paragraph {
  ID: number
  CreatedAt: string
  UpdatedAt: string
  DeletedAt: string | null
  LanguageID: number | null
  Paragraph: string
  TranslationEN: string
  RelatedWordIDs: number[] | null
  PracticeCount: number
  Context: string
}

export interface ParagraphWithWords {
  paragraph: Paragraph
  words: Word[]
}

// ── New types ─────────────────────────────────────────────────────────────────

export interface Language {
  ID: number
  Code: string
  Name: string
}

export type CEFRLevel = 'A1' | 'A2' | 'B1-1' | 'B1-2' | 'B2' | 'C1' | 'C2'
export type CourseStatus = 'draft' | 'published' | 'archived'
export type LanguageSkill = 'reading' | 'writing' | 'listening' | 'speaking'
export type ContentBlockType = 'vocabulary-list' | 'grammar-note' | 'conversation-example' | 'reading-passage'
export type ExerciseType = 'multiple-choice' | 'fill-in-the-blank' | 'conversation-reconstruction'

export interface ContentBlock {
  ID: number
  LessonID: number
  Type: ContentBlockType
  Content: Record<string, unknown>
  Position: number
}

export interface Exercise {
  ID: number
  LessonID: number
  Type: ExerciseType
  Question: string
  Options: string[] | null
  Skills: LanguageSkill[]
  Position: number
  // correct_answer omitted — returned only in answer submission response
}

export interface Lesson {
  ID: number
  CourseID: number
  Title: string
  Position: number
  ContentBlocks?: ContentBlock[]
  Exercises?: Exercise[]
}

export interface Course {
  ID: number
  CreatedAt: string
  UpdatedAt: string
  LanguageID: number
  Language?: Language
  language_code?: string
  Level: CEFRLevel
  Title: string
  Topic: string
  Description: string
  Status: CourseStatus
  Skills: LanguageSkill[]
  Lessons?: Lesson[]
  lesson_count?: number
}

export interface LessonAttempt {
  ID: number
  LessonID: number
  LearnerToken: string
  CurrentPosition: number
  Answers: Record<string, string>
  Score: number | null
  Completed: boolean
  CompletedAt: string | null
}

export interface GrammarSuggestion {
  ID: number
  LanguageID: number
  Level: CEFRLevel
  Topic: string
  Rationale: string
  Status: 'pending' | 'accepted' | 'dismissed'
}

export interface AnswerFeedback {
  correct: boolean
  correct_answer: string
  explanation: string | null
  next_position: number
}
