package domain

import "time"

type Role string

const (
	RoleBuyer     Role = "buyer"
	RoleSeller    Role = "seller"
	RoleInspector Role = "inspector"
	RoleAdmin     Role = "admin"
	RoleOwner     Role = "owner"
)

type User struct {
	ID           int64
	Phone        string
	Email        string
	PasswordHash string
	Type         string // person/company
	CreatedAt    time.Time
	IsActive     bool
	TokenVersion int64
	Roles        []Role
}
