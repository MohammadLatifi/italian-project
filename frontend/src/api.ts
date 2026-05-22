import type {
  Word, Paragraph, ParagraphWithWords,
  Language, Course, Lesson, GrammarSuggestion, LessonAttempt, AnswerFeedback,
} from './types'

const BASE = '/api'

// ── helpers ───────────────────────────────────────────────────────────────────

function getAdminToken(): string | null {
  return localStorage.getItem('admin_token')
}

function getLearnerToken(): string {
  let token = localStorage.getItem('learner_token')
  if (!token) {
    token = crypto.randomUUID()
    localStorage.setItem('learner_token', token)
  }
  return token
}

async function get<T>(path: string, headers?: Record<string, string>): Promise<T> {
  const res = await fetch(BASE + path, { headers })
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`)
  return res.json()
}

async function post<T>(path: string, body?: unknown, headers?: Record<string, string>): Promise<T> {
  const res = await fetch(BASE + path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...headers },
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error ?? res.statusText)
  }
  return res.json()
}

async function patch<T>(path: string, body: unknown, headers?: Record<string, string>): Promise<T> {
  const res = await fetch(BASE + path, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json', ...headers },
    body: JSON.stringify(body),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error ?? res.statusText)
  }
  return res.json()
}

function adminHeaders(): Record<string, string> {
  const token = getAdminToken()
  return token ? { Authorization: `Bearer ${token}` } : {}
}

// ── existing ──────────────────────────────────────────────────────────────────

export const fetchWords = (): Promise<Word[]> => get('/words')
export const fetchWord = (id: number): Promise<Word> => get(`/words/${id}`)
export const fetchParagraphs = (): Promise<Paragraph[]> => get('/paragraphs')
export const fetchParagraphWithWords = (id: number): Promise<ParagraphWithWords> =>
  get(`/paragraphs/${id}/words`)

export async function generateParagraph(wordIds: number[], context: string): Promise<Paragraph> {
  return post('/generate/paragraph', { word_ids: wordIds, context })
}

// ── languages ─────────────────────────────────────────────────────────────────

export const fetchLanguages = (): Promise<Language[]> => get('/languages')

// ── admin auth ────────────────────────────────────────────────────────────────

export async function adminLogin(username: string, password: string): Promise<void> {
  const data = await post<{ token: string }>('/admin/login', { username, password })
  localStorage.setItem('admin_token', data.token)
}

export function adminLogout(): void {
  localStorage.removeItem('admin_token')
}

export function isAdminLoggedIn(): boolean {
  return !!getAdminToken()
}

// ── admin courses ─────────────────────────────────────────────────────────────

export function listAdminCourses(params?: {
  language?: string; level?: string; skill?: string; status?: string
}): Promise<Course[]> {
  const q = new URLSearchParams(params as Record<string, string> ?? {}).toString()
  return get(`/admin/courses${q ? `?${q}` : ''}`, adminHeaders())
}

export function getAdminCourse(id: number): Promise<Course> {
  return get(`/admin/courses/${id}`, adminHeaders())
}

export function generateCourse(body: {
  language_code: string; level: string; skills: string[]; topic: string
}): Promise<Course> {
  return post('/admin/courses/generate', body, adminHeaders())
}

export function patchCourse(id: number, body: Partial<Pick<Course, 'Title' | 'Description' | 'Topic'>>): Promise<Course> {
  return patch(`/admin/courses/${id}`, body, adminHeaders())
}

export function publishCourse(id: number): Promise<Course> {
  return post(`/admin/courses/${id}/publish`, undefined, adminHeaders())
}

export function archiveCourse(id: number): Promise<Course> {
  return post(`/admin/courses/${id}/archive`, undefined, adminHeaders())
}

export function generateLessons(id: number): Promise<Course> {
  return post(`/admin/courses/${id}/generate-lessons`, undefined, adminHeaders())
}

// ── admin grammar suggestions ─────────────────────────────────────────────────

export function getGrammarSuggestions(language: string, level: string): Promise<GrammarSuggestion[]> {
  return get(`/admin/grammar-suggestions?language=${language}&level=${level}`, adminHeaders())
}

export function analyzeGrammarSuggestions(language: string, level: string): Promise<GrammarSuggestion[]> {
  return post(`/admin/grammar-suggestions/analyze?language=${language}&level=${level}`, undefined, adminHeaders())
}

export function acceptSuggestion(id: number): Promise<Course> {
  return post(`/admin/grammar-suggestions/${id}/accept`, undefined, adminHeaders())
}

export async function dismissSuggestion(id: number): Promise<void> {
  const res = await fetch(`/api/admin/grammar-suggestions/${id}/dismiss`, {
    method: 'POST',
    headers: adminHeaders(),
  })
  if (!res.ok) throw new Error('Failed to dismiss')
}

// ── learner courses ───────────────────────────────────────────────────────────

export function listPublishedCourses(params?: { language?: string; level?: string }): Promise<Course[]> {
  const q = new URLSearchParams(params as Record<string, string> ?? {}).toString()
  return get(`/courses${q ? `?${q}` : ''}`)
}

export function getCourse(id: number): Promise<Course> {
  return get(`/courses/${id}`)
}

export function getLesson(courseId: number, lessonId: number): Promise<Lesson & { current_attempt?: LessonAttempt }> {
  return get(`/courses/${courseId}/lessons/${lessonId}`, {
    'X-Learner-Token': getLearnerToken(),
  })
}

// ── learner attempts ──────────────────────────────────────────────────────────

export function startAttempt(courseId: number, lessonId: number): Promise<LessonAttempt> {
  return post(`/courses/${courseId}/lessons/${lessonId}/attempts`, undefined, {
    'X-Learner-Token': getLearnerToken(),
  })
}

export function submitAnswer(
  courseId: number, lessonId: number, attemptId: number,
  exerciseId: number, answer: string,
): Promise<AnswerFeedback> {
  return post(
    `/courses/${courseId}/lessons/${lessonId}/attempts/${attemptId}/answer`,
    { exercise_id: exerciseId, answer },
    { 'X-Learner-Token': getLearnerToken() },
  ) as unknown as Promise<AnswerFeedback>
}

export function completeAttempt(courseId: number, lessonId: number, attemptId: number): Promise<LessonAttempt> {
  return post(
    `/courses/${courseId}/lessons/${lessonId}/attempts/${attemptId}/complete`,
    undefined,
    { 'X-Learner-Token': getLearnerToken() },
  )
}
