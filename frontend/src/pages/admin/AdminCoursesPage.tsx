import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { listAdminCourses, fetchLanguages, isAdminLoggedIn } from '../../api'
import type { Course, Language, CEFRLevel, CourseStatus } from '../../types'

const LEVELS: CEFRLevel[] = ['A1', 'A2', 'B1-1', 'B1-2', 'B2', 'C1', 'C2']
const STATUSES: CourseStatus[] = ['draft', 'published', 'archived']

function StatusBadge({ status }: { status: CourseStatus }) {
  return <span className={`badge badge-${status}`}>{status}</span>
}

export default function AdminCoursesPage() {
  const navigate = useNavigate()
  const [courses, setCourses] = useState<Course[]>([])
  const [languages, setLanguages] = useState<Language[]>([])
  const [filters, setFilters] = useState({ language: '', level: '', status: '' })

  useEffect(() => {
    if (!isAdminLoggedIn()) { navigate('/admin/login'); return }
    fetchLanguages().then(setLanguages)
  }, [navigate])

  useEffect(() => {
    listAdminCourses(filters).then(setCourses)
  }, [filters])

  return (
    <div className="admin-page">
      <div className="admin-page-header">
        <h1>Courses</h1>
        <Link to="/admin/courses/generate" className="btn btn-primary">+ Generate Course</Link>
      </div>

      <div className="filter-bar">
        <select className="filter-select" value={filters.language} onChange={e => setFilters(f => ({ ...f, language: e.target.value }))}>
          <option value="">All languages</option>
          {languages.map(l => <option key={l.ID} value={l.Code}>{l.Name}</option>)}
        </select>
        <select className="filter-select" value={filters.level} onChange={e => setFilters(f => ({ ...f, level: e.target.value }))}>
          <option value="">All levels</option>
          {LEVELS.map(l => <option key={l} value={l}>{l}</option>)}
        </select>
        <select className="filter-select" value={filters.status} onChange={e => setFilters(f => ({ ...f, status: e.target.value }))}>
          <option value="">All statuses</option>
          {STATUSES.map(s => <option key={s} value={s}>{s}</option>)}
        </select>
      </div>

      <div className="admin-card" style={{ padding: 0, overflow: 'hidden' }}>
        <table className="data-table">
          <thead>
            <tr>
              <th>Title</th>
              <th>Language</th>
              <th>Level</th>
              <th>Skills</th>
              <th>Status</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {courses.length === 0 ? (
              <tr><td colSpan={6} style={{ textAlign: 'center', color: 'var(--muted)', padding: '2rem' }}>No courses yet.</td></tr>
            ) : courses.map(c => (
              <tr key={c.ID}>
                <td style={{ fontWeight: 500 }}>{c.Title}</td>
                <td>{c.Language?.Code ?? c.language_code}</td>
                <td><span className="badge badge-published" style={{ background: 'var(--bg)', color: 'var(--text)', border: '1px solid var(--border)' }}>{c.Level}</span></td>
                <td style={{ color: 'var(--muted)', fontSize: '0.85rem' }}>{(c.Skills ?? []).join(', ')}</td>
                <td><StatusBadge status={c.Status} /></td>
                <td><Link to={`/admin/courses/${c.ID}`}>View →</Link></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
