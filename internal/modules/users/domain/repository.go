package domain

import (
	"context"
	"time"
)

type Repository interface {
	// base
	Create(ctx context.Context, u *User) (int64, error)
	GetByPhoneOrEmail(ctx context.Context, login string) (*User, error)
	GetByID(ctx context.Context, id int64) (*User, error)
	GetRoles(ctx context.Context, userID int64) ([]Role, error)

	// security flags
	GetTokenVersion(ctx context.Context, userID int64) (int64, error)
	IncrementTokenVersion(ctx context.Context, userID int64) error
	IsActive(ctx context.Context, userID int64) (bool, error)
	SetActive(ctx context.Context, userID int64, active bool) error

	// user management
	UpdatePasswordHash(ctx context.Context, userID int64, hash string) error
	ReplaceRoles(ctx context.Context, userID int64, roles []Role) error

	// refresh tokens
	SaveRefreshToken(ctx context.Context, userID int64, jti, tokenHash, deviceID string, expiresAt time.Time) error
	GetRefreshToken(ctx context.Context, jti string) (tokenHash string, revokedAt *time.Time, expiresAt time.Time, userID int64, deviceID string, err error)
	RevokeRefreshToken(ctx context.Context, jti string) error
	RevokeRefreshTokensByUser(ctx context.Context, userID int64) error
	RevokeRefreshTokensByUserDevice(ctx context.Context, userID int64, deviceID string) error
}
