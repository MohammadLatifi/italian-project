import type { ContentBlock } from '../../types'

interface GrammarNoteContent {
  title: string
  body: string
  examples?: string[]
}

export function GrammarNote({ block }: { block: ContentBlock }) {
  const c = block.Content as unknown as GrammarNoteContent
  return (
    <div className="block-grammar">
      <h3>{c.title}</h3>
      <p>{c.body}</p>
      {c.examples && c.examples.length > 0 && (
        <ul>
          {c.examples.map((ex, i) => <li key={i}>{ex}</li>)}
        </ul>
      )}
    </div>
  )
}
