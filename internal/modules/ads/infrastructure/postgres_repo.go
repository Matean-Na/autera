package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"autera/internal/modules/ads/domain"
)

type PostgresRepo struct{ db *sql.DB }

func NewPostgresRepo(db *sql.DB) *PostgresRepo { return &PostgresRepo{db: db} }

func (r *PostgresRepo) Create(ctx context.Context, ad *domain.Ad) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO ads (seller_id, brand, model, year, mileage, price, vin, city, status, inspection_status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id
	`,
		ad.SellerID, ad.Brand, ad.Model, ad.Year, ad.Mileage, ad.Price, ad.VIN, ad.City, string(ad.Status), string(ad.InspectionState),
	).Scan(&id)
	return id, err
}

func (r *PostgresRepo) Get(ctx context.Context, id int64) (*domain.Ad, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, seller_id, brand, model, year, mileage, price, vin, city, status, inspection_status
		FROM ads WHERE id=$1
	`, id)

	var ad domain.Ad
	var st, ins string
	if err := row.Scan(&ad.ID, &ad.SellerID, &ad.Brand, &ad.Model, &ad.Year, &ad.Mileage, &ad.Price, &ad.VIN, &ad.City, &st, &ins); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("ad not found")
		}
		return nil, err
	}
	ad.Status = domain.AdStatus(st)
	ad.InspectionState = domain.InspectionStatus(ins)
	return &ad, nil
}

func (r *PostgresRepo) List(ctx context.Context, f domain.ListFilter) ([]domain.Ad, int64, error) {
	// MVP: без динамического SQL билдера — позже заменишь на sql builder
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, seller_id, brand, model, year, mileage, price, vin, city, status, inspection_status
		FROM ads
		ORDER BY id DESC
		LIMIT $1 OFFSET $2
	`, f.Limit, f.Offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []domain.Ad
	for rows.Next() {
		var ad domain.Ad
		var st, ins string
		if err := rows.Scan(&ad.ID, &ad.SellerID, &ad.Brand, &ad.Model, &ad.Year, &ad.Mileage, &ad.Price, &ad.VIN, &ad.City, &st, &ins); err != nil {
			return nil, 0, err
		}
		ad.Status = domain.AdStatus(st)
		ad.InspectionState = domain.InspectionStatus(ins)
		items = append(items, ad)
	}

	var total int64
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM ads`).Scan(&total)

	return items, total, nil
}

func (r *PostgresRepo) SubmitToModeration(ctx context.Context, adID, sellerID int64) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE ads SET status='moderation'
		WHERE id=$1 AND seller_id=$2 AND status IN ('draft','rejected')
	`, adID, sellerID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("cannot submit to moderation")
	}
	return nil
}

func (r *PostgresRepo) Moderate(ctx context.Context, adID int64, decision string) error {
	status := "rejected"
	if decision == "approve" {
		status = "published"
	}
	_, err := r.db.ExecContext(ctx, `UPDATE ads SET status=$2 WHERE id=$1`, adID, status)
	return err
}
