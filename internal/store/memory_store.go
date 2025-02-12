package store

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// MemoryStore — in-memory хранилище постов и комментариев
type MemoryStore struct {
	mu          sync.RWMutex               // Защита от гонок при доступе к хранилищу
	posts       map[string]*Post           // Посты
	comments    map[string][]*Comment      // Комментарии к постам
	subscribers map[string][]chan *Comment // Подписчики на новые комментарии
}

// NewMemoryStore создаёт новый in-memory store
// и инициализирует его пустыми списками постов и комментариев
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		posts:       make(map[string]*Post),
		comments:    make(map[string][]*Comment),
		subscribers: make(map[string][]chan *Comment),
	}
}

// Создание поста
// allowComments — разрешены ли комментарии к посту
func (s *MemoryStore) CreatePost(id, title, content, author string, allowComments bool) (*Post, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.posts[id]; exists {
		return nil, errors.New("post with this ID already exists")
	}

	post := &Post{
		ID:            id,
		Title:         title,
		Content:       content,
		Author:        author,
		CreatedAt:     time.Now(),
		AllowComments: allowComments,
	}
	s.posts[id] = post
	return post, nil
}

// Получение копии постов с пагинацией
// page — номер страницы, pageSize — количество постов на странице
func (s *MemoryStore) GetPosts(page, pageSize int) ([]*Post, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	posts := make([]*Post, 0, len(s.posts))
	for _, post := range s.posts {
		posts = append(posts, &Post{
			ID:            post.ID,
			Title:         post.Title,
			Content:       post.Content,
			Author:        post.Author,
			CreatedAt:     post.CreatedAt,
			AllowComments: post.AllowComments,
		})
	}

	start := (page - 1) * pageSize
	if start >= len(posts) {
		return []*Post{}, nil
	}

	end := start + pageSize
	if end > len(posts) {
		end = len(posts)
	}

	return posts[start:end], nil
}

// Получение поста по ID
func (s *MemoryStore) GetPostByID(id string) (*Post, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Поиск поста в хранилище
	post, exists := s.posts[id]
	if !exists {
		return nil, errors.New("post not found")
	}
	return post, nil
}

// Обновление разрешения на комментарии
// allowComments — разрешены ли комментарии к посту
func (s *MemoryStore) UpdatePostCommentsPermission(postID string, allowComments bool) (*Post, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Поиск поста в хранилище
	post, exists := s.posts[postID]
	if !exists {
		return nil, errors.New("post not found")
	}

	// Обновление разрешения на комментарии
	post.AllowComments = allowComments
	return post, nil
}

// Создание комментария
// parentID == nil — комментарий к посту
func (s *MemoryStore) CreateComment(id, postID string, parentID *string, content string, author string) (*Comment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	comment := &Comment{
		ID:        id,
		PostID:    postID,
		ParentID:  parentID,
		Content:   content,
		Author:    author,
		CreatedAt: time.Now(),
	}

	// Запись комментария в хранилище
	s.comments[postID] = append(s.comments[postID], comment)
	return comment, nil
}

// Получение комментариев по ID поста
// ParentID == nil — комментарий к посту
func (s *MemoryStore) GetCommentsByPostID(postID string, page, pageSize *int) ([]*Comment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*Comment

	// Пагинация
	start := 0
	end := len(s.comments[postID])

	// Если указаны page и pageSize, применяем пагинацию
	if page != nil && pageSize != nil {
		fmt.Println("page:", *page, "pageSize:", *pageSize)
		start = (*page - 1) * *pageSize
		end = start + *pageSize
		if start > len(s.comments[postID]) {
			start = len(s.comments[postID])
		}
		if end > len(s.comments[postID]) {
			end = len(s.comments[postID])
		}
	}
	// Формируем список комментариев к посту
	for i := start; i < end; i++ {
		c := s.comments[postID][i]
		if c.ParentID == nil {
			result = append(result, c)
		}
	}

	return result, nil
}

// Получение ответов на комментарий
func (s *MemoryStore) GetCommentsByPostIDAndParentID(postID string, parentID *string, page, pageSize *int) ([]*Comment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []*Comment

	// Отфильтруем комментарии по parentID
	for _, c := range s.comments[postID] {
		if (parentID == nil && c.ParentID == nil) ||
			(parentID != nil && c.ParentID != nil && *c.ParentID == *parentID) {
			filtered = append(filtered, c)
		}
	}

	// Применим пагинацию после фильтрации
	start, end := 0, len(filtered)
	if page != nil && pageSize != nil {
		start = (*page - 1) * *pageSize
		end = start + *pageSize

		// Гарантируем, что индексы в пределах массива
		if start >= len(filtered) {
			return []*Comment{}, nil // Пустой список, если страница выходит за границы
		}
		if end > len(filtered) {
			end = len(filtered)
		}
	}

	return filtered[start:end], nil
}

// Subscribe — добавляет подписчика
func (s *MemoryStore) Subscribe(postID string) (<-chan *Comment, func()) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Создаём новый канал для подписчика
	ch := make(chan *Comment, 1)
	s.subscribers[postID] = append(s.subscribers[postID], ch)

	// Функция для отписки
	unsubscribe := func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		channels := s.subscribers[postID]
		for i, c := range channels {
			if c == ch {
				s.subscribers[postID] = append(channels[:i], channels[i+1:]...)
				if !isClosed(c) {
					close(c)
				}
				break
			}
		}
	}

	return ch, unsubscribe
}

// isClosed проверяет, закрыт ли канал
func isClosed(ch chan *Comment) bool {
	select {
	case <-ch:
		return true // Канал закрыт
	default:
		return false // Канал открыт
	}
}

// Publish — отправляет новый комментарий подписчикам
func (s *MemoryStore) Publish(comment *Comment) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Поиск подписчиков поста
	channels, exists := s.subscribers[comment.PostID]
	if !exists {
		return
	}

	// Рассылаем комментарий всем подписчикам
	for _, ch := range channels {
		select {
		case ch <- comment: // Если клиент готов читать, отправляем
		default: // Если клиент не читает, пропускаем
			fmt.Printf("Warning: comment %s was not delivered to a subscriber\n", comment.ID)
		}
	}
}
