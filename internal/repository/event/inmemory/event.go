package inmemory

import (
	"context"
	"errors"
	"homework/internal/domain"
	"homework/internal/usecase"
	"sync"
	"time"
)

type EventRepository struct {
	eventsById map[int64][]*domain.Event
	mu         sync.Mutex
}

func NewEventRepository() *EventRepository {
	return &EventRepository{eventsById: make(map[int64][]*domain.Event)}
}

func (r *EventRepository) SaveEvent(ctx context.Context, event *domain.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if event == nil {
		return errors.New("event is nil")
	}
	r.eventsById[event.SensorID] = append(r.eventsById[event.SensorID], event)
	return nil
}

func (r *EventRepository) GetLastEventBySensorID(ctx context.Context, id int64) (*domain.Event, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	events, ok := r.eventsById[id]
	if !ok {
		return nil, usecase.ErrEventNotFound
	}
	lastEvent := events[len(events)-1]
	for i := len(events) - 1; i >= 0; i-- {
		if events[i].Timestamp.After(lastEvent.Timestamp) {
			lastEvent = events[i]
		}
	}
	return lastEvent, nil
}

func (r *EventRepository) GetSensorHistory(ctx context.Context, id int64, start, end time.Time) ([]domain.Event, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	events, ok := r.eventsById[id]
	if !ok {
		return nil, usecase.ErrSensorNotFound
	}
	var history []domain.Event
	for _, event := range events {
		if event.Timestamp.After(start) && event.Timestamp.Before(end) {
			history = append(history, *event)
		}
	}
	return history, nil
}
