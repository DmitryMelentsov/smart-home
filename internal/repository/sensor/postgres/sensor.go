package postgres

import (
	"context"
	"errors"
	"homework/internal/domain"
	"homework/internal/usecase"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	saveSensorQuery = `
		INSERT INTO sensors (serial_number, type, current_state, description, is_active, registered_at, last_activity)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	saveSensorQueryWithID = `
		UPDATE sensors 
		Set serial_number = $1, 
		    type = $2, 
		    current_state = $3, 
		    description = $4, 
		    is_active = $5, 
		    registered_at = $6, 
		    last_activity = $7
		WHERE id = $8
	`

	getSensorsQuery = `
		SELECT id, serial_number, type, current_state, description, is_active, registered_at, last_activity
		FROM sensors
	`

	getSensorByIDQuery = `
		SELECT id, serial_number, type, current_state, description, is_active, registered_at, last_activity
		FROM sensors
		WHERE id = $1
	`

	getSensorBySerialQuery = `
		SELECT id, serial_number, type, current_state, description, is_active, registered_at, last_activity
		FROM sensors
		WHERE serial_number = $1`
)

type SensorRepository struct {
	pool *pgxpool.Pool
}

func NewSensorRepository(pool *pgxpool.Pool) *SensorRepository {
	return &SensorRepository{
		pool: pool,
	}
}

func (r *SensorRepository) SaveSensor(ctx context.Context, sensor *domain.Sensor) error {
	if sensor.ID == 0 {
		sensor.RegisteredAt = time.Now()
		return r.pool.QueryRow(ctx, saveSensorQuery, sensor.SerialNumber, sensor.Type, sensor.CurrentState,
			sensor.Description, sensor.IsActive, sensor.RegisteredAt, sensor.LastActivity).Scan(&sensor.ID)
	}
	_, err := r.pool.Exec(ctx, saveSensorQueryWithID, sensor.SerialNumber, sensor.Type, sensor.CurrentState,
		sensor.Description, sensor.IsActive, sensor.RegisteredAt, sensor.LastActivity, sensor.ID)
	return err
}

func (r *SensorRepository) GetSensors(ctx context.Context) ([]domain.Sensor, error) {
	rows, err := r.pool.Query(ctx, getSensorsQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sensors []domain.Sensor
	for rows.Next() {
		var s domain.Sensor
		err := rows.Scan(
			&s.ID,
			&s.SerialNumber,
			&s.Type,
			&s.CurrentState,
			&s.Description,
			&s.IsActive,
			&s.RegisteredAt,
			&s.LastActivity,
		)
		if err != nil {
			return nil, err
		}
		sensors = append(sensors, s)
	}
	return sensors, nil
}

func (r *SensorRepository) GetSensorByID(ctx context.Context, id int64) (*domain.Sensor, error) {
	var s domain.Sensor
	err := r.pool.QueryRow(ctx, getSensorByIDQuery, id).Scan(
		&s.ID,
		&s.SerialNumber,
		&s.Type,
		&s.CurrentState,
		&s.Description,
		&s.IsActive,
		&s.RegisteredAt,
		&s.LastActivity,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, usecase.ErrSensorNotFound
	}
	return &s, err
}

func (r *SensorRepository) GetSensorBySerialNumber(ctx context.Context, sn string) (*domain.Sensor, error) {
	var s domain.Sensor
	err := r.pool.QueryRow(ctx, getSensorBySerialQuery, sn).Scan(
		&s.ID,
		&s.SerialNumber,
		&s.Type,
		&s.CurrentState,
		&s.Description,
		&s.IsActive,
		&s.RegisteredAt,
		&s.LastActivity,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, usecase.ErrSensorNotFound
	}
	return &s, err
}
