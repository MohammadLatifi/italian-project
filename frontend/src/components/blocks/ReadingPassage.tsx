import type { ContentBlock } from '../../types'
import { SpeakButton } from '../SpeakButton'

interface ReadingPassageContent {
  title?: string
  text: string
  translation?: string
}

export function ReadingPassage({ block }: { block: ContentBlock }) {
  const c = block.Content as unknown as ReadingPassageContent
  return (
    <div className="block-reading">
      {c.title && (
        <div className="block-reading-header">
          <h3>{c.title}</h3>
          <SpeakButton text={c.text} />
        </div>
      )}
      <p>{c.text}</p>
      {c.translation && (
        <details>
          <summary>Show translation</summary>
          <p>{c.translation}</p>
        </details>
      )}
    </div>
  )
}
