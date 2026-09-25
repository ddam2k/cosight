package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddress    string
	AllowedOrigins []string
	Keycloak       Keycloak
	PostgreSQL     PostgreSQL
}

type PostgreSQL struct {
	URL             string
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
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

	postgres, err := loadPostgreSQL(getenv)
	if err != nil {
		return Config{}, err
	}

	origins := splitCSV(valueOr(getenv("COSIGHT_ALLOWED_ORIGINS"), "http://localhost:5173"))
	return Config{
		HTTPAddress:    valueOr(getenv("COSIGHT_HTTP_ADDRESS"), ":8080"),
		AllowedOrigins: origins,
		Keycloak:       k,
		PostgreSQL:     postgres,
	}, nil
}

func loadPostgreSQL(getenv func(string) string) (PostgreSQL, error) {
	postgres := PostgreSQL{
		URL:             strings.TrimSpace(getenv("COSIGHT_POSTGRES_URL")),
		Host:            valueOr(getenv("COSIGHT_POSTGRES_HOST"), "127.0.0.1"),
		Port:            5432,
		User:            strings.TrimSpace(getenv("COSIGHT_POSTGRES_USER")),
		Password:        getenv("COSIGHT_POSTGRES_PASSWORD"),
		Database:        strings.TrimSpace(getenv("COSIGHT_POSTGRES_DATABASE")),
		SSLMode:         valueOr(getenv("COSIGHT_POSTGRES_SSLMODE"), "require"),
		MaxOpenConns:    20,
		MaxIdleConns:    5,
		ConnMaxLifetime: 30 * time.Minute,
	}
	var err error
	if postgres.MaxOpenConns, err = intSetting(getenv("COSIGHT_POSTGRES_MAX_OPEN_CONNS"), postgres.MaxOpenConns); err != nil {
		return PostgreSQL{}, fmt.Errorf("COSIGHT_POSTGRES_MAX_OPEN_CONNS: %w", err)
	}
	if postgres.MaxIdleConns, err = intSetting(getenv("COSIGHT_POSTGRES_MAX_IDLE_CONNS"), postgres.MaxIdleConns); err != nil {
		return PostgreSQL{}, fmt.Errorf("COSIGHT_POSTGRES_MAX_IDLE_CONNS: %w", err)
	}
	if postgres.ConnMaxLifetime, err = durationSetting(getenv("COSIGHT_POSTGRES_CONN_MAX_LIFETIME"), postgres.ConnMaxLifetime); err != nil {
		return PostgreSQL{}, fmt.Errorf("COSIGHT_POSTGRES_CONN_MAX_LIFETIME: %w", err)
	}
	if postgres.MaxOpenConns < 1 || postgres.MaxIdleConns < 0 || postgres.MaxIdleConns > postgres.MaxOpenConns {
		return PostgreSQL{}, fmt.Errorf("PostgreSQL pool sizes must satisfy max_open >= 1 and 0 <= max_idle <= max_open")
	}

	if postgres.URL != "" {
		parsed, parseErr := url.Parse(postgres.URL)
		if parseErr != nil || (parsed.Scheme != "postgres" && parsed.Scheme != "postgresql") || parsed.Host == "" || strings.TrimPrefix(parsed.Path, "/") == "" {
			return PostgreSQL{}, fmt.Errorf("COSIGHT_POSTGRES_URL must be a valid postgres URL")
		}
		return postgres, nil
	}
	if postgres.Port, err = intSetting(getenv("COSIGHT_POSTGRES_PORT"), postgres.Port); err != nil {
		return PostgreSQL{}, fmt.Errorf("COSIGHT_POSTGRES_PORT: %w", err)
	}
	if postgres.User == "" || postgres.Password == "" || postgres.Database == "" {
		return PostgreSQL{}, fmt.Errorf("COSIGHT_POSTGRES_USER, COSIGHT_POSTGRES_PASSWORD and COSIGHT_POSTGRES_DATABASE are required when COSIGHT_POSTGRES_URL is not set")
	}
	if postgres.Port < 1 || postgres.Port > 65535 {
		return PostgreSQL{}, fmt.Errorf("COSIGHT_POSTGRES_PORT must be between 1 and 65535")
	}
	return postgres, nil
}

func (p PostgreSQL) DSN() string {
	if p.URL != "" {
		return p.URL
	}
	dsn := &url.URL{Scheme: "postgres", Host: fmt.Sprintf("%s:%d", p.Host, p.Port), Path: p.Database}
	dsn.User = url.UserPassword(p.User, p.Password)
	query := dsn.Query()
	query.Set("sslmode", p.SSLMode)
	dsn.RawQuery = query.Encode()
	return dsn.String()
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

func intSetting(value string, fallback int) (int, error) {
	if strings.TrimSpace(value) == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("must be an integer")
	}
	return parsed, nil
}

func durationSetting(value string, fallback time.Duration) (time.Duration, error) {
	if strings.TrimSpace(value) == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("must be a positive duration such as 30m")
	}
	return parsed, nil
}
