import { useState, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { getGrammarSuggestions, analyzeGrammarSuggestions, acceptSuggestion, dismissSuggestion, fetchLanguages } from '../../api'
import type { GrammarSuggestion, Language, CEFRLevel } from '../../types'

const LEVELS: CEFRLevel[] = ['A1', 'A2', 'B1-1', 'B1-2', 'B2', 'C1', 'C2']

export default function AdminGrammarPage() {
  const navigate = useNavigate()
  const [languages, setLanguages] = useState<Language[]>([])
  const [langCode, setLangCode] = useState('it')
  const [level, setLevel] = useState<CEFRLevel>('A1')
  const [suggestions, setSuggestions] = useState<GrammarSuggestion[]>([])
  const [loadingExisting, setLoadingExisting] = useState(false)
  const [analyzing, setAnalyzing] = useState(false)
  const [accepting, setAccepting] = useState<number | null>(null)
  const [error, setError] = useState('')

  useEffect(() => { fetchLanguages().then(setLanguages) }, [])

  // Auto-load existing pending suggestions whenever language or level changes
  const loadExisting = useCallback(async () => {
    setLoadingExisting(true)
    setError('')
    try {
      const data = await getGrammarSuggestions(langCode, level)
      setSuggestions(data)
    } catch {
      setSuggestions([])
    } finally {
      setLoadingExisting(false)
    }
  }, [langCode, level])

  useEffect(() => { loadExisting() }, [loadExisting])

  async function handleAnalyze() {
    setAnalyzing(true)
    setError('')
    try {
      const data = await analyzeGrammarSuggestions(langCode, level)
      setSuggestions(prev => {
        // merge new suggestions with existing, avoiding duplicates by topic
        const existingTopics = new Set(prev.map(s => s.Topic.toLowerCase()))
        const fresh = data.filter(s => !existingTopics.has(s.Topic.toLowerCase()))
        return [...prev, ...fresh]
      })
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Analysis failed')
    } finally {
      setAnalyzing(false)
    }
  }

  async function handleAccept(id: number) {
    setAccepting(id)
    try {
      const course = await acceptSuggestion(id)
      navigate(`/admin/courses/${course.ID}`)
    } catch {
      setAccepting(null)
    }
  }

  async function handleDismiss(id: number) {
    await dismissSuggestion(id)
    setSuggestions(prev => prev.filter(s => s.ID !== id))
  }

  return (
    <div className="admin-page">
      <div className="admin-page-header">
        <h1>Grammar Coverage</h1>
      </div>
      <p style={{ color: 'var(--muted)', fontSize: '0.9rem', marginBottom: '1.5rem' }}>
        See which grammar topics are missing for a language and level, then generate a course to fill the gap.
      </p>

      <div className="admin-card">
        <p className="admin-card-title">Filter</p>
        <div style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap', alignItems: 'flex-end' }}>
          <div style={{ flex: 1, minWidth: 140 }}>
            <label className="form-label">Language</label>
            <select className="form-select" value={langCode} onChange={e => setLangCode(e.target.value)}>
              {languages.map(l => <option key={l.ID} value={l.Code}>{l.Name}</option>)}
            </select>
          </div>
          <div style={{ flex: 1, minWidth: 120 }}>
            <label className="form-label">Level</label>
            <select className="form-select" value={level} onChange={e => setLevel(e.target.value as CEFRLevel)}>
              {LEVELS.map(l => <option key={l} value={l}>{l}</option>)}
            </select>
          </div>
          <button className="btn btn-primary" onClick={handleAnalyze} disabled={analyzing || loadingExisting}>
            {analyzing ? 'Analysing…' : '✦ Check for New Topics'}
          </button>
        </div>
      </div>

      {error && <p className="admin-error" style={{ marginBottom: '1rem' }}>{error}</p>}

      {loadingExisting ? (
        <p style={{ color: 'var(--muted)', fontSize: '0.9rem' }}>Loading suggestions…</p>
      ) : suggestions.length > 0 ? (
        <div>
          <p style={{ fontSize: '0.75rem', fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.08em', color: 'var(--muted)', marginBottom: '0.75rem' }}>
            {suggestions.length} pending topic{suggestions.length !== 1 ? 's' : ''}
          </p>
          {suggestions.map(s => (
            <div key={s.ID} className="suggestion-card">
              <div className="suggestion-body">
                <p className="suggestion-topic">{s.Topic}</p>
                <p className="suggestion-rationale">{s.Rationale}</p>
              </div>
              <div className="suggestion-actions">
                <button
                  className="btn btn-primary btn-sm"
                  onClick={() => handleAccept(s.ID)}
                  disabled={accepting === s.ID}
                >
                  {accepting === s.ID ? 'Generating…' : 'Generate Course'}
                </button>
                <button className="btn btn-ghost btn-sm" onClick={() => handleDismiss(s.ID)}>
                  Dismiss
                </button>
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div style={{ textAlign: 'center', padding: '2.5rem 0', color: 'var(--muted)' }}>
          <p style={{ fontSize: '0.95rem', marginBottom: '0.5rem' }}>No pending suggestions for {langCode.toUpperCase()} · {level}.</p>
          <p style={{ fontSize: '0.875rem' }}>Click "Check for New Topics" to ask AI for grammar gap suggestions.</p>
        </div>
      )}
    </div>
  )
}
