package application

import (
	"autera/internal/modules/users/domain"
	"autera/pkg/auth"
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type LoginInput struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	DeviceID string `json:"device_id"` // "web:chrome", "mobile:android", UUID девайса
}

type LoginOutput struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func (s *Service) Login(ctx context.Context, in LoginInput) (*LoginOutput, error) {
	u, err := s.repo.GetByPhoneOrEmail(ctx, in.Login)
	if err != nil {
		return nil, err
	}

	active, err := s.repo.IsActive(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	if !active {
		return nil, errors.New("user blocked")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(in.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	roles, err := s.repo.GetRoles(ctx, u.ID)
	if err != nil {
		return nil, err
	}

	tv, err := s.repo.GetTokenVersion(ctx, u.ID)
	if err != nil {
		return nil, err
	}

	rs := make([]string, 0, len(roles))
	for _, r := range roles {
		rs = append(rs, string(r))
	}

	access, _, accessExp, err := s.jwt.IssueAccess(u.ID, rs, tv, in.DeviceID)
	if err != nil {
		return nil, err
	}

	refresh, refreshJTI, refreshExp, err := s.jwt.IssueRefresh(u.ID, tv, in.DeviceID)
	if err != nil {
		return nil, err
	}

	// храним hash refresh токена (не сам токен)
	if err := s.repo.SaveRefreshToken(ctx, u.ID, refreshJTI, auth.HashToken(refresh), in.DeviceID, refreshExp); err != nil {
		return nil, err
	}

	return &LoginOutput{AccessToken: access, RefreshToken: refresh, ExpiresAt: accessExp}, nil
}

type RegisterInput struct {
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Type     string `json:"type"` // person/company

	Role string `json:"role"` // buyer/seller (только эти)
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (int64, error) {
	if in.Password == "" {
		return 0, errors.New("password required")
	}

	allowed := map[string]domain.Role{
		string(domain.RoleBuyer):  domain.RoleBuyer,
		string(domain.RoleSeller): domain.RoleSeller,
	}
	role, ok := allowed[in.Role]
	if !ok {
		return 0, errors.New("invalid role")
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
		Roles:        []domain.Role{role},
	}
	return s.repo.Create(ctx, u)
}
