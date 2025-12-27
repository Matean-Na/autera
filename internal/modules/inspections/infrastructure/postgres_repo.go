package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"autera/internal/modules/inspections/domain"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{
		db: db,
	}
}

func (r *PostgresRepo) Request(ctx context.Context, adID, sellerID int64) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO inspections (ad_id, seller_id, status)
		VALUES ($1,$2,'requested')
		RETURNING id
	`, adID, sellerID).Scan(&id)
	return id, err
}

func (r *PostgresRepo) Assign(ctx context.Context, inspectionID, inspectorID int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE inspections SET inspector_id=$2, status='assigned'
		WHERE id=$1 AND status IN ('requested','assigned')
	`, inspectionID, inspectorID)
	return err
}

func (r *PostgresRepo) ListAssigned(ctx context.Context, inspectorID int64) ([]domain.Inspection, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, ad_id, seller_id, inspector_id, status
		FROM inspections
		WHERE inspector_id=$1
		ORDER BY id DESC
	`, inspectorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.Inspection
	for rows.Next() {
		var it domain.Inspection
		var inspID sql.NullInt64
		var st string
		if err := rows.Scan(&it.ID, &it.AdID, &it.SellerID, &inspID, &st); err != nil {
			return nil, err
		}
		if inspID.Valid {
			v := inspID.Int64
			it.InspectorID = &v
		}
		it.Status = domain.Status(st)
		items = append(items, it)
	}
	return items, nil
}

func (r *PostgresRepo) Submit(ctx context.Context, inspectionID, inspectorID int64) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE inspections SET status='submitted'
		WHERE id=$1 AND inspector_id=$2 AND status IN ('assigned','in_progress')
	`, inspectionID, inspectorID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("cannot submit")
	}
	return nil
}
