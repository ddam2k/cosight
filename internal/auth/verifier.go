package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
)

const ClaimsContextKey = "auth.claims"

type Claims struct {
	Subject           string               `json:"sub"`
	Email             string               `json:"email"`
	Name              string               `json:"name"`
	PreferredUsername string               `json:"preferred_username"`
	AuthorizedParty   string               `json:"azp"`
	RealmAccess       RealmAccess          `json:"realm_access"`
	ResourceAccess    map[string]RoleGrant `json:"resource_access"`
}

type RealmAccess struct {
	Roles []string `json:"roles"`
}
type RoleGrant struct {
	Roles []string `json:"roles"`
}

type TokenVerifier interface {
	Verify(context.Context, string) (Claims, error)
}

type OIDCVerifier struct {
	verifier *oidc.IDTokenVerifier
	clientID string
}

func NewVerifier(ctx context.Context, issuer, clientID string) (*OIDCVerifier, error) {
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("discover OIDC provider: %w", err)
	}
	return &OIDCVerifier{
		verifier: provider.Verifier(&oidc.Config{SkipClientIDCheck: true}),
		clientID: clientID,
	}, nil
}

func (v *OIDCVerifier) Verify(ctx context.Context, rawToken string) (Claims, error) {
	token, err := v.verifier.Verify(ctx, rawToken)
	if err != nil {
		return Claims{}, err
	}
	var claims Claims
	if err := token.Claims(&claims); err != nil {
		return Claims{}, fmt.Errorf("decode claims: %w", err)
	}
	if claims.AuthorizedParty != v.clientID && !contains(token.Audience, v.clientID) {
		return Claims{}, errors.New("token was not issued for this client")
	}
	if claims.Subject == "" {
		return Claims{}, errors.New("token has no subject")
	}
	return claims, nil
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func Middleware(verifier TokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.Fields(header)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "AUTH_TOKEN_REQUIRED", "message": "Bearer token is required"})
			return
		}
		claims, err := verifier.Verify(c.Request.Context(), parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "AUTH_TOKEN_INVALID", "message": "Access token is invalid or expired"})
			return
		}
		c.Set(ClaimsContextKey, claims)
		c.Next()
	}
}
