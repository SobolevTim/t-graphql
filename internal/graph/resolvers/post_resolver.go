package resolvers

import (
	"context"
	"time"

	"github.com/SobolevTim/t-graphql/internal/graph/model"
)

// CreatePost создаёт новый пост.
func (r *mutationResolver) CreatePost(ctx context.Context, input model.CreatePostInput) (*model.Post, error) {
	post, err := r.PostService.CreatePost(input.Title, input.Content, input.Author, *input.AllowComments)
	if err != nil {
		return nil, err
	}

	return &model.Post{
		ID:            post.ID,
		Title:         post.Title,
		Content:       post.Content,
		AllowComments: post.AllowComments,
		CreatedAt:     post.CreatedAt.Format(time.RFC3339),
	}, nil
}

// UpdatePostCommentsPermission обновляет разрешение на комментарии.
func (r *mutationResolver) UpdatePostCommentsPermission(ctx context.Context, postID string, allowComments bool) (*model.Post, error) {
	post, err := r.PostService.UpdatePostCommentsPermission(postID, allowComments)
	if err != nil {
		return nil, err
	}

	return &model.Post{
		ID:            post.ID,
		Title:         post.Title,
		Content:       post.Content,
		AllowComments: post.AllowComments,
		CreatedAt:     post.CreatedAt.Format(time.RFC3339),
	}, nil
}

// Post возвращает пост по ID.
func (r *queryResolver) Post(ctx context.Context, id string) (*model.Post, error) {
	post, err := r.PostService.GetPostByID(id)
	if err != nil {
		return nil, err
	}

	// GraphQL сам вызовет r.Replies, когда запросят вложенные комментарии
	return &model.Post{
		ID:            post.ID,
		Title:         post.Title,
		Content:       post.Content,
		Author:        post.Author,
		AllowComments: post.AllowComments,
		CreatedAt:     post.CreatedAt.Format(time.RFC3339),
	}, nil
}

// Posts возвращает все посты.
func (r *queryResolver) Posts(ctx context.Context, page *int, pageSize *int) ([]*model.Post, error) {
	posts, err := r.PostService.GetPosts(*page, *pageSize)
	if err != nil {
		return nil, err
	}

	var gqlPosts []*model.Post
	for _, post := range posts {
		gqlPosts = append(gqlPosts, &model.Post{
			ID:            post.ID,
			Title:         post.Title,
			Content:       post.Content,
			Author:        post.Author,
			AllowComments: post.AllowComments,
			CreatedAt:     post.CreatedAt.Format(time.RFC3339),
		})
	}

	return gqlPosts, nil
}
