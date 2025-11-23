package service

import (
	"context"
	"fmt"

	"github.com/felipedavid/chatting/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserService struct {
	queries *storage.Queries
}

func NewUserService(queries *storage.Queries) *UserService {
	return &UserService{queries: queries}
}

type CreateUserRequest struct {
	PhoneNumber string
	DisplayName string
	About       string
}

type UserResponse struct {
	ID          pgtype.UUID
	PhoneNumber string
	DisplayName string
	About       string
	CreatedAt   pgtype.Timestamptz
}

func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*UserResponse, error) {
	if req.PhoneNumber == "" {
		return nil, fmt.Errorf("phone number is required")
	}

	// Check if user already exists
	existingUser, err := s.queries.GetUserByPhoneNumber(ctx, req.PhoneNumber)
	if err == nil && existingUser.ID.Valid {
		return nil, fmt.Errorf("user with phone number %s already exists", req.PhoneNumber)
	}

	params := storage.CreateUserParams{
		PhoneNumber: req.PhoneNumber,
		DisplayName: pgtype.Text{String: req.DisplayName, Valid: req.DisplayName != ""},
		About:       pgtype.Text{String: req.About, Valid: req.About != ""},
	}

	user, err := s.queries.CreateUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &UserResponse{
		ID:          user.ID,
		PhoneNumber: user.PhoneNumber,
		DisplayName: user.DisplayName.String,
		About:       user.About.String,
		CreatedAt:   user.CreatedAt,
	}, nil
}

func (s *UserService) GetUser(ctx context.Context, userID pgtype.UUID) (*UserResponse, error) {
	user, err := s.queries.GetUser(ctx, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &UserResponse{
		ID:          user.ID,
		PhoneNumber: user.PhoneNumber,
		DisplayName: user.DisplayName.String,
		About:       user.About.String,
		CreatedAt:   user.CreatedAt,
	}, nil
}

func (s *UserService) GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (*UserResponse, error) {
	if phoneNumber == "" {
		return nil, fmt.Errorf("phone number is required")
	}

	user, err := s.queries.GetUserByPhoneNumber(ctx, phoneNumber)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by phone number: %w", err)
	}

	return &UserResponse{
		ID:          user.ID,
		PhoneNumber: user.PhoneNumber,
		DisplayName: user.DisplayName.String,
		About:       user.About.String,
		CreatedAt:   user.CreatedAt,
	}, nil
}

func (s *UserService) ListUsers(ctx context.Context) ([]UserResponse, error) {
	users, err := s.queries.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	var responses []UserResponse
	for _, user := range users {
		responses = append(responses, UserResponse{
			ID:          user.ID,
			PhoneNumber: user.PhoneNumber,
			DisplayName: user.DisplayName.String,
			About:       user.About.String,
			CreatedAt:   user.CreatedAt,
		})
	}

	return responses, nil
}

func (s *UserService) DeleteUser(ctx context.Context, userID pgtype.UUID) error {
	err := s.queries.DeleteUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}
