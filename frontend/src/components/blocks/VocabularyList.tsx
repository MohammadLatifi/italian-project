import type { ContentBlock } from '../../types'
import { SpeakButton } from '../SpeakButton'

interface VocabItem {
  word: string
  translation: string
  example?: string
}

interface VocabularyListContent {
  title?: string
  items: VocabItem[]
}

export function VocabularyList({ block }: { block: ContentBlock }) {
  const c = block.Content as unknown as VocabularyListContent
  return (
    <div className="block-vocab">
      {c.title && <h3>{c.title}</h3>}
      {c.items.map((item, i) => (
        <div key={i} className="vocab-item">
          <SpeakButton text={item.word} />
          <div style={{ flex: 1 }}>
            <span className="vocab-word">{item.word}</span>
            <span className="vocab-translation"> — {item.translation}</span>
            {item.example && <div className="vocab-example">{item.example}</div>}
          </div>
        </div>
      ))}
    </div>
  )
}
