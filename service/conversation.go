package service

import (
	"context"
	"fmt"

	"github.com/felipedavid/chatting/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type ConversationService struct {
	queries *storage.Queries
}

func NewConversationService(queries *storage.Queries) *ConversationService {
	return &ConversationService{queries: queries}
}

type CreateConversationRequest struct {
	IsGroup   bool
	Title     string
	CreatedBy pgtype.UUID
}

type ConversationResponse struct {
	ID        pgtype.UUID
	IsGroup   bool
	Title     string
	CreatedBy pgtype.UUID
	CreatedAt pgtype.Timestamptz
}

type ConversationWithCreatorResponse struct {
	ID           pgtype.UUID
	IsGroup      bool
	Title        string
	CreatedBy    pgtype.UUID
	CreatedAt    pgtype.Timestamptz
	CreatorPhone string
	CreatorName  string
}

func (s *ConversationService) CreateConversation(ctx context.Context, req CreateConversationRequest) (*ConversationResponse, error) {
	if !req.CreatedBy.Valid {
		return nil, fmt.Errorf("created by user ID is required")
	}

	params := storage.CreateConversationParams{
		IsGroup:   req.IsGroup,
		Title:     pgtype.Text{String: req.Title, Valid: req.Title != ""},
		CreatedBy: req.CreatedBy,
	}

	conversation, err := s.queries.CreateConversation(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	return &ConversationResponse{
		ID:        conversation.ID,
		IsGroup:   conversation.IsGroup,
		Title:     conversation.Title.String,
		CreatedBy: conversation.CreatedBy,
		CreatedAt: conversation.CreatedAt,
	}, nil
}

func (s *ConversationService) GetConversation(ctx context.Context, conversationID pgtype.UUID) (*ConversationResponse, error) {
	if !conversationID.Valid {
		return nil, fmt.Errorf("conversation ID is required")
	}

	conversation, err := s.queries.GetConversation(ctx, conversationID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("conversation not found")
		}
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	return &ConversationResponse{
		ID:        conversation.ID,
		IsGroup:   conversation.IsGroup,
		Title:     conversation.Title.String,
		CreatedBy: conversation.CreatedBy,
		CreatedAt: conversation.CreatedAt,
	}, nil
}

func (s *ConversationService) GetConversationWithCreator(ctx context.Context, conversationID pgtype.UUID) (*ConversationWithCreatorResponse, error) {
	if !conversationID.Valid {
		return nil, fmt.Errorf("conversation ID is required")
	}

	conversation, err := s.queries.GetConversationByIdWithCreator(ctx, conversationID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("conversation not found")
		}
		return nil, fmt.Errorf("failed to get conversation with creator: %w", err)
	}

	return &ConversationWithCreatorResponse{
		ID:           conversation.ID,
		IsGroup:      conversation.IsGroup,
		Title:        conversation.Title.String,
		CreatedBy:    conversation.CreatedBy,
		CreatedAt:    conversation.CreatedAt,
		CreatorPhone: conversation.CreatorPhone.String,
		CreatorName:  conversation.CreatorName.String,
	}, nil
}

func (s *ConversationService) DeleteConversation(ctx context.Context, conversationID pgtype.UUID) error {
	if !conversationID.Valid {
		return fmt.Errorf("conversation ID is required")
	}

	// First delete all messages in the conversation
	if err := s.queries.DeleteConversationMessages(ctx, conversationID); err != nil {
		return fmt.Errorf("failed to delete conversation messages: %w", err)
	}

	// Then delete the conversation
	if err := s.queries.DeleteConversation(ctx, conversationID); err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}

	return nil
}

func (s *ConversationService) AddParticipant(ctx context.Context, conversationID, userID pgtype.UUID, role string) error {
	if !conversationID.Valid || !userID.Valid {
		return fmt.Errorf("conversation ID and user ID are required")
	}

	params := storage.AddConversationParticipantParams{
		ConversationID: conversationID,
		UserID:         userID,
		Role:           pgtype.Text{String: role, Valid: role != ""},
	}

	if _, err := s.queries.AddConversationParticipant(ctx, params); err != nil {
		return fmt.Errorf("failed to add participant: %w", err)
	}

	return nil
}

func (s *ConversationService) RemoveParticipant(ctx context.Context, conversationID, userID pgtype.UUID) error {
	if !conversationID.Valid || !userID.Valid {
		return fmt.Errorf("conversation ID and user ID are required")
	}

	if err := s.queries.RemoveConversationParticipant(ctx, storage.RemoveConversationParticipantParams{
		ConversationID: conversationID,
		UserID:         userID,
	}); err != nil {
		return fmt.Errorf("failed to remove participant: %w", err)
	}

	return nil
}

func (s *ConversationService) GetConversationParticipants(ctx context.Context, conversationID pgtype.UUID) ([]ParticipantResponse, error) {
	if !conversationID.Valid {
		return nil, fmt.Errorf("conversation ID is required")
	}

	participants, err := s.queries.ListConversationParticipantsWithDetails(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation participants: %w", err)
	}

	var responses []ParticipantResponse
	for _, participant := range participants {
		responses = append(responses, ParticipantResponse{
			ConversationID: participant.ConversationID,
			UserID:         participant.UserID,
			Role:           participant.Role.String,
			JoinedAt:       participant.JoinedAt,
			PhoneNumber:    participant.PhoneNumber,
			DisplayName:    participant.DisplayName.String,
		})
	}

	return responses, nil
}

type ParticipantResponse struct {
	ConversationID pgtype.UUID
	UserID         pgtype.UUID
	Role           string
	JoinedAt       pgtype.Timestamptz
	PhoneNumber    string
	DisplayName    string
}
