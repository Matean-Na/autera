package domain

import "context"

type Repository interface {
	Create(ctx context.Context, u *User) (int64, error)
	GetByPhoneOrEmail(ctx context.Context, login string) (*User, error)
	GetRoles(ctx context.Context, userID int64) ([]Role, error)
}
