export interface Word {
  ID: number
  CreatedAt: string
  UpdatedAt: string
  DeletedAt: string | null
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
