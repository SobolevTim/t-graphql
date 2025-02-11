package service

import (
	"errors"

	"github.com/SobolevTim/t-graphql/internal/store"
	"github.com/google/uuid"
)

const (
	defaultPageSize      = 10
	defaultPage          = 1
	defaultAllowComments = true
)

type PostService struct {
	store store.Store
}

func NewPostService(store store.Store) *PostService {
	return &PostService{store: store}
}

// CreatePost создаёт новый пост
func (s *PostService) CreatePost(title, content, author string, allowComments bool) (*store.Post, error) {
	// Проверка обязательных полей
	if title == "" {
		return nil, errors.New("title is required")
	}
	if content == "" {
		return nil, errors.New("content is required")
	}
	if author == "" {
		return nil, errors.New("author is required")
	}
	if allowComments {
		allowComments = defaultAllowComments
	}

	// Генерация уникального идентификатора
	id := uuid.NewString()

	return s.store.CreatePost(id, title, content, author, allowComments)
}

// GetPosts возвращает список постов с пагинацией
func (s *PostService) GetPosts(page, pageSize int) ([]*store.Post, error) {
	if page <= 0 {
		page = defaultPage
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}

	return s.store.GetPosts(page, pageSize)
}

// GetPostByID возвращает пост по идентификатору
func (s *PostService) GetPostByID(id string) (*store.Post, error) {
	return s.store.GetPostByID(id)
}

// UpdatePostCommentsPermission обновляет разрешение на комментарии к посту
func (s *PostService) UpdatePostCommentsPermission(postID string, allowComments bool) (*store.Post, error) {
	return s.store.UpdatePostCommentsPermission(postID, allowComments)
}
