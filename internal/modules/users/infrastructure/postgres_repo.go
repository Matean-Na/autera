package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"autera/internal/modules/users/domain"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{
		db: db,
	}
}

func (r *PostgresRepo) Create(ctx context.Context, u *domain.User) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	var id int64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO users (phone, email, password_hash, type)
		VALUES ($1,$2,$3,$4)
		RETURNING id
	`, u.Phone, u.Email, u.PasswordHash, u.Type).Scan(&id)
	if err != nil {
		return 0, err
	}

	for _, role := range u.Roles {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO user_roles (user_id, role) VALUES ($1,$2)
		`, id, string(role))
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *PostgresRepo) GetByPhoneOrEmail(ctx context.Context, login string) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, phone, email, password_hash, type
		FROM users
		WHERE phone=$1 OR email=$1
	`, login)

	var u domain.User
	if err := row.Scan(&u.ID, &u.Phone, &u.Email, &u.PasswordHash, &u.Type); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &u, nil
}

func (r *PostgresRepo) GetRoles(ctx context.Context, userID int64) ([]domain.Role, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT role FROM user_roles WHERE user_id=$1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []domain.Role
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, domain.Role(role))
	}
	return roles, nil
}
