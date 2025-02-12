package service

import (
	"errors"
	"fmt"

	"github.com/SobolevTim/t-graphql/internal/store"
	"github.com/google/uuid"
)

const (
	defaultCommentSize     = 2000
	defaultCommentPageSize = 10
	defaultCommentPage     = 1
)

// CommentService отвечает за работу с комментариями
type CommentService struct {
	store store.Store
}

// NewCommentService создаёт сервис для работы с комментариями
func NewCommentService(store store.Store) *CommentService {
	return &CommentService{store: store}
}

// CreateComment создаёт новый комментарий к посту
func (s *CommentService) AddComment(postID, content, author string, ParentID *string) (*store.Comment, error) {
	// Проверка наличия поста
	post, err := s.store.GetPostByID(postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	// Проверка размера комментария
	if len(content) > defaultCommentSize {
		return nil, errors.New("comment is too long")
	}

	// Проверка разрешения на комментарии
	if !post.AllowComments {
		return nil, errors.New("comments are not allowed")
	}

	id := uuid.New().String()

	// Создаём комментарий
	comment, err := s.store.CreateComment(
		id,
		postID,
		ParentID,
		content,
		author,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	return comment, nil
}

// GetCommentsByPostID возвращает комментарии к посту c постраничным выводом
func (s *CommentService) GetCommentsByPostID(postID string, page, pageSize int) ([]*store.Comment, error) {
	if page <= 0 {
		page = defaultCommentPage
	}
	if pageSize <= 0 {
		pageSize = defaultCommentPageSize
	}
	return s.store.GetCommentsByPostID(postID, &page, &pageSize)
}

// GetCommentsByPostIDAndParentID возвращает ответы на комментарий к посту c постраничным выводом
func (s *CommentService) GetCommentsByPostIDAndParentID(postID string, parentID *string, page, pageSize int) ([]*store.Comment, error) {
	if page <= 0 {
		page = defaultCommentPage
	}
	if pageSize <= 0 {
		pageSize = defaultCommentPageSize
	}
	return s.store.GetCommentsByPostIDAndParentID(postID, parentID, &page, &pageSize)
}
