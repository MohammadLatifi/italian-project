import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { listPublishedCourses, fetchLanguages } from '../api'
import type { Course, Language, CEFRLevel } from '../types'

const LEVELS: CEFRLevel[] = ['A1', 'A2', 'B1-1', 'B1-2', 'B2', 'C1', 'C2']

export default function LearnCoursesPage() {
  const [courses, setCourses] = useState<Course[]>([])
  const [languages, setLanguages] = useState<Language[]>([])
  const [filters, setFilters] = useState({ language: '', level: '' })
  const [loading, setLoading] = useState(true)

  useEffect(() => { fetchLanguages().then(setLanguages) }, [])

  useEffect(() => {
    setLoading(true)
    listPublishedCourses(filters).then(data => { setCourses(data); setLoading(false) })
  }, [filters])

  return (
    <div>
      <h1 style={{ fontSize: '1.5rem', fontWeight: 700, marginBottom: '1.25rem' }}>Courses</h1>

      <div className="filter-bar">
        <select className="filter-select" value={filters.language} onChange={e => setFilters(f => ({ ...f, language: e.target.value }))}>
          <option value="">All languages</option>
          {languages.map(l => <option key={l.ID} value={l.Code}>{l.Name}</option>)}
        </select>
        <select className="filter-select" value={filters.level} onChange={e => setFilters(f => ({ ...f, level: e.target.value }))}>
          <option value="">All levels</option>
          {LEVELS.map(l => <option key={l} value={l}>{l}</option>)}
        </select>
      </div>

      {loading ? (
        <p className="loading">Loading courses…</p>
      ) : courses.length === 0 ? (
        <p className="empty">No published courses yet.</p>
      ) : (
        <div className="course-grid">
          {courses.map(c => (
            <Link key={c.ID} to={`/courses/${c.ID}`} className="course-card">
              <div className="course-card-header">
                <span className="course-card-title">{c.Title}</span>
                <span className="course-card-level">{c.Level}</span>
              </div>
              <p className="course-card-desc">{c.Description}</p>
              <div className="course-card-footer">
                <span>{(c.Skills ?? []).join(', ')}</span>
                <span>·</span>
                <span>{(c.Lessons ?? []).length || c.lesson_count || 0} lessons</span>
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}
