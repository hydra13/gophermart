package user

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/hydra13/gophermart/internal/models"
	repositories "github.com/hydra13/gophermart/internal/repositories"
)

type UserRepository interface {
	Create(ctx context.Context, login, passwordHash string) (int64, error)
	GetByLogin(ctx context.Context, login string) (*models.User, error)
}

type TransactionRepository interface {
	CreateUser(ctx context.Context, user models.User) (int64, error)
}

type UserService struct {
	userRepo UserRepository
	txRepo   TransactionRepository
}

func NewUserService(
	userRepo UserRepository,
	txRepo TransactionRepository,
) *UserService {
	return &UserService{
		userRepo: userRepo,
		txRepo:   txRepo,
	}
}

func (s *UserService) Register(ctx context.Context, login, password string) (int64, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}

	userID, err := s.txRepo.CreateUser(ctx, models.User{
		Login:        login,
		PasswordHash: string(passwordHash),
	})
	if err != nil {
		if err == repositories.ErrLoginExists {
			return 0, models.ErrUserAlreadyExists
		}
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	return userID, nil
}

func (s *UserService) Login(ctx context.Context, login, password string) (int64, error) {
	user, err := s.userRepo.GetByLogin(ctx, login)
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
