package postgres

import (
	"context"
	"homework/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	saveSensorOwnerQuery = `
		INSERT INTO sensors_users (user_id, sensor_id)
		VALUES ($1, $2)
	`

	getSensorsByUserIDQuery = `
		SELECT user_id, sensor_id
		FROM sensors_users
		WHERE user_id = $1
	`
)

type SensorOwnerRepository struct {
	pool *pgxpool.Pool
}

func NewSensorOwnerRepository(pool *pgxpool.Pool) *SensorOwnerRepository {
	return &SensorOwnerRepository{
		pool,
	}
}

func (r *SensorOwnerRepository) SaveSensorOwner(ctx context.Context, sensorOwner domain.SensorOwner) error {
	_, err := r.pool.Exec(ctx, saveSensorOwnerQuery, sensorOwner.UserID, sensorOwner.SensorID)
	return err
}

func (r *SensorOwnerRepository) GetSensorsByUserID(ctx context.Context, userID int64) ([]domain.SensorOwner, error) {
	rows, err := r.pool.Query(ctx, getSensorsByUserIDQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sensors []domain.SensorOwner
	for rows.Next() {
		var so domain.SensorOwner
		err := rows.Scan(&so.UserID, &so.SensorID)
		if err != nil {
			return nil, err
		}
		sensors = append(sensors, so)
	}
	return sensors, nil
}
