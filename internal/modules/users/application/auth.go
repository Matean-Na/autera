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
	Login    string `json:"login"` // phone or email
	Password string `json:"password"`
	DeviceID string `json:"device_id"` // required: web/mobile device key
}

type LoginOutput struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func (s *Service) Login(ctx context.Context, in LoginInput) (*LoginOutput, error) {
	if in.DeviceID == "" {
		return nil, errors.New("device_id required")
	}

	u, err := s.repo.GetByPhoneOrEmail(ctx, in.Login)
	if err != nil {
		return nil, err
	}

	if !u.IsActive {
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
	Role     string `json:"role"` // only buyer/seller
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

type RefreshInput struct {
	RefreshToken string `json:"refresh_token"`
	DeviceID     string `json:"device_id"`
}

type RefreshOutput struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func (s *Service) Refresh(ctx context.Context, in RefreshInput) (*RefreshOutput, error) {
	if in.DeviceID == "" {
		return nil, errors.New("device_id required")
	}

	claims, err := s.jwt.Parse(in.RefreshToken)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != auth.TokenRefresh {
		return nil, errors.New("not a refresh token")
	}
	if claims.DeviceID != in.DeviceID {
		return nil, errors.New("device mismatch")
	}

	active, err := s.repo.IsActive(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if !active {
		return nil, errors.New("user blocked")
	}

	tv, err := s.repo.GetTokenVersion(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if tv != claims.TokenVersion {
		return nil, errors.New("token outdated")
	}

	storedHash, revokedAt, exp, userID, deviceID, err := s.repo.GetRefreshToken(ctx, claims.ID)
	if err != nil {
		return nil, err
	}
	if revokedAt != nil {
		return nil, errors.New("refresh revoked")
	}
	if time.Now().After(exp) {
		return nil, errors.New("refresh expired")
	}
	if userID != claims.UserID || deviceID != in.DeviceID {
		return nil, errors.New("refresh mismatch")
	}
	if storedHash != auth.HashToken(in.RefreshToken) {
		return nil, errors.New("refresh invalid")
	}

	// rotate: revoke old refresh
	if err := s.repo.RevokeRefreshToken(ctx, claims.ID); err != nil {
		return nil, err
	}

	roles, err := s.repo.GetRoles(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	rs := make([]string, 0, len(roles))
	for _, r := range roles {
		rs = append(rs, string(r))
	}

	access, _, accessExp, err := s.jwt.IssueAccess(claims.UserID, rs, tv, in.DeviceID)
	if err != nil {
		return nil, err
	}

	newRefresh, newJTI, newRefreshExp, err := s.jwt.IssueRefresh(claims.UserID, tv, in.DeviceID)
	if err != nil {
		return nil, err
	}
	if err := s.repo.SaveRefreshToken(ctx, claims.UserID, newJTI, auth.HashToken(newRefresh), in.DeviceID, newRefreshExp); err != nil {
		return nil, err
	}

	return &RefreshOutput{AccessToken: access, RefreshToken: newRefresh, ExpiresAt: accessExp}, nil
}

type LogoutInput struct {
	DeviceID string `json:"device_id"`
}

func (s *Service) Logout(ctx context.Context, userID int64, in LogoutInput) error {
	if in.DeviceID == "" {
		return errors.New("device_id required")
	}
	return s.repo.RevokeRefreshTokensByUserDevice(ctx, userID, in.DeviceID)
}
