import { useEffect, useState } from 'react'
import { fetchLanguages } from '../../api'
import type { Language } from '../../types'

const BASE = '/api'

function adminHeaders(): Record<string, string> {
  const token = localStorage.getItem('admin_token')
  return token ? { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' } : { 'Content-Type': 'application/json' }
}

async function createLanguage(code: string, name: string): Promise<Language> {
  const res = await fetch(`${BASE}/admin/languages`, {
    method: 'POST',
    headers: adminHeaders(),
    body: JSON.stringify({ code, name }),
  })
  const data = await res.json()
  if (!res.ok) throw new Error(data.error ?? 'Failed to create')
  return data
}

async function updateLanguage(id: number, name: string): Promise<Language> {
  const res = await fetch(`${BASE}/admin/languages/${id}`, {
    method: 'PATCH',
    headers: adminHeaders(),
    body: JSON.stringify({ name }),
  })
  const data = await res.json()
  if (!res.ok) throw new Error(data.error ?? 'Failed to update')
  return data
}

async function deleteLanguage(id: number): Promise<void> {
  const res = await fetch(`${BASE}/admin/languages/${id}`, {
    method: 'DELETE',
    headers: adminHeaders(),
  })
  if (!res.ok) {
    const data = await res.json().catch(() => ({}))
    throw new Error(data.error ?? 'Failed to delete')
  }
}

export default function AdminLanguagesPage() {
  const [languages, setLanguages] = useState<Language[]>([])
  const [newCode, setNewCode] = useState('')
  const [newName, setNewName] = useState('')
  const [adding, setAdding] = useState(false)
  const [editId, setEditId] = useState<number | null>(null)
  const [editName, setEditName] = useState('')
  const [error, setError] = useState('')

  useEffect(() => { fetchLanguages().then(setLanguages) }, [])

  async function handleAdd(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setAdding(true)
    try {
      const lang = await createLanguage(newCode.toLowerCase().trim(), newName.trim())
      setLanguages(prev => [...prev, lang])
      setNewCode('')
      setNewName('')
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to add language')
    } finally {
      setAdding(false)
    }
  }

  async function handleUpdate(id: number) {
    setError('')
    try {
      const updated = await updateLanguage(id, editName.trim())
      setLanguages(prev => prev.map(l => l.ID === id ? updated : l))
      setEditId(null)
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to update')
    }
  }

  async function handleDelete(id: number, name: string) {
    if (!confirm(`Delete "${name}"? This cannot be undone.`)) return
    setError('')
    try {
      await deleteLanguage(id)
      setLanguages(prev => prev.filter(l => l.ID !== id))
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to delete')
    }
  }

  return (
    <div className="admin-page">
      <div className="admin-page-header">
        <h1>Languages</h1>
      </div>
      <p style={{ color: 'var(--muted)', fontSize: '0.9rem', marginBottom: '1.75rem' }}>
        Manage the languages available for course generation.
      </p>

      {error && <p className="admin-error" style={{ marginBottom: '1rem' }}>{error}</p>}

      {/* Language list */}
      <div className="admin-card" style={{ padding: 0, overflow: 'hidden', marginBottom: '1.5rem' }}>
        <table className="data-table">
          <thead>
            <tr>
              <th>Code</th>
              <th>Name</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {languages.length === 0 ? (
              <tr><td colSpan={3} style={{ textAlign: 'center', color: 'var(--muted)', padding: '2rem' }}>No languages yet.</td></tr>
            ) : languages.map(l => (
              <tr key={l.ID}>
                <td>
                  <span style={{ fontFamily: 'monospace', background: 'var(--bg)', padding: '2px 8px', borderRadius: 4, fontSize: '0.85rem', fontWeight: 600 }}>
                    {l.Code}
                  </span>
                </td>
                <td>
                  {editId === l.ID ? (
                    <div style={{ display: 'flex', gap: '0.5rem' }}>
                      <input
                        className="form-input"
                        style={{ padding: '6px 10px', fontSize: '0.875rem' }}
                        value={editName}
                        onChange={e => setEditName(e.target.value)}
                        autoFocus
                        onKeyDown={e => { if (e.key === 'Enter') handleUpdate(l.ID); if (e.key === 'Escape') setEditId(null) }}
                      />
                      <button className="btn btn-primary btn-sm" onClick={() => handleUpdate(l.ID)}>Save</button>
                      <button className="btn btn-ghost btn-sm" onClick={() => setEditId(null)}>Cancel</button>
                    </div>
                  ) : (
                    <span style={{ fontWeight: 500 }}>{l.Name}</span>
                  )}
                </td>
                <td style={{ textAlign: 'right' }}>
                  {editId !== l.ID && (
                    <div style={{ display: 'flex', gap: '0.5rem', justifyContent: 'flex-end' }}>
                      <button className="btn btn-outline btn-sm" onClick={() => { setEditId(l.ID); setEditName(l.Name) }}>
                        Edit
                      </button>
                      <button className="btn btn-danger btn-sm" onClick={() => handleDelete(l.ID, l.Name)}>
                        Delete
                      </button>
                    </div>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Add language form */}
      <div className="admin-card">
        <p className="admin-card-title">Add Language</p>
        <form onSubmit={handleAdd} style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap', alignItems: 'flex-end' }}>
          <div style={{ flex: '0 0 120px' }}>
            <label className="form-label">Code</label>
            <input
              className="form-input"
              value={newCode}
              onChange={e => setNewCode(e.target.value)}
              required
              placeholder="e.g. fr"
              maxLength={10}
              pattern="[a-zA-Z\-]+"
              title="Letters and hyphens only"
            />
          </div>
          <div style={{ flex: 1, minWidth: 160 }}>
            <label className="form-label">Name</label>
            <input
              className="form-input"
              value={newName}
              onChange={e => setNewName(e.target.value)}
              required
              placeholder="e.g. French"
            />
          </div>
          <div>
            <button type="submit" className="btn btn-primary" disabled={adding}>
              {adding ? 'Adding…' : '+ Add Language'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
