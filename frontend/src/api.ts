import type { Word, Paragraph, ParagraphWithWords } from './types'

const BASE = '/api'

async function get<T>(path: string): Promise<T> {
  const res = await fetch(BASE + path)
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`)
  return res.json()
}

export const fetchWords = (): Promise<Word[]> => get('/words')
export const fetchWord = (id: number): Promise<Word> => get(`/words/${id}`)
export const fetchParagraphs = (): Promise<Paragraph[]> => get('/paragraphs')
export const fetchParagraphWithWords = (id: number): Promise<ParagraphWithWords> =>
  get(`/paragraphs/${id}/words`)

export async function generateParagraph(wordIds: number[], context: string): Promise<Paragraph> {
  const res = await fetch('/api/generate/paragraph', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ word_ids: wordIds, context }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error ?? res.statusText)
  }
  return res.json()
}
