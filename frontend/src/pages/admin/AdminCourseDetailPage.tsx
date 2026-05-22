import { useEffect, useState } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { getAdminCourse, publishCourse, archiveCourse, patchCourse, generateLessons, isAdminLoggedIn } from '../../api'
import type { Course } from '../../types'

function StatusBadge({ status }: { status: string }) {
  return <span className={`badge badge-${status}`}>{status}</span>
}

export default function AdminCourseDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [course, setCourse] = useState<Course | null>(null)
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [saving, setSaving] = useState(false)
  const [generatingLessons, setGeneratingLessons] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!isAdminLoggedIn()) { navigate('/admin/login'); return }
    getAdminCourse(Number(id)).then(c => { setCourse(c); setTitle(c.Title); setDescription(c.Description) })
  }, [id, navigate])

  async function handleSave() {
    if (!course) return
    setSaving(true)
    try {
      const updated = await patchCourse(course.ID, { Title: title, Description: description })
      setCourse(updated)
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Save failed')
    } finally {
      setSaving(false)
    }
  }

  async function handlePublish() {
    if (!course) return
    setError('')
    try {
      const updated = await publishCourse(course.ID)
      setCourse(prev => prev ? { ...prev, Status: updated.Status } : prev)
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Publish failed')
    }
  }

  async function handleGenerateLessons() {
    if (!course) return
    setGeneratingLessons(true)
    setError('')
    try {
      const updated = await generateLessons(course.ID)
      setCourse(updated)
      setTitle(updated.Title)
      setDescription(updated.Description)
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Lesson generation failed')
    } finally {
      setGeneratingLessons(false)
    }
  }

  async function handleArchive() {
    if (!course) return
    try {
      await archiveCourse(course.ID)
      navigate('/admin/courses')
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Archive failed')
    }
  }

  if (!course) return <p style={{ padding: '2rem', color: 'var(--muted)' }}>Loading…</p>

  const lessons = course.Lessons ?? []

  return (
    <div className="admin-page">
      <Link to="/admin/courses" className="lesson-back">← All courses</Link>

      <div className="admin-page-header">
        <div>
          <h1 style={{ marginBottom: '0.35rem' }}>{course.Title || 'Untitled Course'}</h1>
          <div className="meta-line">
            <StatusBadge status={course.Status} />
            <span className="meta-sep">·</span>
            <span>{course.Level}</span>
            <span className="meta-sep">·</span>
            <span>{(course.Skills ?? []).join(', ')}</span>
          </div>
        </div>
        <div className="action-row">
          {course.Status === 'draft' && <button className="btn btn-primary" onClick={handlePublish}>Publish</button>}
          {course.Status === 'published' && <button className="btn btn-danger" onClick={handleArchive}>Archive</button>}
        </div>
      </div>

      {error && <p className="admin-error" style={{ marginBottom: '1rem' }}>{error}</p>}

      {/* Edit metadata */}
      {course.Status === 'draft' && (
        <div className="admin-card">
          <p className="admin-card-title">Course Details</p>
          <div className="form-group">
            <label className="form-label">Title</label>
            <input className="form-input" value={title} onChange={e => setTitle(e.target.value)} />
          </div>
          <div className="form-group" style={{ marginBottom: 0 }}>
            <label className="form-label">Description</label>
            <textarea className="form-textarea" value={description} onChange={e => setDescription(e.target.value)} />
          </div>
          <div style={{ marginTop: '1rem' }}>
            <button className="btn btn-outline" onClick={handleSave} disabled={saving}>
              {saving ? 'Saving…' : 'Save Changes'}
            </button>
          </div>
        </div>
      )}

      {/* AI Lessons CTA */}
      {course.Status === 'draft' && (
        <div className="admin-card" style={{ borderStyle: 'dashed', background: 'var(--bg)' }}>
          <p className="admin-card-title">AI-Generated Lessons</p>
          <p style={{ fontSize: '0.875rem', color: 'var(--muted)', marginBottom: '1rem' }}>
            {lessons.length === 0
              ? 'No lessons yet. Use AI to generate a full set of lessons, content blocks, and exercises.'
              : `${lessons.length} lesson${lessons.length !== 1 ? 's' : ''} generated. Regenerate to replace with a fresh set.`}
          </p>
          <button className="btn btn-primary" onClick={handleGenerateLessons} disabled={generatingLessons}>
            {generatingLessons ? 'Generating… (up to 30s)' : lessons.length === 0 ? '✨ Generate Lessons with AI' : '🔄 Regenerate Lessons'}
          </button>
        </div>
      )}

      {/* Lesson list */}
      <div style={{ marginTop: '0.5rem' }}>
        <p style={{ fontSize: '0.75rem', fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.08em', color: 'var(--muted)', marginBottom: '0.75rem' }}>
          Lessons ({lessons.length})
        </p>
        {lessons.length === 0 ? (
          <p style={{ color: 'var(--muted)', fontSize: '0.9rem' }}>No lessons yet.</p>
        ) : lessons.map((lesson, i) => (
          <Link key={lesson.ID} to={`/courses/${course.ID}/lessons/${lesson.ID}`} className="lesson-card">
            <div className="lesson-card-title">
              <span style={{ color: 'var(--muted)', fontWeight: 400, marginRight: '0.5rem' }}>{i + 1}.</span>
              {lesson.Title}
            </div>
            <div className="lesson-card-meta">
              <span>{(lesson.ContentBlocks ?? []).length} content blocks</span>
              <span>·</span>
              <span>{(lesson.Exercises ?? []).length} exercises</span>
              <span className="lesson-card-preview">Preview as learner →</span>
            </div>
          </Link>
        ))}
      </div>
    </div>
  )
}
