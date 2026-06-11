package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	db "github.com/yourusername/user-api/db/sqlc"
	"github.com/yourusername/user-api/internal/models"
	"github.com/yourusername/user-api/internal/repository"
)

// ErrUserNotFound is returned when a user cannot be found by ID.
var ErrUserNotFound = errors.New("user not found")

// UserService defines the business logic interface for user operations.
type UserService interface {
	CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.UserResponse, error)
	GetUserByID(ctx context.Context, id int32) (*models.UserResponse, error)
	UpdateUser(ctx context.Context, id int32, req models.UpdateUserRequest) (*models.UserResponse, error)
	DeleteUser(ctx context.Context, id int32) error
	ListUsers(ctx context.Context, page, pageSize int) (*models.PaginatedUsersResponse, error)
}

type userService struct {
	repo repository.UserRepository
	log  *zap.Logger
}

// NewUserService creates a new UserService.
func NewUserService(repo repository.UserRepository, log *zap.Logger) UserService {
	return &userService{repo: repo, log: log}
}

// CreateUser creates a new user after parsing and validating the DOB.
func (s *userService) CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.UserResponse, error) {
	dob, err := time.Parse("2006-01-02", req.Dob)
	if err != nil {
		return nil, fmt.Errorf("invalid dob format, expected YYYY-MM-DD: %w", err)
	}

	user, err := s.repo.Create(ctx, db.CreateUserParams{
		Name: req.Name,
		Dob:  dob,
	})
	if err != nil {
		s.log.Error("Failed to create user", zap.String("name", req.Name), zap.Error(err))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.log.Info("User created", zap.Int32("id", user.ID), zap.String("name", user.Name))
	return toResponse(user, true), nil
}

// GetUserByID retrieves a user by ID and calculates their current age.
func (s *userService) GetUserByID(ctx context.Context, id int32) (*models.UserResponse, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		s.log.Error("Failed to get user", zap.Int32("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	s.log.Info("User fetched", zap.Int32("id", user.ID))
	return toResponse(user, true), nil
}

// UpdateUser updates an existing user's name and DOB.
func (s *userService) UpdateUser(ctx context.Context, id int32, req models.UpdateUserRequest) (*models.UserResponse, error) {
	// Ensure user exists first
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	dob, err := time.Parse("2006-01-02", req.Dob)
	if err != nil {
		return nil, fmt.Errorf("invalid dob format, expected YYYY-MM-DD: %w", err)
	}

	updated, err := s.repo.Update(ctx, db.UpdateUserParams{
		ID:   id,
		Name: req.Name,
		Dob:  dob,
	})
	if err != nil {
		s.log.Error("Failed to update user", zap.Int32("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	s.log.Info("User updated", zap.Int32("id", updated.ID), zap.String("name", updated.Name))
	// Update response does NOT include age (per the spec response example)
	return toResponse(updated, false), nil
}

// DeleteUser removes a user by ID after confirming they exist.
func (s *userService) DeleteUser(ctx context.Context, id int32) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to fetch user: %w", err)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.log.Error("Failed to delete user", zap.Int32("id", id), zap.Error(err))
		return fmt.Errorf("failed to delete user: %w", err)
	}

	s.log.Info("User deleted", zap.Int32("id", id))
	return nil
}

// ListUsers returns a paginated list of users with their calculated ages.
func (s *userService) ListUsers(ctx context.Context, page, pageSize int) (*models.PaginatedUsersResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	users, err := s.repo.List(ctx, db.ListUsersParams{
		Limit:  int32(pageSize),
		Offset: int32(offset),
	})
	if err != nil {
		s.log.Error("Failed to list users", zap.Error(err))
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	total, err := s.repo.Count(ctx)
	if err != nil {
		s.log.Error("Failed to count users", zap.Error(err))
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	responses := make([]models.UserResponse, 0, len(users))
	for _, u := range users {
		responses = append(responses, *toResponse(u, true))
	}

	s.log.Info("Users listed", zap.Int("page", page), zap.Int("count", len(responses)))

	return &models.PaginatedUsersResponse{
		Data:       responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// toResponse converts a db.User to a models.UserResponse.
// If withAge is true, the age field is calculated and included.
func toResponse(u db.User, withAge bool) *models.UserResponse {
	resp := &models.UserResponse{
		ID:   u.ID,
		Name: u.Name,
		Dob:  u.Dob.Format("2006-01-02"),
	}
	if withAge {
		resp.Age = models.CalculateAge(u.Dob)
	}
	return resp
}
