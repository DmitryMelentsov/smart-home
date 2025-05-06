package inmemory

import (
	"context"
	"homework/internal/domain"
	"homework/internal/usecase"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestEventRepository_SaveEvent(t *testing.T) {
	t.Run("err, event is nil", func(t *testing.T) {
		er := NewEventRepository()
		err := er.SaveEvent(context.Background(), nil)
		assert.Error(t, err)
	})

	t.Run("fail, ctx cancelled", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := er.SaveEvent(ctx, &domain.Event{})
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("fail, ctx deadline exceeded", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()

		err := er.SaveEvent(ctx, &domain.Event{})
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("ok, save and get one", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		event := &domain.Event{
			Timestamp:          time.Now(),
			SensorSerialNumber: "0123456789",
			Payload:            0,
		}

		err := er.SaveEvent(ctx, event)
		assert.NoError(t, err)

		actualEvent, err := er.GetLastEventBySensorID(ctx, event.SensorID)
		assert.NoError(t, err)
		assert.NotNil(t, actualEvent)
		assert.Equal(t, event.Timestamp, actualEvent.Timestamp)
		assert.Equal(t, event.SensorSerialNumber, actualEvent.SensorSerialNumber)
		assert.Equal(t, event.Payload, actualEvent.Payload)
	})

	t.Run("ok, collision test", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		wg := sync.WaitGroup{}
		var lastEvent domain.Event
		for i := 0; i < 1000; i++ {
			event := &domain.Event{
				Timestamp:          time.Now(),
				SensorSerialNumber: "0123456789",
				Payload:            0,
			}
			lastEvent = *event
			wg.Add(1)
			go func() {
				defer wg.Done()
				assert.NoError(t, er.SaveEvent(ctx, event))
			}()
		}

		wg.Wait()

		actualEvent, err := er.GetLastEventBySensorID(ctx, lastEvent.SensorID)
		assert.NoError(t, err)
		assert.NotNil(t, actualEvent)
		assert.Equal(t, lastEvent.Timestamp, actualEvent.Timestamp)
		assert.Equal(t, lastEvent.SensorSerialNumber, actualEvent.SensorSerialNumber)
		assert.Equal(t, lastEvent.Payload, actualEvent.Payload)
	})
}

func TestEventRepository_GetLastEventBySensorID(t *testing.T) {
	t.Run("fail, ctx cancelled", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := er.GetLastEventBySensorID(ctx, 0)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("fail, ctx deadline exceeded", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()

		_, err := er.GetLastEventBySensorID(ctx, 0)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("fail, event not found", func(t *testing.T) {
		er := NewEventRepository()
		_, err := er.GetLastEventBySensorID(context.Background(), 234)
		assert.ErrorIs(t, err, usecase.ErrEventNotFound)
	})

	t.Run("ok, save and get one", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sensorID := int64(12345)
		var lastEvent *domain.Event
		for i := 0; i < 10; i++ {
			lastEvent = &domain.Event{
				Timestamp: time.Now(),
				SensorID:  sensorID,
				Payload:   0,
			}
			time.Sleep(10 * time.Millisecond)
			assert.NoError(t, er.SaveEvent(ctx, lastEvent))
		}

		for i := 0; i < 10; i++ {
			event := &domain.Event{
				Timestamp: time.Now(),
				SensorID:  54321,
				Payload:   0,
			}
			assert.NoError(t, er.SaveEvent(ctx, event))
		}

		actualEvent, err := er.GetLastEventBySensorID(ctx, lastEvent.SensorID)
		assert.NoError(t, err)
		assert.NotNil(t, actualEvent)
		assert.Equal(t, lastEvent.Timestamp, actualEvent.Timestamp)
		assert.Equal(t, lastEvent.SensorSerialNumber, actualEvent.SensorSerialNumber)
		assert.Equal(t, lastEvent.Payload, actualEvent.Payload)
	})
}

func TestEventRepository_GetSensorHistory(t *testing.T) {
	now := time.Now()
	sensorID := int64(12345)
	start := now.Add(-time.Hour)
	end := now.Add(time.Hour)

	type testCase struct {
		name        string
		ctxFunc     func() (context.Context, context.CancelFunc)
		sensorID    int64
		prepare     func(r *EventRepository)
		expectedErr error
		expectedLen int
	}

	testCases := []testCase{
		{
			name: "fail, ctx cancelled",
			ctxFunc: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx, cancel
			},
			sensorID:    0,
			expectedErr: context.Canceled,
		},
		{
			name: "fail, ctx deadline exceeded",
			ctxFunc: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Nanosecond)
			},
			sensorID:    0,
			expectedErr: context.DeadlineExceeded,
		},
		{
			name: "err, sensor not found",
			ctxFunc: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			sensorID:    234,
			expectedErr: usecase.ErrSensorNotFound,
		},
		{
			name: "ok, no error",
			ctxFunc: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			sensorID: sensorID,
			prepare: func(r *EventRepository) {
				for i := 0; i < 10; i++ {
					_ = r.SaveEvent(context.Background(), &domain.Event{
						Timestamp: now.Add(time.Duration(i) * time.Minute),
						SensorID:  sensorID,
						Payload:   0,
					})
				}
			},
			expectedLen: 10,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			er := NewEventRepository()
			ctx, cancel := tc.ctxFunc()
			defer cancel()

			if tc.prepare != nil {
				tc.prepare(er)
			}

			history, err := er.GetSensorHistory(ctx, tc.sensorID, start, end)

			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
				return
			}

			require.NoError(t, err)
			assert.Len(t, history, tc.expectedLen)

			for _, event := range history {
				assert.True(t, event.Timestamp.After(start))
				assert.True(t, event.Timestamp.Before(end))
			}
		})
	}
}
