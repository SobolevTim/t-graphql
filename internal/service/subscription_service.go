package service

import (
	"github.com/SobolevTim/t-graphql/internal/store"
)

// CommentService отвечает за работу с комментариями
type SubscriptionService struct {
	store store.Store
}

// NewCommentService создаёт сервис для работы с комментариями
func NewSubscriptionService(store store.Store) *SubscriptionService {
	return &SubscriptionService{store: store}
}

// Subscribe создаёт подписку на новые комментарии к посту
func (s *SubscriptionService) Subscribe(postID string) (<-chan *store.Comment, func()) {
	return s.store.Subscribe(postID)
}

// Publish публикует новый комментарий
func (s *SubscriptionService) Publish(comment *store.Comment) {
	s.store.Publish(comment)
}
