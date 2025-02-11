package resolvers

import (
	"github.com/SobolevTim/t-graphql/internal/graph/generated"
	"github.com/SobolevTim/t-graphql/internal/service"
)

// Resolver — главный резолвер, передающий зависимости.
type Resolver struct {
	PostService         *service.PostService
	CommentService      *service.CommentService
	SubscriptionService *service.SubscriptionService
}

// NewResolver — конструктор резолвера.
func NewResolver(postService *service.PostService, commentService *service.CommentService, subscriptionService *service.SubscriptionService) *Resolver {
	return &Resolver{
		PostService:         postService,
		CommentService:      commentService,
		SubscriptionService: subscriptionService,
	}
}

// Comment returns generated.CommentResolver implementation.
func (r *Resolver) Comment() generated.CommentResolver { return &commentResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Post returns generated.PostResolver implementation.
func (r *Resolver) Post() generated.PostResolver { return &postResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Subscription returns generated.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated.SubscriptionResolver { return &subscriptionResolver{r} }

type commentResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type postResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
