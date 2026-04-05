package repository

import (
	"backend/internal/models"
	"errors"
	"sync"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByUsername(username string) (*models.User, error)
	FindByID(id uint) (*models.User, error)
	FindAll() ([]models.User, error)
	UpdateRole(id uint, role, status string) error
}

type memoryUserRepo struct {
	mu     sync.RWMutex
	users  map[uint]*models.User
	nextID uint
}

func NewUserRepository() UserRepository {
	return &memoryUserRepo{
		users:  make(map[uint]*models.User),
		nextID: 1,
	}
}

func (r *memoryUserRepo) Create(user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, u := range r.users {
		if u.Username == user.Username {
			return errors.New("username already exists")
		}
	}

	user.ID = r.nextID
	r.nextID++
	r.users[user.ID] = user
	return nil
}

func (r *memoryUserRepo) FindByUsername(username string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Username == username {
			// return a copy to prevent accidental mutation
			u := *user
			return &u, nil
		}
	}
	return nil, nil // Not found
}

func (r *memoryUserRepo) FindByID(id uint) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if user, exists := r.users[id]; exists {
		u := *user
		return &u, nil
	}
	return nil, errors.New("user not found")
}

func (r *memoryUserRepo) FindAll() ([]models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]models.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, *user)
	}
	return users, nil
}

func (r *memoryUserRepo) UpdateRole(id uint, role, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if user, exists := r.users[id]; exists {
		user.Role = role
		user.Status = status
		return nil
	}
	return errors.New("user not found")
}
