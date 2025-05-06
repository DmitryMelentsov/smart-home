package inmemory

import (
	"context"
	"errors"
	"homework/internal/domain"
	"homework/internal/usecase"
	"sync"
	"time"
)

type SensorRepository struct {
	sensorsById map[int64]*domain.Sensor
	sensorsBySN map[string]*domain.Sensor
	mu          sync.Mutex
}

func NewSensorRepository() *SensorRepository {
	return &SensorRepository{
		sensorsById: make(map[int64]*domain.Sensor),
		sensorsBySN: make(map[string]*domain.Sensor),
	}
}

func (r *SensorRepository) SaveSensor(ctx context.Context, sensor *domain.Sensor) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if sensor == nil {
		return errors.New("nil sensor")
	}

	sensor.ID = int64(len(r.sensorsById) + 1)
	sensor.RegisteredAt = time.Now()

	r.sensorsById[sensor.ID] = sensor
	r.sensorsBySN[sensor.SerialNumber] = sensor
	return nil
}

func (r *SensorRepository) GetSensors(ctx context.Context) ([]domain.Sensor, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	sensors := make([]domain.Sensor, 0, len(r.sensorsById))
	for _, s := range r.sensorsById {
		sensors = append(sensors, *s)
	}
	return sensors, nil
}

func (r *SensorRepository) GetSensorByID(ctx context.Context, id int64) (*domain.Sensor, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if _, ok := r.sensorsById[id]; !ok {
		return nil, usecase.ErrSensorNotFound
	}
	return r.sensorsById[id], nil
}

func (r *SensorRepository) GetSensorBySerialNumber(ctx context.Context, sn string) (*domain.Sensor, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if _, ok := r.sensorsBySN[sn]; !ok {
		return nil, usecase.ErrSensorNotFound
	}
	return r.sensorsBySN[sn], nil
}
