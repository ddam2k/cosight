package config

import "testing"

func TestLoadKeycloakConfig(t *testing.T) {
	values := map[string]string{
		"COSIGHT_KEYCLOAK_URL":       "https://sso.example.com/",
		"COSIGHT_KEYCLOAK_REALM":     "cosight",
		"COSIGHT_KEYCLOAK_CLIENT_ID": "cosight-web",
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

func TestLoadRequiresPublicKeycloakSettings(t *testing.T) {
	_, err := load(func(string) string { return "" })
	if err == nil {
		t.Fatal("expected required setting error")
	}
}
