package inmemory

import (
	"context"
	"errors"
	"homework/internal/domain"
	"homework/internal/usecase"
	"sync"
)

type UserRepository struct {
	usersByID map[int64]*domain.User
	nextID    int64
	mu        sync.Mutex
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		usersByID: make(map[int64]*domain.User),
		nextID:    1,
	}
}

func (r *UserRepository) SaveUser(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if user == nil {
		return errors.New("nil user")
	}
	user.ID = r.nextID
	r.nextID++
	r.usersByID[user.ID] = user
	return nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	user, ok := r.usersByID[id]
	if !ok {
		return nil, usecase.ErrUserNotFound
	}
	return user, nil
}
