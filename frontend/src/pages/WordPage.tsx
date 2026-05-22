import { useState, useEffect } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { fetchWord, fetchParagraphs } from '../api'
import { SpeakButton } from '../components/SpeakButton'
import type { Word, Paragraph } from '../types'

export default function WordPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const wordId = Number(id)

  const [word, setWord] = useState<Word | null>(null)
  const [paragraphs, setParagraphs] = useState<Paragraph[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!wordId) return
    setLoading(true)
    Promise.all([fetchWord(wordId), fetchParagraphs()])
      .then(([w, all]) => {
        setWord(w)
        setParagraphs(all.filter(p => p.RelatedWordIDs?.includes(wordId)))
      })
      .catch(() => setError('Could not load data.'))
      .finally(() => setLoading(false))
  }, [wordId])

  if (loading) return <p className="loading">Loading…</p>
  if (error) return <p className="error-msg">{error}</p>
  if (!word) return <p className="error-msg">Word not found.</p>

  return (
    <>
      <button className="back-btn" onClick={() => navigate(-1)}>
        ← Back
      </button>

      <div className="detail-card">
        <div className="detail-word-row">
          <div className="detail-word">{word.Word}</div>
          <SpeakButton text={word.Word} size="md" />
        </div>
        <div className="detail-translation-row">
          <div className="detail-translation">{word.TranslationEN}</div>
          <SpeakButton text={word.TranslationEN} lang="en-US" size="sm" />
        </div>
        {word.ExampleSentence && (
          <div className="detail-example-row">
            <div className="detail-example">{word.ExampleSentence}</div>
            <SpeakButton text={word.ExampleSentence} size="sm" />
          </div>
        )}
      </div>

      <p className="section-heading">
        Paragraphs using this word
        <span>({paragraphs.length})</span>
      </p>

      <div className="card-grid">
        {paragraphs.map(p => (
          <Link key={p.ID} to={`/paragraphs/${p.ID}`} className="para-card">
            <div className="para-italian">{p.Paragraph}</div>
            <div className="para-translation">{p.TranslationEN}</div>
            {p.Context && <span className="context-badge">{p.Context}</span>}
          </Link>
        ))}
        {paragraphs.length === 0 && (
          <p className="empty">No paragraphs linked to this word yet.</p>
        )}
      </div>
    </>
  )
}
