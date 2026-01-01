package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

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
		INSERT INTO users (phone, email, password_hash, type, is_active, token_version)
		VALUES ($1,$2,$3,$4, TRUE, 0)
		RETURNING id
	`, u.Phone, u.Email, u.PasswordHash, u.Type).Scan(&id)
	if err != nil {
		return 0, err
	}

	for _, role := range u.Roles {
		_, err := tx.ExecContext(ctx, `INSERT INTO user_roles (user_id, role) VALUES ($1,$2)`, id, string(role))
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
		SELECT id, phone, email, password_hash, type, is_active, token_version
		FROM users
		WHERE phone=$1 OR email=$1
	`, login)

	var u domain.User
	if err := row.Scan(&u.ID, &u.Phone, &u.Email, &u.PasswordHash, &u.Type, &u.IsActive, &u.TokenVersion); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &u, nil
}

func (r *PostgresRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, phone, email, password_hash, type, is_active, token_version
		FROM users
		WHERE id=$1
	`, id)

	var u domain.User
	if err := row.Scan(&u.ID, &u.Phone, &u.Email, &u.PasswordHash, &u.Type, &u.IsActive, &u.TokenVersion); err != nil {
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

	roles := make([]domain.Role, 0)
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, domain.Role(role))
	}
	return roles, rows.Err()
}

func (r *PostgresRepo) GetTokenVersion(ctx context.Context, userID int64) (int64, error) {
	var v int64
	err := r.db.QueryRowContext(ctx, `SELECT token_version FROM users WHERE id=$1`, userID).Scan(&v)
	return v, err
}

func (r *PostgresRepo) IncrementTokenVersion(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET token_version = token_version + 1 WHERE id=$1`, userID)
	return err
}

func (r *PostgresRepo) IsActive(ctx context.Context, userID int64) (bool, error) {
	var a bool
	err := r.db.QueryRowContext(ctx, `SELECT is_active FROM users WHERE id=$1`, userID).Scan(&a)
	return a, err
}

func (r *PostgresRepo) SetActive(ctx context.Context, userID int64, active bool) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET is_active=$2 WHERE id=$1`, userID, active)
	return err
}

func (r *PostgresRepo) UpdatePasswordHash(ctx context.Context, userID int64, hash string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET password_hash=$2 WHERE id=$1`, userID, hash)
	return err
}

func (r *PostgresRepo) ReplaceRoles(ctx context.Context, userID int64, roles []domain.Role) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM user_roles WHERE user_id=$1`, userID); err != nil {
		return err
	}

	for _, role := range roles {
		if _, err := tx.ExecContext(ctx, `INSERT INTO user_roles (user_id, role) VALUES ($1,$2)`, userID, string(role)); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PostgresRepo) SaveRefreshToken(ctx context.Context, userID int64, jti, tokenHash, deviceID string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_refresh_tokens (user_id, jti, token_hash, device_id, expires_at)
		VALUES ($1,$2,$3,$4,$5)
	`, userID, jti, tokenHash, deviceID, expiresAt)
	return err
}

func (r *PostgresRepo) GetRefreshToken(ctx context.Context, jti string) (string, *time.Time, time.Time, int64, string, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT token_hash, revoked_at, expires_at, user_id, device_id
		FROM user_refresh_tokens
		WHERE jti=$1
	`, jti)

	var tokenHash string
	var revoked sql.NullTime
	var exp time.Time
	var userID int64
	var deviceID string

	if err := row.Scan(&tokenHash, &revoked, &exp, &userID, &deviceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil, time.Time{}, 0, "", errors.New("refresh not found")
		}
		return "", nil, time.Time{}, 0, "", err
	}

	var revokedAt *time.Time
	if revoked.Valid {
		revokedAt = &revoked.Time
	}
	return tokenHash, revokedAt, exp, userID, deviceID, nil
}

func (r *PostgresRepo) RevokeRefreshToken(ctx context.Context, jti string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE user_refresh_tokens
		SET revoked_at = now()
		WHERE jti=$1 AND revoked_at IS NULL
	`, jti)
	return err
}

func (r *PostgresRepo) RevokeRefreshTokensByUser(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE user_refresh_tokens
		SET revoked_at = now()
		WHERE user_id=$1 AND revoked_at IS NULL
	`, userID)
	return err
}

func (r *PostgresRepo) RevokeRefreshTokensByUserDevice(ctx context.Context, userID int64, deviceID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE user_refresh_tokens
		SET revoked_at = now()
		WHERE user_id=$1 AND device_id=$2 AND revoked_at IS NULL
	`, userID, deviceID)
	return err
}
