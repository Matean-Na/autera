package application

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

type ChangePasswordInput struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
	DeviceID    string `json:"device_id"`
	LogoutAll   bool   `json:"logout_all"`
}

func (s *Service) ChangePassword(ctx context.Context, userID int64, in ChangePasswordInput) error {
	if in.NewPassword == "" {
		return errors.New("new_password required")
	}

	u, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if !u.IsActive {
		return errors.New("user blocked")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(in.OldPassword)); err != nil {
		return errors.New("old password invalid")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.repo.UpdatePasswordHash(ctx, userID, string(hash)); err != nil {
		return err
	}

	// invalidate everything
	if err := s.repo.IncrementTokenVersion(ctx, userID); err != nil {
		return err
	}

	if in.LogoutAll || in.DeviceID == "" {
		return s.repo.RevokeRefreshTokensByUser(ctx, userID)
	}
	return s.repo.RevokeRefreshTokensByUserDevice(ctx, userID, in.DeviceID)
}
