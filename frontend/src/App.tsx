import { Routes, Route, Link, NavLink, Navigate } from 'react-router-dom'
import HomePage from './pages/HomePage'
import WordPage from './pages/WordPage'
import ParagraphPage from './pages/ParagraphPage'
import GeneratePage from './pages/GeneratePage'
import LearnCoursesPage from './pages/LearnCoursesPage'
import LearnCourseDetailPage from './pages/LearnCourseDetailPage'
import LearnLessonPage from './pages/LearnLessonPage'
import AdminLoginPage from './pages/admin/AdminLoginPage'
import AdminLayout from './pages/admin/AdminLayout'
import AdminCoursesPage from './pages/admin/AdminCoursesPage'
import AdminGeneratePage from './pages/admin/AdminGeneratePage'
import AdminCourseDetailPage from './pages/admin/AdminCourseDetailPage'
import AdminGrammarPage from './pages/admin/AdminGrammarPage'
import AdminLanguagesPage from './pages/admin/AdminLanguagesPage'
import { isAdminLoggedIn } from './api'

export default function App() {
  return (
    <div className="app">
      <header className="site-header">
        <Link to="/" className="site-logo">Italiano</Link>
        <nav className="site-nav">
          <NavLink to="/" end className={({ isActive }) => isActive ? 'nav-link active' : 'nav-link'}>Words</NavLink>
          <NavLink to="/generate" className={({ isActive }) => isActive ? 'nav-link active' : 'nav-link'}>Generate</NavLink>
          <NavLink to="/courses" className={({ isActive }) => isActive ? 'nav-link active' : 'nav-link'}>Courses</NavLink>
          {isAdminLoggedIn() && (
            <NavLink to="/admin/courses" className={({ isActive }) => isActive ? 'nav-link active' : 'nav-link'}>Admin</NavLink>
          )}
        </nav>
      </header>

      <Routes>
        {/* Public pages — centered layout */}
        <Route path="/" element={<main className="site-main"><HomePage /></main>} />
        <Route path="/words/:id" element={<main className="site-main"><WordPage /></main>} />
        <Route path="/paragraphs/:id" element={<main className="site-main"><ParagraphPage /></main>} />
        <Route path="/generate" element={<main className="site-main"><GeneratePage /></main>} />
        <Route path="/courses" element={<main className="site-main"><LearnCoursesPage /></main>} />
        <Route path="/courses/:courseId" element={<main className="site-main"><LearnCourseDetailPage /></main>} />
        <Route path="/courses/:courseId/lessons/:lessonId" element={<LearnLessonPage />} />

        {/* Admin login — no sidebar */}
        <Route path="/admin/login" element={<main className="site-main"><AdminLoginPage /></main>} />

        {/* Admin — sidebar layout via AdminLayout + Outlet */}
        <Route path="/admin" element={<AdminLayout />}>
          <Route index element={<Navigate to="/admin/courses" replace />} />
          <Route path="courses" element={<AdminCoursesPage />} />
          <Route path="courses/generate" element={<AdminGeneratePage />} />
          <Route path="courses/:id" element={<AdminCourseDetailPage />} />
          <Route path="grammar" element={<AdminGrammarPage />} />
          <Route path="languages" element={<AdminLanguagesPage />} />
        </Route>
      </Routes>
    </div>
  )
}
