package application

import (
	"autera/pkg/auth"
	"context"
	"errors"
	"time"
)

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

	// проверяем token_version (если сменили роли/пароль/заблокировали)
	tv, err := s.repo.GetTokenVersion(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if tv != claims.TokenVersion {
		return nil, errors.New("token version outdated")
	}

	// проверяем refresh в БД (существует, не ревокнут, не просрочен, hash совпал)
	hash, revokedAt, exp, userID, deviceID, err := s.repo.GetRefreshTokenHash(ctx, claims.ID)
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
	if hash != auth.HashToken(in.RefreshToken) {
		return nil, errors.New("refresh invalid")
	}

	// rotate: отзываем старый refresh
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
