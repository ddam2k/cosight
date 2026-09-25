package httpapi

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/wasming/cosight/internal/auth"
	"github.com/wasming/cosight/internal/config"
)

func NewRouter(cfg config.Config, verifier auth.TokenVerifier, dependencies Dependencies) http.Handler {
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
	users := userHandler{issuer: cfg.Keycloak.Issuer(), clientID: cfg.Keycloak.ClientID, repository: dependencies.Users}
	projects := projectHandler{users: users, repository: dependencies.Projects}
	authorized.GET("/me", users.current)
	authorized.GET("/projects", projects.list)
	authorized.POST("/projects", projects.create)
	return r
}

func writeError(c *gin.Context, status int, code, message string, fieldErrors map[string]string) {
	details := gin.H{}
	if len(fieldErrors) > 0 {
		details["fieldErrors"] = fieldErrors
	}
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message, "details": details}})
}

func internalError(c *gin.Context, err error) {
	slog.ErrorContext(c.Request.Context(), "request failed", "error", err, "method", c.Request.Method, "path", c.Request.URL.Path)
	writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "요청을 처리하지 못했습니다.", nil)
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
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
