package service

import "github.com/felipedavid/chatting/storage"

type Container struct {
	UserService         *UserService
	ConversationService *ConversationService
	MessageService      *MessageService
}

func NewContainer(queries *storage.Queries) *Container {
	return &Container{
		UserService:         NewUserService(queries),
		ConversationService: NewConversationService(queries),
		MessageService:      NewMessageService(queries),
	}
}
