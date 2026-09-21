package httpapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/wasming/cosight/internal/auth"
	"github.com/wasming/cosight/internal/config"
)

func NewRouter(cfg config.Config, verifier auth.TokenVerifier) http.Handler {
	r := gin.New()
	r.Use(gin.Recovery(), cors(cfg.AllowedOrigins))

	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/api/v1/public/auth-config", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusOK, gin.H{
			"url":      cfg.Keycloak.URL,
			"realm":    cfg.Keycloak.Realm,
			"clientId": cfg.Keycloak.ClientID,
			"scope":    cfg.Keycloak.Scope,
		})
	})

	authorized := r.Group("/api/v1")
	authorized.Use(auth.Middleware(verifier))
	authorized.GET("/me", func(c *gin.Context) {
		claims := c.MustGet(auth.ClaimsContextKey).(auth.Claims)
		c.JSON(http.StatusOK, gin.H{
			"id":                claims.Subject,
			"email":             claims.Email,
			"displayName":       claims.Name,
			"preferredUsername": claims.PreferredUsername,
			"realmRoles":        claims.RealmAccess.Roles,
			"clientRoles":       claims.ResourceAccess[cfg.Keycloak.ClientID].Roles,
		})
	})
	return r
}

func cors(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[origin] = struct{}{}
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, ok := allowed[origin]; ok {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		}
		if c.Request.Method == http.MethodOptions {
			if origin == "" || !contains(allowedOrigins, origin) {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}
	return false
}
