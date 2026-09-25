package model

import (
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	ProjectScopePersonal     = "personal"
	ProjectScopeOrganization = "organization"
	RepositoryLocalPath      = "local_path"
	RepositoryGitURL         = "git_url"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,62}$`)

type Project struct {
	ID                   string     `db:"id" json:"id"`
	Scope                string     `db:"scope" json:"scope"`
	OrganizationID       *string    `db:"organization_id" json:"organizationId"`
	Name                 string     `db:"name" json:"name"`
	Slug                 string     `db:"slug" json:"slug"`
	Description          *string    `db:"description" json:"description"`
	RepositoryAlias      string     `db:"repository_alias" json:"repositoryAlias"`
	RepositorySourceType string     `db:"repository_source_type" json:"-"`
	RepositoryPath       *string    `db:"repository_path" json:"-"`
	RepositoryURL        *string    `db:"repository_url" json:"-"`
	RepositoryCredential *string    `db:"repository_credential_ref" json:"-"`
	DefaultBranch        *string    `db:"default_branch" json:"defaultBranch"`
	CurrentCommit        *string    `db:"current_commit" json:"currentCommit"`
	Status               string     `db:"status" json:"status"`
	Role                 string     `db:"role" json:"role"`
	ActiveIndexVersion   int64      `db:"active_index_version" json:"activeIndexVersion"`
	Version              int64      `db:"version" json:"version"`
	CreatedAt            time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt            time.Time  `db:"updated_at" json:"updatedAt"`
	DeletedAt            *time.Time `db:"deleted_at" json:"-"`
}

type CreateProject struct {
	Scope                string   `json:"scope"`
	OrganizationID       *string  `json:"organizationId"`
	Name                 string   `json:"name"`
	Slug                 string   `json:"slug"`
	Description          *string  `json:"description"`
	RepositoryAlias      string   `json:"repositoryAlias"`
	RepositorySourceType string   `json:"repositorySourceType"`
	RepositoryPath       *string  `json:"repositoryPath"`
	RepositoryURL        *string  `json:"repositoryUrl"`
	CredentialRef        *string  `json:"credentialRef"`
	DefaultBranch        *string  `json:"defaultBranch"`
	ExcludePatterns      []string `json:"excludePatterns"`
}

func (input *CreateProject) Normalize() {
	input.Scope = strings.TrimSpace(input.Scope)
	input.Name = strings.TrimSpace(input.Name)
	input.Slug = strings.TrimSpace(input.Slug)
	input.RepositoryAlias = strings.TrimSpace(input.RepositoryAlias)
	input.RepositorySourceType = strings.TrimSpace(input.RepositorySourceType)
	input.OrganizationID = trimmedPointer(input.OrganizationID)
	input.Description = trimmedPointer(input.Description)
	input.RepositoryPath = trimmedPointer(input.RepositoryPath)
	input.RepositoryURL = trimmedPointer(input.RepositoryURL)
	input.CredentialRef = trimmedPointer(input.CredentialRef)
	input.DefaultBranch = trimmedPointer(input.DefaultBranch)
	patterns := make([]string, 0, len(input.ExcludePatterns))
	for _, pattern := range input.ExcludePatterns {
		if pattern = strings.TrimSpace(pattern); pattern != "" {
			patterns = append(patterns, pattern)
		}
	}
	input.ExcludePatterns = patterns
}

func (input CreateProject) Validate() map[string]string {
	errors := make(map[string]string)
	if input.Scope != ProjectScopePersonal && input.Scope != ProjectScopeOrganization {
		errors["scope"] = "개인 또는 조직 프로젝트 범위를 선택해 주세요."
	}
	if input.Scope == ProjectScopePersonal && input.OrganizationID != nil {
		errors["organizationId"] = "개인 프로젝트에는 조직을 지정할 수 없습니다."
	}
	if input.Scope == ProjectScopeOrganization && input.OrganizationID == nil {
		errors["organizationId"] = "조직 프로젝트에는 조직이 필요합니다."
	}
	if len(input.Name) < 1 || len([]rune(input.Name)) > 120 {
		errors["name"] = "프로젝트 이름은 1~120자로 입력해 주세요."
	}
	if !slugPattern.MatchString(input.Slug) {
		errors["slug"] = "slug는 2~63자의 영문 소문자, 숫자, 하이픈만 사용할 수 있습니다."
	}
	if input.Description != nil && len([]rune(*input.Description)) > 2000 {
		errors["description"] = "설명은 2,000자 이하로 입력해 주세요."
	}
	if input.RepositoryAlias == "" || len([]rune(input.RepositoryAlias)) > 120 {
		errors["repositoryAlias"] = "저장소 표시 이름은 1~120자로 입력해 주세요."
	}
	switch input.RepositorySourceType {
	case RepositoryLocalPath:
		if input.RepositoryPath == nil || !strings.HasPrefix(*input.RepositoryPath, "/") {
			errors["repositoryPath"] = "서버 저장소의 절대 경로를 입력해 주세요."
		}
		if input.RepositoryURL != nil {
			errors["repositoryUrl"] = "서버 경로 방식에는 Git URL을 지정할 수 없습니다."
		}
	case RepositoryGitURL:
		if input.RepositoryURL == nil || !validRepositoryURL(*input.RepositoryURL) {
			errors["repositoryUrl"] = "HTTPS 또는 SSH 형식의 Git 저장소 URL을 입력해 주세요."
		}
		if input.RepositoryPath != nil {
			errors["repositoryPath"] = "Git URL 방식에는 서버 경로를 지정할 수 없습니다."
		}
	default:
		errors["repositorySourceType"] = "저장소 연결 방식을 선택해 주세요."
	}
	if len(input.ExcludePatterns) > 100 {
		errors["excludePatterns"] = "제외 패턴은 최대 100개까지 입력할 수 있습니다."
	}
	return errors
}

func validRepositoryURL(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && parsed.Host != "" && (parsed.Scheme == "https" || parsed.Scheme == "ssh")
}

func trimmedPointer(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
