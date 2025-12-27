package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"autera/internal/modules/reports/domain"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{
		db: db,
	}
}

func (r *PostgresRepo) GetByAdID(ctx context.Context, adID int64) (*domain.Report, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT rp.id, rp.inspection_id, rp.total_score, rp.label
		FROM reports rp
		JOIN inspections i ON i.id = rp.inspection_id
		WHERE i.ad_id = $1
		ORDER BY rp.id DESC
		LIMIT 1
	`, adID)

	var rep domain.Report
	if err := row.Scan(&rep.ID, &rep.InspectionID, &rep.TotalScore, &rep.Label); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("report not found")
		}
		return nil, err
	}
	return &rep, nil
}

func (r *PostgresRepo) Dashboard(ctx context.Context) (map[string]any, error) {
	// MVP-заглушка: считаем количество проверок
	var inspections int64
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM inspections`).Scan(&inspections)

	return map[string]any{
		"inspections_total": inspections,
	}, nil
}
