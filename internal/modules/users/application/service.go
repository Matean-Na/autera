package application

import (
	"context"
	"errors"

	"autera/internal/modules/users/domain"
	"autera/pkg/auth"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo domain.Repository
	jwt  *auth.JWT
}

func NewService(repo domain.Repository, jwt *auth.JWT) *Service {
	return &Service{repo: repo, jwt: jwt}
}

type RegisterInput struct {
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Type     string `json:"type"` // person/company
	Role     string `json:"role"` // buyer/seller/inspector/admin/owner (для MVP можно ограничить)
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (int64, error) {
	if in.Password == "" {
		return 0, errors.New("password required")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	u := &domain.User{
		Phone:        in.Phone,
		Email:        in.Email,
		PasswordHash: string(hash),
		Type:         in.Type,
		Roles:        []domain.Role{domain.Role(in.Role)},
	}
	return s.repo.Create(ctx, u)
}

type LoginInput struct {
	Login    string `json:"login"` // phone or email
	Password string `json:"password"`
}

func (s *Service) Login(ctx context.Context, in LoginInput) (string, error) {
	u, err := s.repo.GetByPhoneOrEmail(ctx, in.Login)
	if err != nil {
		return "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(in.Password)); err != nil {
		return "", errors.New("invalid credentials")
	}
	roles, err := s.repo.GetRoles(ctx, u.ID)
	if err != nil {
		return "", err
	}
	rs := make([]string, 0, len(roles))
	for _, r := range roles {
		rs = append(rs, string(r))
	}
	return s.jwt.Issue(u.ID, rs)
}
