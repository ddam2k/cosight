package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/wasming/cosight/internal/model"
)

type UserRepository interface {
	Ensure(context.Context, model.UserIdentity) (model.User, error)
	Organizations(context.Context, string) ([]model.Organization, error)
}

type SQLXUserRepository struct{ db *sqlx.DB }

func NewSQLXUserRepository(db *sqlx.DB) *SQLXUserRepository {
	return &SQLXUserRepository{db: db}
}

func (repository *SQLXUserRepository) Ensure(ctx context.Context, identity model.UserIdentity) (model.User, error) {
	var user model.User
	email := nullableString(identity.Email)
	displayName := strings.TrimSpace(identity.DisplayName)
	if displayName == "" {
		displayName = identity.Subject
	}
	err := repository.db.GetContext(ctx, &user, `
		INSERT INTO users (issuer, subject, email, display_name, last_login_at)
		VALUES ($1, $2, $3, $4, now())
		ON CONFLICT (issuer, subject) DO UPDATE SET
			email = EXCLUDED.email,
			display_name = EXCLUDED.display_name,
			last_login_at = now(),
			updated_at = now()
		RETURNING id, issuer, subject, email, display_name, status, last_login_at`,
		identity.Issuer, identity.Subject, email, displayName,
	)
	if err != nil {
		return model.User{}, fmt.Errorf("ensure user: %w", err)
	}
	return user, nil
}

func (repository *SQLXUserRepository) Organizations(ctx context.Context, userID string) ([]model.Organization, error) {
	organizations := make([]model.Organization, 0)
	err := repository.db.SelectContext(ctx, &organizations, `
		SELECT o.id, o.name, o.slug, o.status
		FROM organizations o
		JOIN organization_members om ON om.organization_id = o.id
		WHERE om.user_id = $1 AND o.status = 'active'
		ORDER BY lower(o.name), o.id`, userID)
	if err != nil {
		return nil, fmt.Errorf("list user organizations: %w", err)
	}
	return organizations, nil
}

func nullableString(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}
