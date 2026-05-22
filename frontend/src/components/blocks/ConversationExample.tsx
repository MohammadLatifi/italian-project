import type { ContentBlock } from '../../types'
import { SpeakButton } from '../SpeakButton'

interface ConversationLine {
  speaker: string
  text: string
  translation?: string
}

interface ConversationContent {
  title?: string
  lines: ConversationLine[]
}

export function ConversationExample({ block }: { block: ContentBlock }) {
  const c = block.Content as unknown as ConversationContent
  return (
    <div className="block-conversation">
      {c.title && <h3>{c.title}</h3>}
      {c.lines.map((line, i) => (
        <div key={i} className="convo-line">
          <span className="convo-speaker">{line.speaker}</span>
          <div className="convo-text-wrap">
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
              <span className="convo-text">{line.text}</span>
              <SpeakButton text={line.text} />
            </div>
            {line.translation && <div className="convo-translation">{line.translation}</div>}
          </div>
        </div>
      ))}
    </div>
  )
}
