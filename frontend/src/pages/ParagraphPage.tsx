import { useState, useEffect } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { fetchParagraphWithWords } from '../api'
import { SpeakButton } from '../components/SpeakButton'
import type { ParagraphWithWords } from '../types'

export default function ParagraphPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()

  const [data, setData] = useState<ParagraphWithWords | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!id) return
    fetchParagraphWithWords(Number(id))
      .then(setData)
      .catch(() => setError('Could not load paragraph.'))
      .finally(() => setLoading(false))
  }, [id])

  if (loading) return <p className="loading">Loading…</p>
  if (error) return <p className="error-msg">{error}</p>
  if (!data) return <p className="error-msg">Paragraph not found.</p>

  const { paragraph, words } = data

  return (
    <>
      <button className="back-btn" onClick={() => navigate(-1)}>
        ← Back
      </button>

      <div className="detail-card">
        <div className="detail-para-row">
          <div className="detail-para">{paragraph.Paragraph}</div>
          <SpeakButton text={paragraph.Paragraph} size="md" />
        </div>
        <div className="detail-para-translation-row">
          <div className="detail-para-translation">{paragraph.TranslationEN}</div>
          <SpeakButton text={paragraph.TranslationEN} lang="en-US" size="sm" />
        </div>

        {paragraph.Context && (
          <span className="context-badge">{paragraph.Context}</span>
        )}

        {words && words.length > 0 && (
          <>
            <p className="keywords-label">Key words</p>
            <div className="keyword-list">
              {words.map(w => (
                <span key={w.ID} className="keyword-chip-wrap">
                  <Link to={`/words/${w.ID}`} className="keyword-chip">
                    {w.Word}
                  </Link>
                  <SpeakButton text={w.Word} size="sm" />
                </span>
              ))}
            </div>
          </>
        )}
      </div>
    </>
  )
}
