package inmemory

import (
	"context"
	"homework/internal/domain"
	"sync"
)

type SensorOwnerRepository struct {
	sensorOwners map[int64][]domain.SensorOwner
	mu           sync.Mutex
}

func NewSensorOwnerRepository() *SensorOwnerRepository {
	return &SensorOwnerRepository{
		sensorOwners: make(map[int64][]domain.SensorOwner),
	}
}

func (r *SensorOwnerRepository) SaveSensorOwner(ctx context.Context, sensorOwner domain.SensorOwner) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, exists := r.sensorOwners[sensorOwner.UserID]; !exists {
		r.sensorOwners[sensorOwner.UserID] = []domain.SensorOwner{}
	}
	r.sensorOwners[sensorOwner.UserID] = append(r.sensorOwners[sensorOwner.UserID], sensorOwner)
	return nil
}

func (r *SensorOwnerRepository) GetSensorsByUserID(ctx context.Context, userID int64) ([]domain.SensorOwner, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	sensorOwners, ok := r.sensorOwners[userID]
	if !ok {
		return nil, nil
	}
	return sensorOwners, nil
}
