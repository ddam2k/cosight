import { accessToken } from './auth'

interface ErrorPayload {
  code?: string
  message?: string
  fieldErrors?: Record<string, string>
  details?: { fieldErrors?: Record<string, string> }
  error?: {
    code?: string
    message?: string
    details?: { fieldErrors?: Record<string, string> }
  }
}

export class ApiError extends Error {
  constructor(
    message: string,
    public readonly status: number,
    public readonly code = '',
    public readonly fieldErrors: Record<string, string> = {},
  ) { super(message) }
}

export async function apiFetch<T>(path: string, init: RequestInit = {}): Promise<T> {
  const token = await accessToken()
  const headers = new Headers(init.headers)
  headers.set('Authorization', `Bearer ${token}`)
  if (init.body && !headers.has('Content-Type')) headers.set('Content-Type', 'application/json')
  const response = await fetch(path, { ...init, headers })
  if (response.status === 401) throw new ApiError('로그인이 만료되었습니다. 다시 로그인해 주세요.', 401)
  if (!response.ok) {
    let payload: ErrorPayload = {}
    try { payload = await response.json() as ErrorPayload } catch { /* empty or non-JSON response */ }
    const fallback = response.status === 409 ? '같은 범위에 이미 사용 중인 slug입니다.' : `API 요청에 실패했습니다. (${response.status})`
    const apiError = payload.error
    throw new ApiError(
      apiError?.message || payload.message || fallback,
      response.status,
      apiError?.code || payload.code,
      payload.fieldErrors || payload.details?.fieldErrors || apiError?.details?.fieldErrors || {},
    )
  }
  if (response.status === 204) return undefined as T
  return response.json() as Promise<T>
}
