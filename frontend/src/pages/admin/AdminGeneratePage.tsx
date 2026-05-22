import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { generateCourse, fetchLanguages, isAdminLoggedIn } from '../../api'
import type { Language, CEFRLevel, LanguageSkill } from '../../types'

const LEVELS: CEFRLevel[] = ['A1', 'A2', 'B1-1', 'B1-2', 'B2', 'C1', 'C2']
const ALL_SKILLS: LanguageSkill[] = ['reading', 'writing', 'listening', 'speaking']

export default function AdminGeneratePage() {
  const navigate = useNavigate()
  const [languages, setLanguages] = useState<Language[]>([])
  const [langCode, setLangCode] = useState('it')
  const [level, setLevel] = useState<CEFRLevel>('A1')
  const [skills, setSkills] = useState<LanguageSkill[]>(['reading'])
  const [topic, setTopic] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!isAdminLoggedIn()) { navigate('/admin/login'); return }
    fetchLanguages().then(langs => { setLanguages(langs); if (langs.length) setLangCode(langs[0].Code) })
  }, [navigate])

  function toggleSkill(skill: LanguageSkill) {
    setSkills(prev => prev.includes(skill) ? prev.filter(s => s !== skill) : [...prev, skill])
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!skills.length) { setError('Select at least one skill'); return }
    setError('')
    setLoading(true)
    try {
      const course = await generateCourse({ language_code: langCode, level, skills, topic })
      navigate(`/admin/courses/${course.ID}`)
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Generation failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="admin-page-narrow">
      <div className="admin-page-header" style={{ marginBottom: '0.5rem' }}>
        <h1>Generate Course</h1>
      </div>
      <p style={{ color: 'var(--muted)', fontSize: '0.9rem', marginBottom: '1.75rem' }}>
        AI will design a full course with lessons, content, and exercises.
      </p>

      <div className="admin-card">
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label className="form-label">Language</label>
            <select className="form-select" value={langCode} onChange={e => setLangCode(e.target.value)}>
              {languages.map(l => <option key={l.ID} value={l.Code}>{l.Name}</option>)}
            </select>
          </div>

          <div className="form-group">
            <label className="form-label">CEFR Level</label>
            <select className="form-select" value={level} onChange={e => setLevel(e.target.value as CEFRLevel)}>
              {LEVELS.map(l => <option key={l} value={l}>{l}</option>)}
            </select>
          </div>

          <div className="form-group">
            <label className="form-label">Language Skills</label>
            <div className="skill-grid">
              {ALL_SKILLS.map(s => (
                <label key={s} className={`skill-chip${skills.includes(s) ? ' skill-chip--active' : ''}`}>
                  <input type="checkbox" checked={skills.includes(s)} onChange={() => toggleSkill(s)} />
                  {s.charAt(0).toUpperCase() + s.slice(1)}
                </label>
              ))}
            </div>
          </div>

          <div className="form-group">
            <label className="form-label">Topic</label>
            <input
              className="form-input"
              value={topic}
              onChange={e => setTopic(e.target.value)}
              required
              placeholder="e.g. Ordering food at a restaurant"
            />
          </div>

          {error && <p className="admin-error">{error}</p>}

          {loading && (
            <div className="admin-info" style={{ marginBottom: '0.75rem' }}>
              Generating course with AI… this can take up to 30 seconds.
            </div>
          )}

          <button type="submit" className="btn btn-primary btn-lg" disabled={loading || !topic}>
            {loading ? 'Generating…' : 'Generate Course'}
          </button>
        </form>
      </div>
    </div>
  )
}
