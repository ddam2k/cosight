package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/wasming/cosight/internal/auth"
	"github.com/wasming/cosight/internal/config"
	"github.com/wasming/cosight/internal/model"
)

type fakeVerifier struct {
	claims auth.Claims
	err    error
}

func (f fakeVerifier) Verify(context.Context, string) (auth.Claims, error) { return f.claims, f.err }

type fakeUserRepository struct{}

func (fakeUserRepository) Ensure(_ context.Context, identity model.UserIdentity) (model.User, error) {
	email := identity.Email
	return model.User{ID: "00000000-0000-0000-0000-000000000001", Email: &email, DisplayName: identity.DisplayName, Status: "active"}, nil
}
func (fakeUserRepository) Organizations(context.Context, string) ([]model.Organization, error) {
	return []model.Organization{}, nil
}

type fakeProjectRepository struct {
	created *model.CreateProject
}

func (repository *fakeProjectRepository) Create(_ context.Context, _ string, input model.CreateProject) (model.Project, error) {
	repository.created = &input
	return model.Project{ID: "00000000-0000-0000-0000-000000000010", Scope: input.Scope, Name: input.Name, Slug: input.Slug, RepositoryAlias: input.RepositoryAlias, Status: "draft", Role: "project_admin", Version: 1}, nil
}
func (*fakeProjectRepository) ListByUser(context.Context, string) ([]model.Project, error) {
	return []model.Project{}, nil
}

func testDependencies() Dependencies {
	return Dependencies{Users: fakeUserRepository{}, Projects: &fakeProjectRepository{}}
}

func testConfig() config.Config {
	return config.Config{Keycloak: config.Keycloak{URL: "https://sso.example.com", Realm: "cosight", ClientID: "cosight-web", Scope: "openid profile email"}, AllowedOrigins: []string{"http://localhost:5173"}}
}

func TestPublicAuthConfig(t *testing.T) {
	r := NewRouter(testConfig(), fakeVerifier{}, testDependencies())
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/auth-config", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["realm"] != "cosight" || body["clientId"] != "cosight-web" {
		t.Fatalf("unexpected response: %#v", body)
	}
}

func TestMeRequiresBearerToken(t *testing.T) {
	r := NewRouter(testConfig(), fakeVerifier{err: errors.New("invalid")}, testDependencies())
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestMeReturnsVerifiedClaims(t *testing.T) {
	r := NewRouter(testConfig(), fakeVerifier{claims: auth.Claims{Subject: "user-1", Email: "dev@example.com", Name: "개발자"}}, testDependencies())
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer valid")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", w.Code, w.Body.String())
	}
	var body struct {
		RealmRoles    []string         `json:"realmRoles"`
		ClientRoles   []string         `json:"clientRoles"`
		Organizations []map[string]any `json:"organizations"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.RealmRoles == nil || body.ClientRoles == nil || body.Organizations == nil {
		t.Fatalf("collection fields must not be null: %s", w.Body.String())
	}
}

func TestCreateProject(t *testing.T) {
	projects := &fakeProjectRepository{}
	r := NewRouter(testConfig(), fakeVerifier{claims: auth.Claims{Subject: "user-1", Name: "개발자"}}, Dependencies{Users: fakeUserRepository{}, Projects: projects})
	body := `{"scope":"personal","name":"Cosight","slug":"cosight","repositoryAlias":"main","repositorySourceType":"local_path","repositoryPath":"/repositories/cosight","excludePatterns":["dist/**"]}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid")
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		response, _ := io.ReadAll(w.Body)
		t.Fatalf("status = %d body=%s", w.Code, response)
	}
	if projects.created == nil || projects.created.Slug != "cosight" || len(projects.created.ExcludePatterns) != 1 {
		t.Fatalf("unexpected repository input: %#v", projects.created)
	}
}

func TestCreateProjectReturnsFieldValidation(t *testing.T) {
	r := NewRouter(testConfig(), fakeVerifier{claims: auth.Claims{Subject: "user-1"}}, testDependencies())
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", strings.NewReader(`{"scope":"personal","name":"","slug":"INVALID"}`))
	req.Header.Set("Authorization", "Bearer valid")
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "fieldErrors") {
		t.Fatalf("status = %d body=%s", w.Code, w.Body.String())
	}
}
