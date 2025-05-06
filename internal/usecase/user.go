package usecase

import (
	"context"
	"errors"
	"homework/internal/domain"
)

type User struct {
	ur  UserRepository
	sor SensorOwnerRepository
	sr  SensorRepository
}

func NewUser(ur UserRepository, sor SensorOwnerRepository, sr SensorRepository) *User {
	return &User{
		ur:  ur,
		sor: sor,
		sr:  sr,
	}
}

func (u *User) RegisterUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	if user == nil {
		return nil, errors.New("nil user")
	}
	if user.Name == "" {
		return nil, ErrInvalidUserName
	}
	if err := u.ur.SaveUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (u *User) AttachSensorToUser(ctx context.Context, userID, sensorID int64) error {
	if _, err := u.ur.GetUserByID(ctx, userID); err != nil {
		return err
	}
	if _, err := u.sr.GetSensorByID(ctx, sensorID); err != nil {
		return err
	}
	return u.sor.SaveSensorOwner(ctx, domain.SensorOwner{
		UserID:   userID,
		SensorID: sensorID,
	})
}

func (u *User) GetUserSensors(ctx context.Context, userID int64) ([]domain.Sensor, error) {
	if _, err := u.ur.GetUserByID(ctx, userID); err != nil {
		return nil, err
	}
	sensorOwners, err := u.sor.GetSensorsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	sensors := make([]domain.Sensor, len(sensorOwners))
	for i, sensorOwner := range sensorOwners {
		sensor, err := u.sr.GetSensorByID(ctx, sensorOwner.SensorID)
		if err != nil {
			return nil, err
		}
		sensors[i] = *sensor
	}
	return sensors, err
}
