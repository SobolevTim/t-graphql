package resolvers

import (
	"context"
	"time"

	"github.com/SobolevTim/t-graphql/internal/graph/model"
)

// CommentAdded — резолвер для подписки на новые комментарии
func (r *subscriptionResolver) CommentAdded(ctx context.Context, postID string) (<-chan *model.Comment, error) {
	chStore, unsubscribe := r.SubscriptionService.Subscribe(postID)
	ch := make(chan *model.Comment)

	go func() {
		for comment := range chStore {
			ch <- &model.Comment{
				ID:        comment.ID,
				PostID:    comment.PostID,
				ParentID:  comment.ParentID,
				Content:   comment.Content,
				Author:    comment.Author,
				CreatedAt: comment.CreatedAt.Format(time.RFC3339),
			}
		}
		close(ch)
	}()

	// Отписка при завершении запроса
	go func() {
		<-ctx.Done()
		unsubscribe()
	}()

	return ch, nil
}
