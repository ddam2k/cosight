# Cosight MVP PostgreSQL schema

> 상태: Draft 0.1  
> 기준: PostgreSQL 16 이상

## 원칙

- UUID는 `gen_random_uuid()`를 사용한다.
- 시간은 `timestamptz` UTC로 저장한다.
- 프로젝트 데이터는 항상 `project_id`와 `index_version`으로 제한한다.
- API transaction은 `SET LOCAL app.user_id`, `app.project_id`, `app.organization_id`를 설정한다.
- 애플리케이션 DB 계정에는 `BYPASSRLS`를 부여하지 않는다.
- migration은 expand → backfill → contract 순서로 수행한다.

## Extensions와 enum

```sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TYPE user_status AS ENUM ('active', 'suspended');
CREATE TYPE organization_status AS ENUM ('active', 'suspended');
CREATE TYPE project_scope AS ENUM ('personal', 'organization');
CREATE TYPE project_status AS ENUM ('draft', 'active', 'error', 'deleting', 'deleted');
CREATE TYPE project_role AS ENUM ('project_admin', 'developer', 'viewer');
CREATE TYPE invitation_status AS ENUM ('pending', 'accepted', 'rejected', 'revoked', 'expired');
CREATE TYPE index_status AS ENUM ('building', 'active', 'superseded', 'failed', 'cancelled');
CREATE TYPE job_status AS ENUM ('queued', 'running', 'cancel_requested', 'completed', 'failed', 'cancelled');
CREATE TYPE confidence_level AS ENUM ('confirmed', 'probable', 'inferred');
```

## 사용자와 세션

```sql
CREATE TABLE users (
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
CREATE UNIQUE INDEX users_email_ci_uq
  ON users (lower(email)) WHERE email IS NOT NULL;

CREATE TABLE user_sessions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  session_hash text NOT NULL UNIQUE,
  expires_at timestamptz NOT NULL,
  last_seen_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  revoked_at timestamptz
);
CREATE INDEX user_sessions_user_active_idx
  ON user_sessions (user_id, expires_at) WHERE revoked_at IS NULL;
```

## 조직

```sql
CREATE TABLE organizations (
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

CREATE TABLE organization_members (
  organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  assigned_by uuid NOT NULL REFERENCES users(id),
  assigned_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (organization_id, user_id)
);
CREATE INDEX organization_members_user_idx ON organization_members (user_id, organization_id);
```

## 프로젝트와 멤버십

```sql
CREATE TABLE projects (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  scope project_scope NOT NULL,
  organization_id uuid REFERENCES organizations(id) ON DELETE RESTRICT,
  name text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 120),
  slug text NOT NULL CHECK (slug ~ '^[a-z0-9][a-z0-9-]{1,62}$'),
  description text CHECK (char_length(description) <= 2000),
  repository_alias text NOT NULL,
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
  CHECK (
    (scope = 'personal' AND organization_id IS NULL) OR
    (scope = 'organization' AND organization_id IS NOT NULL)
  ),
  CHECK (
    (repository_source_type = 'local_path' AND repository_path IS NOT NULL AND repository_url IS NULL) OR
    (repository_source_type = 'git_url' AND repository_url IS NOT NULL AND repository_path IS NULL)
  )
);
CREATE UNIQUE INDEX projects_org_slug_uq
  ON projects (organization_id, slug)
  WHERE scope = 'organization' AND deleted_at IS NULL;
CREATE UNIQUE INDEX projects_personal_slug_uq
  ON projects (created_by, slug)
  WHERE scope = 'personal' AND deleted_at IS NULL;

CREATE TABLE project_exclude_patterns (
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  position integer NOT NULL CHECK (position >= 0),
  pattern text NOT NULL CHECK (pattern <> ''),
  PRIMARY KEY (project_id, position)
);

CREATE TABLE project_members (
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  role project_role NOT NULL,
  created_by uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (project_id, user_id)
);
CREATE INDEX project_members_user_idx ON project_members (user_id, project_id);

CREATE TABLE project_invitations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  invitee_user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role project_role NOT NULL,
  status invitation_status NOT NULL DEFAULT 'pending',
  expires_at timestamptz NOT NULL,
  accepted_at timestamptz,
  rejected_at timestamptz,
  revoked_at timestamptz,
  created_by uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX project_invitations_pending_user_uq
  ON project_invitations (project_id, invitee_user_id)
  WHERE status = 'pending';
```

마지막 Project Admin과 조직 membership 제약은 여러 row를 검사하므로 service transaction에서 project row를 `FOR UPDATE`로 잠근 뒤 검증한다. 조직원 해제도 organization row와 영향을 받는 project row를 잠그고 대체 관리자 지정, 초대 취소와 membership 삭제를 같은 transaction에서 수행한다.

## 인덱스 버전과 코드 그래프

```sql
CREATE TABLE code_index_versions (
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  id bigint NOT NULL,
  git_commit text,
  snapshot_id text,
  status index_status NOT NULL,
  analysis_schema_version text NOT NULL,
  analyzer_versions jsonb NOT NULL,
  settings_hash text NOT NULL,
  file_count integer NOT NULL DEFAULT 0,
  node_count integer NOT NULL DEFAULT 0,
  edge_count integer NOT NULL DEFAULT 0,
  warning_count integer NOT NULL DEFAULT 0,
  started_at timestamptz NOT NULL DEFAULT now(),
  completed_at timestamptz,
  activated_at timestamptz,
  created_by uuid NOT NULL REFERENCES users(id),
  PRIMARY KEY (project_id, id)
);
CREATE UNIQUE INDEX code_index_versions_active_uq
  ON code_index_versions (project_id) WHERE status = 'active';

CREATE TABLE code_files (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id uuid NOT NULL,
  index_version bigint NOT NULL,
  path text NOT NULL,
  language text NOT NULL CHECK (language IN ('go', 'typescript', 'vue')),
  content_hash text NOT NULL,
  is_test boolean NOT NULL,
  is_generated boolean NOT NULL,
  parse_status text NOT NULL CHECK (parse_status IN ('pending', 'parsed', 'partial', 'failed', 'skipped')),
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (project_id, path, index_version),
  FOREIGN KEY (project_id, index_version)
    REFERENCES code_index_versions(project_id, id) ON DELETE CASCADE
);

CREATE TABLE code_nodes (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id uuid NOT NULL,
  index_version bigint NOT NULL,
  file_id uuid REFERENCES code_files(id) ON DELETE CASCADE,
  stable_key text NOT NULL,
  kind text NOT NULL,
  name text NOT NULL,
  qualified_name text,
  language text NOT NULL,
  start_line integer,
  start_column integer,
  end_line integer,
  end_column integer,
  parent_node_id uuid REFERENCES code_nodes(id) DEFERRABLE INITIALLY DEFERRED,
  metadata jsonb NOT NULL DEFAULT '{}',
  UNIQUE (project_id, stable_key, index_version),
  FOREIGN KEY (project_id, index_version)
    REFERENCES code_index_versions(project_id, id) ON DELETE CASCADE
);

CREATE TABLE code_edges (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id uuid NOT NULL,
  index_version bigint NOT NULL,
  source_node_id uuid NOT NULL REFERENCES code_nodes(id) ON DELETE CASCADE,
  target_node_id uuid NOT NULL REFERENCES code_nodes(id) ON DELETE CASCADE,
  kind text NOT NULL,
  confidence confidence_level NOT NULL,
  metadata jsonb NOT NULL DEFAULT '{}',
  FOREIGN KEY (project_id, index_version)
    REFERENCES code_index_versions(project_id, id) ON DELETE CASCADE
);
CREATE INDEX code_edges_source_idx ON code_edges (project_id, index_version, source_node_id, kind);
CREATE INDEX code_edges_target_idx ON code_edges (project_id, index_version, target_node_id, kind);

CREATE TABLE source_evidence (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id uuid NOT NULL,
  index_version bigint NOT NULL,
  edge_id uuid REFERENCES code_edges(id) ON DELETE CASCADE,
  node_id uuid REFERENCES code_nodes(id) ON DELETE CASCADE,
  analyzer text NOT NULL,
  analyzer_version text NOT NULL,
  file_id uuid NOT NULL REFERENCES code_files(id) ON DELETE CASCADE,
  start_line integer NOT NULL,
  start_column integer NOT NULL,
  end_line integer NOT NULL,
  end_column integer NOT NULL,
  reason text NOT NULL,
  CHECK ((edge_id IS NOT NULL) <> (node_id IS NOT NULL)),
  FOREIGN KEY (project_id, index_version)
    REFERENCES code_index_versions(project_id, id) ON DELETE CASCADE
);

CREATE TABLE analysis_warnings (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id uuid NOT NULL,
  index_version bigint NOT NULL,
  file_id uuid REFERENCES code_files(id) ON DELETE CASCADE,
  analyzer text NOT NULL,
  code text NOT NULL,
  severity text NOT NULL CHECK (severity IN ('info', 'warning', 'error')),
  message text NOT NULL,
  range jsonb,
  metadata jsonb NOT NULL DEFAULT '{}',
  created_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (project_id, index_version)
    REFERENCES code_index_versions(project_id, id) ON DELETE CASCADE
);
```

## 작업 queue

```sql
CREATE TABLE analysis_jobs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  kind text NOT NULL,
  status job_status NOT NULL DEFAULT 'queued',
  requested_by uuid NOT NULL REFERENCES users(id),
  requested_index_version bigint,
  payload jsonb NOT NULL DEFAULT '{}',
  progress numeric(5,4) NOT NULL DEFAULT 0 CHECK (progress BETWEEN 0 AND 1),
  stage text,
  attempt integer NOT NULL DEFAULT 0,
  max_attempts integer NOT NULL DEFAULT 3,
  lease_owner text,
  lease_expires_at timestamptz,
  heartbeat_at timestamptz,
  cancel_requested_at timestamptz,
  error_code text,
  error_detail jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  started_at timestamptz,
  completed_at timestamptz
);
CREATE UNIQUE INDEX analysis_jobs_one_active_index_uq
  ON analysis_jobs (project_id)
  WHERE kind = 'index' AND status IN ('queued', 'running', 'cancel_requested');
CREATE INDEX analysis_jobs_claim_idx ON analysis_jobs (status, created_at);
```

## 검색 document

```sql
CREATE TABLE search_documents (
  project_id uuid NOT NULL,
  index_version bigint NOT NULL,
  node_id uuid NOT NULL REFERENCES code_nodes(id) ON DELETE CASCADE,
  kind text NOT NULL,
  path text,
  name text NOT NULL,
  qualified_name text,
  normalized_text text NOT NULL,
  document tsvector GENERATED ALWAYS AS
    (setweight(to_tsvector('simple', coalesce(name, '')), 'A') ||
     setweight(to_tsvector('simple', coalesce(qualified_name, '')), 'B') ||
     setweight(to_tsvector('simple', coalesce(path, '')), 'C')) STORED,
  PRIMARY KEY (project_id, index_version, node_id)
);
CREATE INDEX search_documents_fts_idx ON search_documents USING gin (document);
CREATE INDEX search_documents_trgm_idx ON search_documents USING gin (normalized_text gin_trgm_ops);
```

## RLS 기본 정책

모든 project-scoped 테이블에 같은 정책 패턴을 적용한다.

```sql
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;
ALTER TABLE code_nodes ENABLE ROW LEVEL SECURITY;
ALTER TABLE code_edges ENABLE ROW LEVEL SECURITY;

CREATE POLICY projects_member_read ON projects
FOR SELECT USING (
  EXISTS (
    SELECT 1 FROM project_members pm
    WHERE pm.project_id = projects.id
      AND pm.user_id = current_setting('app.user_id')::uuid
  )
  AND (
    projects.scope = 'personal'
    OR EXISTS (
      SELECT 1 FROM organization_members om
      WHERE om.organization_id = projects.organization_id
        AND om.user_id = current_setting('app.user_id')::uuid
    )
  )
);
```

하위 테이블 정책은 `project_id`로 `projects` 정책과 동일한 membership 조건을 검사한다. Worker 전용 DB role은 사용자 RLS를 우회할 수 있지만 API role과 자격증명을 공유하지 않는다.

## Migration과 rollback

1. migration 파일은 순번과 설명을 가진다: `000001_create_identity.up.sql`.
2. destructive 변경은 한 배포에서 수행하지 않는다.
3. enum 값 삭제는 금지하고 새 enum 또는 text+check로 이행한다.
4. 대형 index는 `CREATE INDEX CONCURRENTLY`를 별도 migration으로 실행한다.
5. rollback이 데이터 손실을 만들면 down migration 대신 restore 절차와 compatibility release를 제공한다.
6. production 적용 전 최신 backup에서 restore rehearsal을 수행한다.
