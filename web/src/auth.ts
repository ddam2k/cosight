import Keycloak, { type KeycloakInitOptions } from 'keycloak-js'

export interface PublicAuthConfig {
  url: string
  realm: string
  clientId: string
  scope: string
}

let keycloak: Keycloak | undefined

export async function initializeAuth(): Promise<Keycloak> {
  const response = await fetch('/api/v1/public/auth-config', { cache: 'no-store' })
  if (!response.ok) throw new Error(`인증 설정을 불러오지 못했습니다. (${response.status})`)
  const config = await response.json() as PublicAuthConfig
  if (!config.url || !config.realm || !config.clientId) throw new Error('백엔드의 Keycloak 공개 설정이 올바르지 않습니다.')

  keycloak = new Keycloak({ url: config.url, realm: config.realm, clientId: config.clientId })
  const options: KeycloakInitOptions = {
    onLoad: 'login-required',
    checkLoginIframe: false,
    pkceMethod: 'S256',
    scope: config.scope,
  }
  const authenticated = await keycloak.init(options)
  if (!authenticated) await keycloak.login()
  return keycloak
}

export async function accessToken(): Promise<string> {
  if (!keycloak) throw new Error('Keycloak이 초기화되지 않았습니다.')
  await keycloak.updateToken(30)
  if (!keycloak.token) throw new Error('Keycloak Access Token이 없습니다.')
  return keycloak.token
}

export async function logout(): Promise<void> {
  if (!keycloak) return
  await keycloak.logout({ redirectUri: window.location.origin })
}
