package postgres

import (
	"context"

	"github.com/SternKater/carsharing/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CarsRepository struct {
	pool *pgxpool.Pool
}

func NewCarsRepository(p *pgxpool.Pool) *CarsRepository {
	return &CarsRepository{pool: p}
}

// only 'free' cars(KarKar)
func (r *CarsRepository) GetAvailableCars(ctx context.Context) ([]*domain.Car, error) {
	query := `
		SELECT id, car_name, car_number, status, price_per_minute 
		FROM cars 
		WHERE status = 'free'
	`

	rows, err := getExecutor(ctx, r.pool).Query(ctx, query)
	if err != nil {
		return nil, domain.ErrInternal
	}
	defer rows.Close()

	var cars []*domain.Car
	for rows.Next() {
		var car domain.Car
		if err := rows.Scan(&car.ID, &car.CarName, &car.CarNumber, &car.Status, &car.PricePerMinute); err != nil {
			return nil, domain.ErrInternal
		}
		cars = append(cars, &car)
	}

	if err := rows.Err(); err != nil {
		return nil, domain.ErrInternal
	}

	return cars, nil
}