package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	HTTPAddress    string
	AllowedOrigins []string
	Keycloak       Keycloak
}

type Keycloak struct {
	URL       string
	Realm     string
	ClientID  string
	Scope     string
	IssuerURL string
}

func Load() (Config, error) {
	return load(os.Getenv)
}

func load(getenv func(string) string) (Config, error) {
	k := Keycloak{
		URL:       strings.TrimRight(strings.TrimSpace(getenv("COSIGHT_KEYCLOAK_URL")), "/"),
		Realm:     strings.TrimSpace(getenv("COSIGHT_KEYCLOAK_REALM")),
		ClientID:  strings.TrimSpace(getenv("COSIGHT_KEYCLOAK_CLIENT_ID")),
		Scope:     valueOr(getenv("COSIGHT_KEYCLOAK_SCOPE"), "openid profile email"),
		IssuerURL: strings.TrimRight(strings.TrimSpace(getenv("COSIGHT_KEYCLOAK_ISSUER")), "/"),
	}
	if k.URL == "" || k.Realm == "" || k.ClientID == "" {
		return Config{}, fmt.Errorf("COSIGHT_KEYCLOAK_URL, COSIGHT_KEYCLOAK_REALM and COSIGHT_KEYCLOAK_CLIENT_ID are required")
	}
	parsed, err := url.Parse(k.URL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return Config{}, fmt.Errorf("COSIGHT_KEYCLOAK_URL must be an absolute URL")
	}
	if k.IssuerURL != "" {
		issuer, err := url.Parse(k.IssuerURL)
		if err != nil || issuer.Scheme == "" || issuer.Host == "" {
			return Config{}, fmt.Errorf("COSIGHT_KEYCLOAK_ISSUER must be an absolute URL")
		}
	}

	origins := splitCSV(valueOr(getenv("COSIGHT_ALLOWED_ORIGINS"), "http://localhost:5173"))
	return Config{
		HTTPAddress:    valueOr(getenv("COSIGHT_HTTP_ADDRESS"), ":8080"),
		AllowedOrigins: origins,
		Keycloak:       k,
	}, nil
}

func (k Keycloak) Issuer() string {
	if k.IssuerURL != "" {
		return k.IssuerURL
	}
	return k.URL + "/realms/" + url.PathEscape(k.Realm)
}

func valueOr(value, fallback string) string {
	if value = strings.TrimSpace(value); value != "" {
		return value
	}
	return fallback
}

func splitCSV(value string) []string {
	var values []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			values = append(values, item)
		}
	}
	return values
}
