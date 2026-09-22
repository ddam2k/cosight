export interface CurrentUser {
  id: string
  email: string
  displayName: string
  preferredUsername: string
  realmRoles: string[]
  clientRoles: string[]
  organizations: Organization[]
}

export interface Organization {
  id: string
  name: string
  slug: string
  status: 'active' | 'suspended'
}

export type ProjectScope = 'personal' | 'organization'
export type RepositorySourceType = 'local_path' | 'git_url'

export interface CreateProjectRequest {
  scope: ProjectScope
  organizationId: string | null
  name: string
  slug: string
  description?: string | null
  repositoryAlias: string
  repositorySourceType: RepositorySourceType
  repositoryPath?: string
  repositoryUrl?: string
  credentialRef?: string | null
  defaultBranch?: string | null
  excludePatterns: string[]
  [key: string]: string | string[] | null | undefined
}

export interface Project {
  id: string
  scope: ProjectScope
  organizationId: string | null
  name: string
  slug: string
  description: string | null
  repositoryAlias: string
  defaultBranch: string | null
  status: 'draft' | 'active' | 'error' | 'deleting' | 'deleted'
  role: 'project_admin' | 'developer' | 'viewer'
  activeIndexVersion: number
  version: number
}

export type CurrentUserResponse = Omit<CurrentUser, 'realmRoles' | 'clientRoles' | 'organizations'> & {
  realmRoles?: string[] | null
  clientRoles?: string[] | null
  organizations?: Organization[] | null
}

export function normalizeCurrentUser(response: CurrentUserResponse): CurrentUser {
  return {
    ...response,
    realmRoles: Array.isArray(response.realmRoles) ? response.realmRoles : [],
    clientRoles: Array.isArray(response.clientRoles) ? response.clientRoles : [],
    organizations: Array.isArray(response.organizations) ? response.organizations : [],
  }
}
