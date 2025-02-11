package resolvers

import (
	"context"
	"time"

	"github.com/SobolevTim/t-graphql/internal/graph/model"
	"github.com/SobolevTim/t-graphql/internal/store"
)

// Comments возвращает комментарии к посту.
func (r *postResolver) Comments(ctx context.Context, obj *model.Post, page, pageSize *int) ([]*model.Comment, error) {
	comments, err := r.CommentService.GetCommentsByPostID(obj.ID, *page, *pageSize)
	if err != nil {
		return nil, err
	}

	var gqlComments []*model.Comment
	for _, comment := range comments {
		gqlComments = append(gqlComments, &model.Comment{
			ID:        comment.ID,
			PostID:    comment.PostID,
			ParentID:  comment.ParentID,
			Content:   comment.Content,
			Author:    comment.Author,
			CreatedAt: comment.CreatedAt.Format(time.RFC3339),
		})
	}

	return gqlComments, nil
}

// Replies возвращает ответы (вложенные комментарии) для конкретного комментария.
func (r *commentResolver) Replies(ctx context.Context, obj *model.Comment, page, pageSize *int) ([]*model.Comment, error) {
	// Получаем ответы для комментария с родительским идентификатором obj.ID.
	// Если метод GetCommentsByPostIDAndParentID ожидает указатели на int, передаём их.
	replies, err := r.CommentService.GetCommentsByPostIDAndParentID(obj.PostID, &obj.ID, *page, *pageSize)
	if err != nil {
		return nil, err
	}

	var gqlReplies []*model.Comment
	for _, reply := range replies {
		gqlReplies = append(gqlReplies, &model.Comment{
			ID:        reply.ID,
			PostID:    reply.PostID,
			ParentID:  reply.ParentID,
			Content:   reply.Content,
			Author:    reply.Author,
			CreatedAt: reply.CreatedAt.Format(time.RFC3339),
		})
	}
	return gqlReplies, nil
}

// AddComment создаёт комментарий.
func (r *mutationResolver) AddComment(ctx context.Context, input model.AddCommentInput) (*model.Comment, error) {
	comment, err := r.CommentService.AddComment(input.PostID, input.Content, input.Author, input.ParentID)
	if err != nil {
		return nil, err
	}

	// Публикуем новый комментарий для подписчиков
	r.SubscriptionService.Publish(&store.Comment{
		ID:        comment.ID,
		PostID:    comment.PostID,
		ParentID:  comment.ParentID,
		Content:   comment.Content,
		Author:    comment.Author,
		CreatedAt: comment.CreatedAt,
	})

	return &model.Comment{
		ID:        comment.ID,
		PostID:    comment.PostID,
		ParentID:  comment.ParentID,
		Content:   comment.Content,
		Author:    comment.Author,
		CreatedAt: comment.CreatedAt.Format(time.RFC3339),
	}, nil
}
