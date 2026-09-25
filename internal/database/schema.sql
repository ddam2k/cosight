DO $$ BEGIN CREATE TYPE user_status AS ENUM ('active', 'suspended'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE organization_status AS ENUM ('active', 'suspended'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE project_scope AS ENUM ('personal', 'organization'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE project_status AS ENUM ('draft', 'active', 'error', 'deleting', 'deleted'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE project_role AS ENUM ('project_admin', 'developer', 'viewer'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;

CREATE TABLE IF NOT EXISTS users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  issuer text NOT NULL,
  subject text NOT NULL,
  email text,
  display_name text NOT NULL,
  status user_status NOT NULL DEFAULT 'active',
  last_login_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (issuer, subject)
);
CREATE UNIQUE INDEX IF NOT EXISTS users_email_ci_uq ON users (lower(email)) WHERE email IS NOT NULL;

CREATE TABLE IF NOT EXISTS organizations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 120),
  slug text NOT NULL CHECK (slug ~ '^[a-z0-9][a-z0-9-]{1,62}$'),
  status organization_status NOT NULL DEFAULT 'active',
  version bigint NOT NULL DEFAULT 1,
  created_by uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (slug)
);

CREATE TABLE IF NOT EXISTS organization_members (
  organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  assigned_by uuid NOT NULL REFERENCES users(id),
  assigned_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (organization_id, user_id)
);
CREATE INDEX IF NOT EXISTS organization_members_user_idx ON organization_members (user_id, organization_id);

CREATE TABLE IF NOT EXISTS projects (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  scope project_scope NOT NULL,
  organization_id uuid REFERENCES organizations(id) ON DELETE RESTRICT,
  name text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 120),
  slug text NOT NULL CHECK (slug ~ '^[a-z0-9][a-z0-9-]{1,62}$'),
  description text CHECK (char_length(description) <= 2000),
  repository_alias text NOT NULL CHECK (char_length(repository_alias) BETWEEN 1 AND 120),
  repository_source_type text NOT NULL CHECK (repository_source_type IN ('local_path', 'git_url')),
  repository_path text,
  repository_url text,
  repository_credential_ref text,
  default_branch text,
  current_commit text,
  status project_status NOT NULL DEFAULT 'draft',
  active_index_version bigint NOT NULL DEFAULT 0,
  version bigint NOT NULL DEFAULT 1,
  created_by uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  CHECK ((scope = 'personal' AND organization_id IS NULL) OR (scope = 'organization' AND organization_id IS NOT NULL)),
  CHECK ((repository_source_type = 'local_path' AND repository_path IS NOT NULL AND repository_url IS NULL) OR (repository_source_type = 'git_url' AND repository_url IS NOT NULL AND repository_path IS NULL))
);
CREATE UNIQUE INDEX IF NOT EXISTS projects_org_slug_uq ON projects (organization_id, slug) WHERE scope = 'organization' AND deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS projects_personal_slug_uq ON projects (created_by, slug) WHERE scope = 'personal' AND deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS project_exclude_patterns (
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  position integer NOT NULL CHECK (position >= 0),
  pattern text NOT NULL CHECK (pattern <> ''),
  PRIMARY KEY (project_id, position)
);

CREATE TABLE IF NOT EXISTS project_members (
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  role project_role NOT NULL,
  created_by uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (project_id, user_id)
);
CREATE INDEX IF NOT EXISTS project_members_user_idx ON project_members (user_id, project_id);
