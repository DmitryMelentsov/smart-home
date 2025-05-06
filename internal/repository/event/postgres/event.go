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

var ErrEventNotFound = errors.New("event not found")

type EventRepository struct {
	pool *pgxpool.Pool
}

func NewEventRepository(pool *pgxpool.Pool) *EventRepository {
	return &EventRepository{
		pool,
	}
}

const (
	saveEventQuery = `
		INSERT INTO events (timestamp, sensor_serial_number, sensor_id, payload)
		VALUES ($1, $2, $3, $4)
	`

	getLastEventQuery = `
		SELECT timestamp, sensor_serial_number, sensor_id, payload
		FROM events
		WHERE sensor_id = $1
		ORDER BY timestamp DESC
		LIMIT 1
	`

	getSensorHistoryQuery = `
		SELECT timestamp, sensor_serial_number, sensor_id, payload
		FROM events
		WHERE sensor_id = $1 AND timestamp BETWEEN $2 AND $3 
	`

	checkSensorExistsQuery = `
		SELECT EXISTS (
			SELECT *
			FROM sensors
			WHERE id = $1
		)
	`
)

func (r *EventRepository) SaveEvent(ctx context.Context, event *domain.Event) error {
	_, err := r.pool.Exec(
		ctx,
		saveEventQuery,
		event.Timestamp,
		event.SensorSerialNumber,
		event.SensorID,
		event.Payload,
	)
	return err
}

func (r *EventRepository) GetLastEventBySensorID(ctx context.Context, id int64) (*domain.Event, error) {
	row := r.pool.QueryRow(ctx, getLastEventQuery, id)
	event := &domain.Event{}
	if err := row.Scan(&event.Timestamp, &event.SensorSerialNumber, &event.SensorID, &event.Payload); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrEventNotFound
		}
		return nil, err
	}
	return event, nil
}

func (r *EventRepository) GetSensorHistory(ctx context.Context, id int64, start, end time.Time) ([]domain.Event, error) {
	var check bool
	err := r.pool.QueryRow(ctx, checkSensorExistsQuery, id).Scan(&check)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, usecase.ErrSensorNotFound
	}
	rows, err := r.pool.Query(ctx, getSensorHistoryQuery, id, start, end)
	if err != nil {
		return nil, err
	}
	var events []domain.Event
	defer rows.Close()
	for rows.Next() {
		var event domain.Event
		err := rows.Scan(&event.Timestamp, &event.SensorSerialNumber, &event.SensorID, &event.Payload)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}
