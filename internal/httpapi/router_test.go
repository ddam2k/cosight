package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wasming/cosight/internal/auth"
	"github.com/wasming/cosight/internal/config"
)

type fakeVerifier struct {
	claims auth.Claims
	err    error
}

func (f fakeVerifier) Verify(context.Context, string) (auth.Claims, error) { return f.claims, f.err }

func testConfig() config.Config {
	return config.Config{Keycloak: config.Keycloak{URL: "https://sso.example.com", Realm: "cosight", ClientID: "cosight-web", Scope: "openid profile email"}, AllowedOrigins: []string{"http://localhost:5173"}}
}

func TestPublicAuthConfig(t *testing.T) {
	r := NewRouter(testConfig(), fakeVerifier{})
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
	r := NewRouter(testConfig(), fakeVerifier{err: errors.New("invalid")})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestMeReturnsVerifiedClaims(t *testing.T) {
	r := NewRouter(testConfig(), fakeVerifier{claims: auth.Claims{Subject: "user-1", Email: "dev@example.com", Name: "개발자"}})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer valid")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", w.Code, w.Body.String())
	}
}
