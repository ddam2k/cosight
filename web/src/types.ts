export interface CurrentUser {
  id: string
  email: string
  displayName: string
  preferredUsername: string
  realmRoles: string[]
  clientRoles: string[]
}
