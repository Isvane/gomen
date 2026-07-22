package repository

import (
	"errors"
	"sync"

	models "github.com/Isvane/gohan/internal/model"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
)

type UserRepo struct {
	mu       sync.RWMutex
	userInfo map[string]int
}

func NewUserRepo() *UserRepo {
	return &UserRepo{
		userInfo: make(map[string]int),
	}
}

func (r *UserRepo) Create(user models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.userInfo[user.Name]; exists {
		return ErrUserAlreadyExists
	}
	r.userInfo[user.Name] = user.Age
	return nil
}

func (r *UserRepo) Get(name string) (models.User, error) {
	r.mu.RLock()
	age, ok := r.userInfo[name]
	r.mu.RUnlock()

	if !ok {
		return models.User{}, ErrUserNotFound
	}
	return models.User{Name: name, Age: age}, nil
}

func (r *UserRepo) Update(name string, updated models.User) (models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.userInfo[name]
	if !ok {
		return models.User{}, ErrUserNotFound
	}

	if updated.Name != "" && updated.Name != name {
		delete(r.userInfo, name)
		r.userInfo[updated.Name] = updated.Age
	} else {
		updated.Name = name
		r.userInfo[name] = updated.Age
	}

	return updated, nil
}

func (r *UserRepo) Delete(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.userInfo[name]
	if !ok {
		return ErrUserNotFound
	}

	delete(r.userInfo, name)
	return nil
}
