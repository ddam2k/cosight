package httpapi

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wasming/cosight/internal/auth"
	"github.com/wasming/cosight/internal/model"
	"github.com/wasming/cosight/internal/repository"
)

type projectHandler struct {
	users      userHandler
	repository repository.ProjectRepository
}

func (handler projectHandler) list(c *gin.Context) {
	claims := c.MustGet(auth.ClaimsContextKey).(auth.Claims)
	user, err := handler.users.ensure(c, claims)
	if err != nil {
		internalError(c, err)
		return
	}
	projects, err := handler.repository.ListByUser(c.Request.Context(), user.ID)
	if err != nil {
		internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": projects, "nextCursor": nil})
}

func (handler projectHandler) create(c *gin.Context) {
	var input model.CreateProject
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_REQUEST", "요청 JSON 형식을 확인해 주세요.", nil)
		return
	}
	input.Normalize()
	if fieldErrors := input.Validate(); len(fieldErrors) > 0 {
		writeError(c, http.StatusBadRequest, "VALIDATION_FAILED", "입력한 프로젝트 정보를 확인해 주세요.", fieldErrors)
		return
	}
	claims := c.MustGet(auth.ClaimsContextKey).(auth.Claims)
	user, err := handler.users.ensure(c, claims)
	if err != nil {
		internalError(c, err)
		return
	}
	project, err := handler.repository.Create(c.Request.Context(), user.ID, input)
	switch {
	case err == nil:
		c.JSON(http.StatusCreated, project)
	case errors.Is(err, repository.ErrProjectSlugConflict):
		writeError(c, http.StatusConflict, "PROJECT_SLUG_CONFLICT", "같은 범위에 이미 사용 중인 slug입니다.", map[string]string{"slug": "같은 범위에 이미 사용 중인 slug입니다."})
	case errors.Is(err, repository.ErrOrganizationMembership):
		writeError(c, http.StatusForbidden, "ORGANIZATION_MEMBERSHIP_REQUIRED", "현재 소속된 조직에만 프로젝트를 만들 수 있습니다.", map[string]string{"organizationId": "현재 소속된 조직을 선택해 주세요."})
	default:
		internalError(c, err)
	}
}
