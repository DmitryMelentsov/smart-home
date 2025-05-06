package usecase

import (
	"context"
	"homework/internal/domain"
	"time"
)

type Event struct {
	er EventRepository
	sr SensorRepository
}

func NewEvent(er EventRepository, sr SensorRepository) *Event {
	return &Event{
		er: er,
		sr: sr,
	}
}

func (e *Event) ReceiveEvent(ctx context.Context, event *domain.Event) error {
	if event.Timestamp.IsZero() {
		return ErrInvalidEventTimestamp
	}
	sensor, err := e.sr.GetSensorBySerialNumber(ctx, event.SensorSerialNumber)
	if err != nil {
		return err
	}
	sensor.CurrentState = event.Payload
	sensor.LastActivity = time.Now()
	event.SensorID = sensor.ID
	if err := e.er.SaveEvent(ctx, event); err != nil {
		return err
	}
	return e.sr.SaveSensor(ctx, sensor)
}

func (e *Event) GetLastEventBySensorID(ctx context.Context, id int64) (*domain.Event, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	event, err := e.er.GetLastEventBySensorID(ctx, id)
	if err != nil {
		return nil, err
	}
	return event, nil
}

func (e *Event) GetSensorHistory(ctx context.Context, id int64, start, end time.Time) ([]domain.Event, error) {
	return e.er.GetSensorHistory(ctx, id, start, end)
}
