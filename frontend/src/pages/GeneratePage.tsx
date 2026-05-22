import { useState, useEffect, useMemo } from 'react'
import { Link } from 'react-router-dom'
import { fetchWords, generateParagraph } from '../api'
import type { Word, Paragraph } from '../types'

const CONTEXTS = [
  'general', 'sport', 'art', 'software', 'finance',
  'law', 'medical', 'science', 'family', 'nature',
]

export default function GeneratePage() {
  const [words, setWords] = useState<Word[]>([])
  const [wordsLoading, setWordsLoading] = useState(true)

  const [search, setSearch] = useState('')
  const [selected, setSelected] = useState<Set<number>>(new Set())
  const [context, setContext] = useState('general')

  const [generating, setGenerating] = useState(false)
  const [result, setResult] = useState<Paragraph | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    fetchWords()
      .then(setWords)
      .finally(() => setWordsLoading(false))
  }, [])

  const filtered = useMemo(() => {
    const q = search.toLowerCase().trim()
    if (!q) return words
    return words.filter(
      w => w.Word.toLowerCase().includes(q) || w.TranslationEN.toLowerCase().includes(q),
    )
  }, [words, search])

  function toggle(id: number) {
    setSelected(prev => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else if (next.size < 7) next.add(id)
      return next
    })
  }

  async function handleGenerate() {
    setError('')
    setResult(null)
    setGenerating(true)
    try {
      const para = await generateParagraph([...selected], context)
      setResult(para)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Generation failed')
    } finally {
      setGenerating(false)
    }
  }

  const canGenerate = selected.size >= 3 && !generating

  return (
    <>
      <h1 className="gen-title">Generate a paragraph</h1>
      <p className="gen-subtitle">
        Pick 3–7 vocabulary words and a context. Claude will compose a natural Italian
        paragraph using all of them, then save it to the database.
      </p>

      <div className="gen-layout">
        {/* ── Left: word picker ── */}
        <div className="gen-picker">
          <div className="gen-picker-header">
            <span className="gen-picker-label">
              Words <span className="gen-count">{selected.size}/7</span>
            </span>
            {selected.size > 0 && (
              <button className="gen-clear" onClick={() => setSelected(new Set())}>
                Clear
              </button>
            )}
          </div>

          <input
            className="search-input gen-search"
            type="search"
            placeholder="Filter words…"
            value={search}
            onChange={e => setSearch(e.target.value)}
          />

          {selected.size > 0 && (
            <div className="gen-chips">
              {[...selected].map(id => {
                const w = words.find(x => x.ID === id)
                return w ? (
                  <button key={id} className="gen-chip-selected" onClick={() => toggle(id)}>
                    {w.Word} ×
                  </button>
                ) : null
              })}
            </div>
          )}

          <div className="gen-word-list">
            {wordsLoading && <p className="loading" style={{ padding: '24px 0' }}>Loading…</p>}
            {filtered.map(w => (
              <label key={w.ID} className={`gen-word-row ${selected.has(w.ID) ? 'checked' : ''}`}>
                <input
                  type="checkbox"
                  checked={selected.has(w.ID)}
                  onChange={() => toggle(w.ID)}
                  disabled={!selected.has(w.ID) && selected.size >= 7}
                />
                <span className="gen-word-italian">{w.Word}</span>
                <span className="gen-word-en">{w.TranslationEN}</span>
              </label>
            ))}
            {!wordsLoading && filtered.length === 0 && (
              <p className="empty" style={{ padding: '24px 0' }}>No matches</p>
            )}
          </div>
        </div>

        {/* ── Right: settings + result ── */}
        <div className="gen-right">
          <label className="gen-field-label">Context</label>
          <select
            className="gen-select"
            value={context}
            onChange={e => setContext(e.target.value)}
          >
            {CONTEXTS.map(c => (
              <option key={c} value={c}>{c}</option>
            ))}
          </select>

          {selected.size < 3 && (
            <p className="gen-hint">Select at least {3 - selected.size} more word{3 - selected.size !== 1 ? 's' : ''}</p>
          )}

          <button
            className="gen-btn"
            onClick={handleGenerate}
            disabled={!canGenerate}
          >
            {generating ? 'Generating…' : 'Generate paragraph'}
          </button>

          {error && <p className="error-msg" style={{ marginTop: 16 }}>{error}</p>}

          {result && (
            <div className="gen-result">
              <div className="gen-result-tag">Saved to database — ID #{result.ID}</div>
              <div className="detail-para">{result.Paragraph}</div>
              <div className="detail-para-translation">{result.TranslationEN}</div>
              {result.Context && <span className="context-badge">{result.Context}</span>}
              <Link to={`/paragraphs/${result.ID}`} className="gen-view-link">
                View full paragraph →
              </Link>
            </div>
          )}
        </div>
      </div>
    </>
  )
}
