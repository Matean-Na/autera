package application

import (
	"autera/internal/modules/users/domain"
	"autera/pkg/auth"
)

type Service struct {
	repo domain.Repository
	jwt  *auth.JWT
}

func NewService(repo domain.Repository, jwt *auth.JWT) *Service {
	return &Service{repo: repo, jwt: jwt}
}
