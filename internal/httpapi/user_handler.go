package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wasming/cosight/internal/auth"
	"github.com/wasming/cosight/internal/model"
	"github.com/wasming/cosight/internal/repository"
)

type userHandler struct {
	issuer     string
	clientID   string
	repository repository.UserRepository
}

func (handler userHandler) current(c *gin.Context) {
	claims := c.MustGet(auth.ClaimsContextKey).(auth.Claims)
	user, err := handler.ensure(c, claims)
	if err != nil {
		internalError(c, err)
		return
	}
	organizations, err := handler.repository.Organizations(c.Request.Context(), user.ID)
	if err != nil {
		internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":                  user.ID,
		"email":               claims.Email,
		"displayName":         user.DisplayName,
		"preferredUsername":   claims.PreferredUsername,
		"status":              user.Status,
		"systemAdministrator": contains(claims.ResourceAccess[handler.clientID].Roles, "cosight-system-admin"),
		"realmRoles":          nonNilStrings(claims.RealmAccess.Roles),
		"clientRoles":         nonNilStrings(claims.ResourceAccess[handler.clientID].Roles),
		"organizations":       organizations,
	})
}

func (handler userHandler) ensure(c *gin.Context, claims auth.Claims) (model.User, error) {
	displayName := claims.Name
	if displayName == "" {
		displayName = claims.PreferredUsername
	}
	return handler.repository.Ensure(c.Request.Context(), model.UserIdentity{
		Issuer:      handler.issuer,
		Subject:     claims.Subject,
		Email:       claims.Email,
		DisplayName: displayName,
	})
}
