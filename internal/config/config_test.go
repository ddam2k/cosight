package config

import "testing"

func TestLoadKeycloakConfig(t *testing.T) {
	values := map[string]string{
		"COSIGHT_KEYCLOAK_URL":       "https://sso.example.com/",
		"COSIGHT_KEYCLOAK_REALM":     "cosight",
		"COSIGHT_KEYCLOAK_CLIENT_ID": "cosight-web",
		"COSIGHT_POSTGRES_URL":       "postgres://cosight:secret@db.example.com:5432/cosight?sslmode=require",
	}
	cfg, err := load(func(key string) string { return values[key] })
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if got := cfg.Keycloak.Issuer(); got != "https://sso.example.com/realms/cosight" {
		t.Fatalf("unexpected issuer: %s", got)
	}
	if cfg.Keycloak.Scope != "openid profile email" {
		t.Fatalf("unexpected default scope: %s", cfg.Keycloak.Scope)
	}
}

func TestLoadPostgreSQLFields(t *testing.T) {
	values := map[string]string{
		"COSIGHT_KEYCLOAK_URL":       "https://sso.example.com",
		"COSIGHT_KEYCLOAK_REALM":     "cosight",
		"COSIGHT_KEYCLOAK_CLIENT_ID": "cosight-web",
		"COSIGHT_POSTGRES_HOST":      "db.internal",
		"COSIGHT_POSTGRES_PORT":      "5433",
		"COSIGHT_POSTGRES_USER":      "cosight",
		"COSIGHT_POSTGRES_PASSWORD":  "p@ss word",
		"COSIGHT_POSTGRES_DATABASE":  "cosight_app",
		"COSIGHT_POSTGRES_SSLMODE":   "verify-full",
	}
	cfg, err := load(func(key string) string { return values[key] })
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	want := "postgres://cosight:p%40ss%20word@db.internal:5433/cosight_app?sslmode=verify-full"
	if got := cfg.PostgreSQL.DSN(); got != want {
		t.Fatalf("DSN = %q, want %q", got, want)
	}
}

func TestLoadRequiresPublicKeycloakSettings(t *testing.T) {
	_, err := load(func(string) string { return "" })
	if err == nil {
		t.Fatal("expected required setting error")
	}
}

func TestLoadRequiresPostgreSQLSettings(t *testing.T) {
	values := map[string]string{
		"COSIGHT_KEYCLOAK_URL":       "https://sso.example.com",
		"COSIGHT_KEYCLOAK_REALM":     "cosight",
		"COSIGHT_KEYCLOAK_CLIENT_ID": "cosight-web",
	}
	_, err := load(func(key string) string { return values[key] })
	if err == nil {
		t.Fatal("expected required PostgreSQL setting error")
	}
}
