package store

import "time"

// Post представляет запись в блоге
// Если AllowReply == false, комментарии к посту запрещены
type Post struct {
	ID            string     // Уникальный идентификатор поста
	Title         string     // Заголовок поста
	Content       string     // Содержимое поста
	Author        string     // Автор поста
	CreatedAt     time.Time  // Время создания поста
	AllowComments bool       // Разрешены ли комментарии
	Comments      []*Comment // Комментарии к посту
}

// Comment представляет комментарий к посту
// Если ParentID == nil, значит комментарий верхнего уровня
type Comment struct {
	ID        string     // Уникальный идентификатор комментария
	PostID    string     // Уникальный идентификатор поста
	ParentID  *string    // Уникальный идентификатор родительского комментария
	Content   string     // Содержимое комментария
	Author    string     // Автор комментария
	CreatedAt time.Time  // Время создания комментария
	Replies   []*Comment // Комментарии к комментарию
}

// Store определяет методы работы с хранилищем
// При возникновении ошибки возвращается nil и ошибка
// Если метод возвращает список, а список пустой, возвращается пустой список и nil
type Store interface {
	// Методы работы с постами
	CreatePost(id, title, content, author string, allowComments bool) (*Post, error)
	GetPosts(page, pageSize int) ([]*Post, error)
	GetPostByID(id string) (*Post, error)
	UpdatePostCommentsPermission(postID string, allowComments bool) (*Post, error)

	// Методы работы с комментариями
	CreateComment(id, postID string, parentID *string, content string, author string) (*Comment, error)
	GetCommentsByPostID(postID string, page, pageSize int) ([]*Comment, error)                              // Только комментарии верхнего уровня
	GetCommentsByPostIDAndParentID(postID string, parentID *string, page, pageSize int) ([]*Comment, error) // Ответы на комментарий

	// Методы работы с подписками
	Subscribe(postID string) (<-chan *Comment, func())
	Publish(comment *Comment)
}
