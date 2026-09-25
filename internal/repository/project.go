package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/wasming/cosight/internal/model"
)

var (
	ErrProjectSlugConflict    = errors.New("project slug already exists")
	ErrOrganizationMembership = errors.New("user is not an active organization member")
)

type ProjectRepository interface {
	Create(context.Context, string, model.CreateProject) (model.Project, error)
	ListByUser(context.Context, string) ([]model.Project, error)
}

type SQLXProjectRepository struct{ db *sqlx.DB }

func NewSQLXProjectRepository(db *sqlx.DB) *SQLXProjectRepository {
	return &SQLXProjectRepository{db: db}
}

const projectColumns = `p.id, p.scope, p.organization_id, p.name, p.slug, p.description,
	p.repository_alias, p.repository_source_type, p.repository_path, p.repository_url,
	p.repository_credential_ref, p.default_branch, p.current_commit, p.status,
	pm.role, p.active_index_version, p.version, p.created_at, p.updated_at, p.deleted_at`

func (repository *SQLXProjectRepository) Create(ctx context.Context, creatorID string, input model.CreateProject) (project model.Project, resultErr error) {
	tx, err := repository.db.BeginTxx(ctx, nil)
	if err != nil {
		return model.Project{}, fmt.Errorf("begin create project: %w", err)
	}
	defer func() {
		if resultErr != nil {
			_ = tx.Rollback()
		}
	}()

	if input.Scope == model.ProjectScopeOrganization {
		var allowed bool
		err = tx.GetContext(ctx, &allowed, `
			SELECT EXISTS (
				SELECT 1 FROM organization_members om
				JOIN organizations o ON o.id = om.organization_id
				WHERE om.organization_id = $1 AND om.user_id = $2 AND o.status = 'active'
			)`, input.OrganizationID, creatorID)
		if err != nil {
			return model.Project{}, fmt.Errorf("check organization membership: %w", err)
		}
		if !allowed {
			return model.Project{}, ErrOrganizationMembership
		}
	}

	var projectID string
	err = tx.GetContext(ctx, &projectID, `
		INSERT INTO projects (
			scope, organization_id, name, slug, description, repository_alias,
			repository_source_type, repository_path, repository_url,
			repository_credential_ref, default_branch, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id`, input.Scope, input.OrganizationID, input.Name, input.Slug,
		input.Description, input.RepositoryAlias, input.RepositorySourceType,
		input.RepositoryPath, input.RepositoryURL, input.CredentialRef,
		input.DefaultBranch, creatorID)
	if err != nil {
		if isSlugConflict(err) {
			return model.Project{}, ErrProjectSlugConflict
		}
		return model.Project{}, fmt.Errorf("insert project: %w", err)
	}

	for position, pattern := range input.ExcludePatterns {
		if _, err = tx.ExecContext(ctx, `INSERT INTO project_exclude_patterns (project_id, position, pattern) VALUES ($1, $2, $3)`, projectID, position, pattern); err != nil {
			return model.Project{}, fmt.Errorf("insert project exclude pattern: %w", err)
		}
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO project_members (project_id, user_id, role, created_by) VALUES ($1, $2, 'project_admin', $2)`, projectID, creatorID); err != nil {
		return model.Project{}, fmt.Errorf("insert project administrator: %w", err)
	}
	if err = tx.GetContext(ctx, &project, `SELECT `+projectColumns+` FROM projects p JOIN project_members pm ON pm.project_id = p.id AND pm.user_id = $2 WHERE p.id = $1`, projectID, creatorID); err != nil {
		return model.Project{}, fmt.Errorf("read created project: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return model.Project{}, fmt.Errorf("commit create project: %w", err)
	}
	return project, nil
}

func (repository *SQLXProjectRepository) ListByUser(ctx context.Context, userID string) ([]model.Project, error) {
	projects := make([]model.Project, 0)
	err := repository.db.SelectContext(ctx, &projects, `
		SELECT `+projectColumns+`
		FROM projects p
		JOIN project_members pm ON pm.project_id = p.id
		WHERE pm.user_id = $1 AND p.deleted_at IS NULL
		ORDER BY p.updated_at DESC, p.id`, userID)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	return projects, nil
}

func isSlugConflict(err error) bool {
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != "23505" {
		return false
	}
	return postgresError.ConstraintName == "projects_org_slug_uq" || postgresError.ConstraintName == "projects_personal_slug_uq"
}
