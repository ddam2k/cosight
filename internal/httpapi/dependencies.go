package httpapi

import "github.com/wasming/cosight/internal/repository"

type Dependencies struct {
	Users    repository.UserRepository
	Projects repository.ProjectRepository
}
