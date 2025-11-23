package service

import (
	"context"
	"fmt"

	"github.com/felipedavid/chatting/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type MessageService struct {
	queries *storage.Queries
}

func NewMessageService(queries *storage.Queries) *MessageService {
	return &MessageService{queries: queries}
}

type CreateMessageRequest struct {
	ConversationID pgtype.UUID
	SenderID       pgtype.UUID
	Content        string
	MessageType    string
	ReplyToID      pgtype.UUID
}

type MessageResponse struct {
	ID             pgtype.UUID
	ConversationID pgtype.UUID
	SenderID       pgtype.UUID
	Content        string
	MessageType    string
	ReplyToID      pgtype.UUID
	CreatedAt      pgtype.Timestamptz
}

func (s *MessageService) CreateMessage(ctx context.Context, req CreateMessageRequest) (*MessageResponse, error) {
	if !req.ConversationID.Valid {
		return nil, fmt.Errorf("conversation ID is required")
	}
	if !req.SenderID.Valid {
		return nil, fmt.Errorf("sender ID is required")
	}
	if req.Content == "" {
		return nil, fmt.Errorf("message content is required")
	}
	if req.MessageType == "" {
		req.MessageType = "text"
	}

	// Check if sender is a participant in the conversation
	isParticipant, err := s.queries.IsUserInConversation(ctx, storage.IsUserInConversationParams{
		ConversationID: req.ConversationID,
		UserID:         req.SenderID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check if user is participant: %w", err)
	}
	if !isParticipant {
		return nil, fmt.Errorf("sender is not a participant in this conversation")
	}

	params := storage.CreateMessageParams{
		ConversationID: req.ConversationID,
		SenderID:       req.SenderID,
		Content:        pgtype.Text{String: req.Content, Valid: true},
		MessageType:    req.MessageType,
		ReplyToID:      req.ReplyToID,
	}

	message, err := s.queries.CreateMessage(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	return &MessageResponse{
		ID:             message.ID,
		ConversationID: message.ConversationID,
		SenderID:       message.SenderID,
		Content:        message.Content.String,
		MessageType:    message.MessageType,
		ReplyToID:      message.ReplyToID,
		CreatedAt:      message.CreatedAt,
	}, nil
}

func (s *MessageService) GetMessage(ctx context.Context, messageID pgtype.UUID) (*MessageResponse, error) {
	if !messageID.Valid {
		return nil, fmt.Errorf("message ID is required")
	}

	message, err := s.queries.GetMessage(ctx, messageID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("message not found")
		}
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return &MessageResponse{
		ID:             message.ID,
		ConversationID: message.ConversationID,
		SenderID:       message.SenderID,
		Content:        message.Content.String,
		MessageType:    message.MessageType,
		ReplyToID:      message.ReplyToID,
		CreatedAt:      message.CreatedAt,
	}, nil
}

func (s *MessageService) GetConversationMessages(ctx context.Context, conversationID pgtype.UUID, limit int32, offset int32) ([]MessageResponse, error) {
	if !conversationID.Valid {
		return nil, fmt.Errorf("conversation ID is required")
	}

	if limit <= 0 {
		limit = 50
	}

	messages, err := s.queries.ListConversationMessagesPaginated(ctx, storage.ListConversationMessagesPaginatedParams{
		ConversationID: conversationID,
		Limit:          limit,
		Offset:         offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation messages: %w", err)
	}

	var responses []MessageResponse
	for _, message := range messages {
		responses = append(responses, MessageResponse{
			ID:             message.ID,
			ConversationID: message.ConversationID,
			SenderID:       message.SenderID,
			Content:        message.Content.String,
			MessageType:    message.MessageType,
			ReplyToID:      message.ReplyToID,
			CreatedAt:      message.CreatedAt,
		})
	}

	return responses, nil
}

func (s *MessageService) GetLatestConversationMessage(ctx context.Context, conversationID pgtype.UUID) (*MessageResponse, error) {
	if !conversationID.Valid {
		return nil, fmt.Errorf("conversation ID is required")
	}

	message, err := s.queries.GetLatestConversationMessage(ctx, conversationID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("no messages found in conversation")
		}
		return nil, fmt.Errorf("failed to get latest message: %w", err)
	}

	return &MessageResponse{
		ID:             message.ID,
		ConversationID: message.ConversationID,
		SenderID:       message.SenderID,
		Content:        message.Content.String,
		MessageType:    message.MessageType,
		ReplyToID:      message.ReplyToID,
		CreatedAt:      message.CreatedAt,
	}, nil
}

func (s *MessageService) DeleteMessage(ctx context.Context, messageID pgtype.UUID) error {
	if !messageID.Valid {
		return fmt.Errorf("message ID is required")
	}

	if err := s.queries.DeleteMessage(ctx, messageID); err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

func (s *MessageService) CountConversationMessages(ctx context.Context, conversationID pgtype.UUID) (int64, error) {
	if !conversationID.Valid {
		return 0, fmt.Errorf("conversation ID is required")
	}

	count, err := s.queries.CountConversationMessages(ctx, conversationID)
	if err != nil {
		return 0, fmt.Errorf("failed to count conversation messages: %w", err)
	}

	return count, nil
}

func (s *MessageService) AddReaction(ctx context.Context, messageID, userID pgtype.UUID, reaction string) error {
	if !messageID.Valid || !userID.Valid {
		return fmt.Errorf("message ID and user ID are required")
	}
	if reaction == "" {
		return fmt.Errorf("reaction is required")
	}

	// Check if user has already reacted
	hasReacted, err := s.queries.HasUserReactedToMessage(ctx, storage.HasUserReactedToMessageParams{
		MessageID: messageID,
		UserID:    userID,
	})
	if err != nil {
		return fmt.Errorf("failed to check existing reaction: %w", err)
	}

	if hasReacted {
		// Update existing reaction
		_, err := s.queries.UpdateMessageReaction(ctx, storage.UpdateMessageReactionParams{
			MessageID: messageID,
			UserID:    userID,
			Reaction:  reaction,
		})
		if err != nil {
			return fmt.Errorf("failed to update reaction: %w", err)
		}

		return nil
	}

	// Add new reaction
	params := storage.AddMessageReactionParams{
		MessageID: messageID,
		UserID:    userID,
		Reaction:  reaction,
	}

	if _, err := s.queries.AddMessageReaction(ctx, params); err != nil {
		return fmt.Errorf("failed to add reaction: %w", err)
	}

	return nil
}

func (s *MessageService) RemoveReaction(ctx context.Context, messageID, userID pgtype.UUID) error {
	if !messageID.Valid || !userID.Valid {
		return fmt.Errorf("message ID and user ID are required")
	}

	if err := s.queries.RemoveMessageReaction(ctx, storage.RemoveMessageReactionParams{
		MessageID: messageID,
		UserID:    userID,
	}); err != nil {
		return fmt.Errorf("failed to remove reaction: %w", err)
	}

	return nil
}

func (s *MessageService) GetMessageReactions(ctx context.Context, messageID pgtype.UUID) ([]MessageReactionResponse, error) {
	if !messageID.Valid {
		return nil, fmt.Errorf("message ID is required")
	}

	reactions, err := s.queries.ListMessageReactionsWithDetails(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message reactions: %w", err)
	}

	var responses []MessageReactionResponse
	for _, reaction := range reactions {
		responses = append(responses, MessageReactionResponse{
			MessageID:   reaction.MessageID,
			UserID:      reaction.UserID,
			Reaction:    reaction.Reaction,
			ReactedAt:   reaction.ReactedAt,
			PhoneNumber: reaction.PhoneNumber,
			DisplayName: reaction.DisplayName.String,
		})
	}

	return responses, nil
}

type MessageReactionResponse struct {
	MessageID   pgtype.UUID
	UserID      pgtype.UUID
	Reaction    string
	ReactedAt   pgtype.Timestamptz
	PhoneNumber string
	DisplayName string
}
