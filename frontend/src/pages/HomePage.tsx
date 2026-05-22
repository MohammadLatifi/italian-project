import { useState, useEffect, useMemo } from 'react'
import { Link } from 'react-router-dom'
import { fetchWords } from '../api'
import { SpeakButton } from '../components/SpeakButton'
import type { Word } from '../types'

export default function HomePage() {
  const [words, setWords] = useState<Word[]>([])
  const [query, setQuery] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    fetchWords()
      .then(setWords)
      .catch(() => setError('Could not load words. Is the Go server running on port 3000?'))
      .finally(() => setLoading(false))
  }, [])

  const filtered = useMemo(() => {
    const q = query.toLowerCase().trim()
    if (!q) return words
    return words.filter(
      w =>
        w.Word.toLowerCase().includes(q) ||
        w.TranslationEN.toLowerCase().includes(q),
    )
  }, [words, query])

  return (
    <>
      <div className="search-wrap">
        <input
          className="search-input"
          type="search"
          placeholder="Search Italian words or English translations…"
          value={query}
          onChange={e => setQuery(e.target.value)}
          autoFocus
        />
      </div>

      {loading && <p className="loading">Loading words…</p>}
      {error && <p className="error-msg">{error}</p>}

      {!loading && !error && (
        <>
          <p className="results-count">
            {filtered.length} word{filtered.length !== 1 ? 's' : ''}
          </p>
          <div className="card-grid">
            {filtered.map(w => (
              <div key={w.ID} className="word-card-wrap">
                <Link to={`/words/${w.ID}`} className="word-card">
                  <div className="word-card-body">
                    <div className="word-italian">{w.Word}</div>
                    <div className="word-translation-row">
                      <div className="word-translation">{w.TranslationEN}</div>
                      <SpeakButton text={w.TranslationEN} lang="en-US" size="sm" />
                    </div>
                    {w.ExampleSentence && (
                      <div className="word-example">{w.ExampleSentence}</div>
                    )}
                  </div>
                </Link>
                <SpeakButton text={w.Word} size="sm" />
              </div>
            ))}
            {filtered.length === 0 && (
              <p className="empty">No words match "{query}"</p>
            )}
          </div>
        </>
      )}
    </>
  )
}
