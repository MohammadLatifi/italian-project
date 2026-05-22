import { Routes, Route, Link, NavLink } from 'react-router-dom'
import HomePage from './pages/HomePage'
import WordPage from './pages/WordPage'
import ParagraphPage from './pages/ParagraphPage'
import GeneratePage from './pages/GeneratePage'

export default function App() {
  return (
    <div className="app">
      <header className="site-header">
        <Link to="/" className="site-logo">Italiano</Link>
        <nav className="site-nav">
          <NavLink to="/" end className={({ isActive }) => isActive ? 'nav-link active' : 'nav-link'}>
            Words
          </NavLink>
          <NavLink to="/generate" className={({ isActive }) => isActive ? 'nav-link active' : 'nav-link'}>
            Generate
          </NavLink>
        </nav>
      </header>
      <main className="site-main">
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/words/:id" element={<WordPage />} />
          <Route path="/paragraphs/:id" element={<ParagraphPage />} />
          <Route path="/generate" element={<GeneratePage />} />
        </Routes>
      </main>
    </div>
  )
}
