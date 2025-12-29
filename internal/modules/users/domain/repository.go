package domain

import (
	"context"
	"time"
)

type Repository interface {
	Create(ctx context.Context, u *User) (int64, error)
	GetByPhoneOrEmail(ctx context.Context, login string) (*User, error)
	GetRoles(ctx context.Context, userID int64) ([]Role, error)

	GetTokenVersion(ctx context.Context, userID int64) (int64, error)
	IncrementTokenVersion(ctx context.Context, userID int64) error
	IsActive(ctx context.Context, userID int64) (bool, error)

	SaveRefreshToken(ctx context.Context, userID int64, jti, tokenHash, deviceID string, expiresAt time.Time) error
	GetRefreshTokenHash(ctx context.Context, jti string) (tokenHash string, revokedAt *time.Time, expiresAt time.Time, userID int64, deviceID string, err error)
	RevokeRefreshToken(ctx context.Context, jti string) error
	RevokeRefreshTokensByUserDevice(ctx context.Context, userID int64, deviceID string) error
}
