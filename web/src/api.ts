import { accessToken } from './auth'

export async function apiFetch<T>(path: string, init: RequestInit = {}): Promise<T> {
  const token = await accessToken()
  const headers = new Headers(init.headers)
  headers.set('Authorization', `Bearer ${token}`)
  if (init.body && !headers.has('Content-Type')) headers.set('Content-Type', 'application/json')
  const response = await fetch(path, { ...init, headers })
  if (response.status === 401) throw new Error('로그인이 만료되었습니다. 다시 로그인해 주세요.')
  if (!response.ok) throw new Error(`API 요청에 실패했습니다. (${response.status})`)
  return response.json() as Promise<T>
}
