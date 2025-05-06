package usecase

import (
	"context"
	"errors"
	"homework/internal/domain"
)

type Sensor struct {
	sr SensorRepository
}

func NewSensor(sr SensorRepository) *Sensor {
	return &Sensor{
		sr: sr,
	}
}

func (s *Sensor) RegisterSensor(ctx context.Context, sensor *domain.Sensor) (*domain.Sensor, error) {
	if sensor == nil {
		return nil, errors.New("nil sensor")
	}
	if sensor.Type != domain.SensorTypeContactClosure && sensor.Type != domain.SensorTypeADC {
		return nil, ErrWrongSensorType
	}
	if len(sensor.SerialNumber) != 10 {
		return nil, ErrWrongSensorSerialNumber
	}
	existingSensor, err := s.sr.GetSensorBySerialNumber(ctx, sensor.SerialNumber)
	if err != nil && !errors.Is(err, ErrSensorNotFound) {
		return nil, err
	}
	if existingSensor != nil {
		return existingSensor, nil
	}

	if err := s.sr.SaveSensor(ctx, sensor); err != nil {
		return nil, err
	}
	return sensor, nil
}

func (s *Sensor) GetSensors(ctx context.Context) ([]domain.Sensor, error) {
	sensors, err := s.sr.GetSensors(ctx)
	if err != nil {
		return nil, err
	}
	return sensors, nil
}

func (s *Sensor) GetSensorByID(ctx context.Context, id int64) (*domain.Sensor, error) {
	sensor, err := s.sr.GetSensorByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return sensor, nil
}

func (s *Sensor) GetSensorBySerialNumber(ctx context.Context, serialNumber string) (*domain.Sensor, error) {
	sensor, err := s.sr.GetSensorBySerialNumber(ctx, serialNumber)
	if err != nil {
		return nil, err
	}
	return sensor, nil
}
