package application

import (
	"autera/internal/modules/users/domain"
	"context"
	"errors"
)

type SetRolesInput struct {
	Roles []string `json:"roles"`
}

func parseRolesStrict(in []string) ([]domain.Role, error) {
	allowed := map[string]domain.Role{
		string(domain.RoleBuyer):     domain.RoleBuyer,
		string(domain.RoleSeller):    domain.RoleSeller,
		string(domain.RoleInspector): domain.RoleInspector,
		string(domain.RoleAdmin):     domain.RoleAdmin,
		string(domain.RoleOwner):     domain.RoleOwner,
	}

	out := make([]domain.Role, 0, len(in))
	for _, r := range in {
		role, ok := allowed[r]
		if !ok {
			return nil, errors.New("invalid role: " + r)
		}
		out = append(out, role)
	}
	return out, nil
}

// Админ: нельзя назначать admin/owner
func (s *Service) SetRolesByAdmin(ctx context.Context, targetUserID int64, in SetRolesInput) error {
	roles, err := parseRolesStrict(in.Roles)
	if err != nil {
		return err
	}
	for _, r := range roles {
		if r == domain.RoleAdmin || r == domain.RoleOwner {
			return errors.New("admin cannot grant admin/owner")
		}
	}

	if err := s.repo.ReplaceRoles(ctx, targetUserID, roles); err != nil {
		return err
	}
	if err := s.repo.IncrementTokenVersion(ctx, targetUserID); err != nil {
		return err
	}
	return s.repo.RevokeRefreshTokensByUser(ctx, targetUserID)
}

// Owner: может назначать admin
func (s *Service) SetRolesByOwner(ctx context.Context, targetUserID int64, in SetRolesInput) error {
	roles, err := parseRolesStrict(in.Roles)
	if err != nil {
		return err
	}
	for _, r := range roles {
		if r == domain.RoleOwner {
			return errors.New("owner role should be managed manually")
		}
	}

	if err := s.repo.ReplaceRoles(ctx, targetUserID, roles); err != nil {
		return err
	}
	if err := s.repo.IncrementTokenVersion(ctx, targetUserID); err != nil {
		return err
	}
	return s.repo.RevokeRefreshTokensByUser(ctx, targetUserID)
}

func (s *Service) SetActiveByAdmin(ctx context.Context, targetUserID int64, active bool) error {
	if err := s.repo.SetActive(ctx, targetUserID, active); err != nil {
		return err
	}
	if err := s.repo.IncrementTokenVersion(ctx, targetUserID); err != nil {
		return err
	}
	return s.repo.RevokeRefreshTokensByUser(ctx, targetUserID)
}
