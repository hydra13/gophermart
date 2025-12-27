package user

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/hydra13/gophermart/internal/models"
	repositories "github.com/hydra13/gophermart/internal/repositories/user"
)

type UserService struct {
	repo *repositories.UserRepository
}

func NewUserService(repo *repositories.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Register(ctx context.Context, login, password string) (int64, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}

	userID, err := s.repo.Create(ctx, login, string(passwordHash))
	if err != nil {
		if err == repositories.ErrLoginExists {
			return 0, models.ErrUserAlreadyExists
		}
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	return userID, nil
}

func (s *UserService) Login(ctx context.Context, login, password string) (int64, error) {
	user, err := s.repo.GetByLogin(ctx, login)
	if err != nil {
		if err == repositories.ErrUserNotFound {
			return 0, models.ErrUserNotFound
		}
		return 0, fmt.Errorf("failed to get user: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return 0, models.ErrUserNotFound
	}

	return user.ID, nil
}
