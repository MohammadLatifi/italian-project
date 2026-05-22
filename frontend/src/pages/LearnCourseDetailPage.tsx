import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { getCourse } from '../api'
import type { Course } from '../types'

export default function LearnCourseDetailPage() {
  const { courseId } = useParams<{ courseId: string }>()
  const [course, setCourse] = useState<Course | null>(null)

  useEffect(() => {
    getCourse(Number(courseId)).then(setCourse)
  }, [courseId])

  if (!course) return <p style={{ padding: '1rem' }}>Loading…</p>

  return (
    <div style={{ padding: '1rem', maxWidth: 700, margin: '0 auto' }}>
      <Link to="/courses" style={{ fontSize: '0.875rem', color: '#555' }}>← All courses</Link>
      <h1 style={{ marginTop: '0.5rem' }}>{course.Title}</h1>
      <p style={{ color: '#555' }}>{course.Description}</p>
      <p style={{ fontSize: '0.875rem', color: '#888' }}>
        Level: {course.Level} · Skills: {(course.Skills ?? []).join(', ')}
      </p>
      <h2>Lessons</h2>
      {(course.Lessons ?? []).length === 0 ? <p>No lessons yet.</p> : (
        <ol style={{ paddingLeft: '1.25rem' }}>
          {(course.Lessons ?? []).map(lesson => (
            <li key={lesson.ID} style={{ marginBottom: '0.5rem' }}>
              <Link to={`/courses/${course.ID}/lessons/${lesson.ID}`}>{lesson.Title}</Link>
            </li>
          ))}
        </ol>
      )}
    </div>
  )
}
