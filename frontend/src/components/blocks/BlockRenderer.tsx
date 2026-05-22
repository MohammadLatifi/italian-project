import type { ContentBlock } from '../../types'
import { GrammarNote } from './GrammarNote'
import { VocabularyList } from './VocabularyList'
import { ConversationExample } from './ConversationExample'
import { ReadingPassage } from './ReadingPassage'

export function BlockRenderer({ block }: { block: ContentBlock }) {
  switch (block.Type) {
    case 'grammar-note': return <GrammarNote block={block} />
    case 'vocabulary-list': return <VocabularyList block={block} />
    case 'conversation-example': return <ConversationExample block={block} />
    case 'reading-passage': return <ReadingPassage block={block} />
    default: return <div style={{ color: '#999', fontSize: '0.875rem' }}>Unknown block type: {block.Type}</div>
  }
}
