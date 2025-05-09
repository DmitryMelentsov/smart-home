package inmemory

import (
	"context"
	"homework/internal/domain"
	"math/rand/v2"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSensorOwnerRepository_SaveSensorOwner(t *testing.T) {
	t.Run("fail, ctx cancelled", func(t *testing.T) {
		sor := NewSensorOwnerRepository()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := sor.SaveSensorOwner(ctx, domain.SensorOwner{})
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("fail, ctx deadline exceeded", func(t *testing.T) {
		sor := NewSensorOwnerRepository()
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()

		err := sor.SaveSensorOwner(ctx, domain.SensorOwner{})
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("ok, save and get one", func(t *testing.T) {
		sor := NewSensorOwnerRepository()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sensorOwner := domain.SensorOwner{
			UserID:   1234,
			SensorID: 5678,
		}

		err := sor.SaveSensorOwner(ctx, sensorOwner)
		assert.NoError(t, err)

		list, err := sor.GetSensorsByUserID(ctx, 1234)
		assert.NoError(t, err)

		assert.Len(t, list, 1)
		assert.Equal(t, list[0].SensorID, int64(5678))
	})

	t.Run("ok, collision test", func(t *testing.T) {
		sr := NewSensorOwnerRepository()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		wg := sync.WaitGroup{}
		for i := int64(0); i < 1000; i++ {
			sensorOwner := domain.SensorOwner{
				UserID:   1234 + i,
				SensorID: 5678 + i,
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				assert.NoError(t, sr.SaveSensorOwner(ctx, sensorOwner))
			}()
		}

		wg.Wait()
	})
}

func TestSensorOwnerRepository_GetSensorsByUserID(t *testing.T) {
	t.Run("fail, ctx cancelled", func(t *testing.T) {
		sor := NewSensorOwnerRepository()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := sor.GetSensorsByUserID(ctx, 1)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("fail, ctx deadline exceeded", func(t *testing.T) {
		sor := NewSensorOwnerRepository()
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()

		_, err := sor.GetSensorsByUserID(ctx, 1)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("ok, get empty list", func(t *testing.T) {
		sr := NewSensorOwnerRepository()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sensors, err := sr.GetSensorsByUserID(ctx, 1)
		assert.NoError(t, err)
		assert.Len(t, sensors, 0)
	})

	t.Run("ok, get list", func(t *testing.T) {
		sr := NewSensorOwnerRepository()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		assert.NoError(t, sr.SaveSensorOwner(ctx, domain.SensorOwner{UserID: 1, SensorID: 1}))
		assert.NoError(t, sr.SaveSensorOwner(ctx, domain.SensorOwner{UserID: 1, SensorID: 2}))
		assert.NoError(t, sr.SaveSensorOwner(ctx, domain.SensorOwner{UserID: 1, SensorID: 3}))
		assert.NoError(t, sr.SaveSensorOwner(ctx, domain.SensorOwner{UserID: 2, SensorID: 4}))
		assert.NoError(t, sr.SaveSensorOwner(ctx, domain.SensorOwner{UserID: 3, SensorID: 5}))

		sensors, err := sr.GetSensorsByUserID(ctx, 1)
		assert.NoError(t, err)
		assert.Len(t, sensors, 3)

		sensors, err = sr.GetSensorsByUserID(ctx, 3)
		assert.NoError(t, err)
		assert.Len(t, sensors, 1)
	})
}

func FuzzSensorOwnerRepository_GetSensorsByUserID(f *testing.F) {
	f.Add(491)

	f.Fuzz(func(t *testing.T, sz int) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sr := NewSensorOwnerRepository()
		usersId := make([]int64, 0)
		results := make(map[int64][]domain.SensorOwner)
		for i := 0; i < sz; i++ {
			sensorId := int64(rand.Int() % sz)
			userId := int64(rand.Int() % sz)
			sensorOwner := domain.SensorOwner{
				UserID:   userId,
				SensorID: sensorId,
			}
			err := sr.SaveSensorOwner(context.Background(), sensorOwner)
			if err != nil {
				continue
			}
			usersId = append(usersId, userId)
			results[userId] = append(results[userId], sensorOwner)
		}
		for _, userId := range usersId {
			sensors, err := sr.GetSensorsByUserID(ctx, userId)
			assert.NoError(t, err)
			assert.Len(t, sensors, len(results[userId]))
			for _, sensor := range sensors {
				assert.Contains(t, results[userId], sensor)
			}
		}
	})
}
