package store

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	DB *pgxpool.Pool
}

// NewPostgresStore создаёт новый экземпляр сервиса
// для работы с базой данных PostgreSQL
func NewPostgresStore(url string) (*Service, error) {
	db, err := pgxpool.New(context.Background(), url)
	if err != nil {
		return nil, fmt.Errorf("could not connect to db: %w", err)
	}

	if err := db.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("could not ping db: %w", err)
	}

	return &Service{DB: db}, nil
}

// Close закрывает соединение с базой данных
func (s *Service) Close() {
	s.DB.Close()
}

// Создание поста
// allowComments — разрешены ли комментарии к посту
func (s *Service) CreatePost(id, title, content, author string, allowComments bool) (*Post, error) {
	query := `
        INSERT INTO posts (id, title, content, author, allow_comments, created_at)
        VALUES ($1, $2, $3, $4, $5, NOW())
        RETURNING id, title, content, author, allow_comments, created_at
        `
	// Выполнение запроса
	row := s.DB.QueryRow(context.Background(), query, id, title, content, author, allowComments)

	post := &Post{}
	if err := row.Scan(&post.ID, &post.Title, &post.Content, &post.Author, &post.AllowComments, &post.CreatedAt); err != nil {
		return nil, fmt.Errorf("could not create post: %w", err)
	}

	return post, nil
}

// Получение копии постов с пагинацией
// page — номер страницы, pageSize — количество постов на странице
func (s *Service) GetPosts(page, pageSize int) ([]*Post, error) {
	query := `
		SELECT id, title, content, author, allow_comments, created_at
		FROM posts
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
		`
	// Выполнение запроса
	rows, err := s.DB.Query(context.Background(), query, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("could not get posts: %w", err)
	}
	defer rows.Close()

	// Обработка результатов запроса
	posts := make([]*Post, 0)
	for rows.Next() {
		post := &Post{}
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Author, &post.AllowComments, &post.CreatedAt); err != nil {
			return nil, fmt.Errorf("could not read post: %w", err)
		}
		posts = append(posts, post)
	}

	return posts, nil
}

// Получение поста по ID
func (s *Service) GetPostByID(id string) (*Post, error) {
	query := `
		SELECT id, title, content, author, allow_comments, created_at
		FROM posts
		WHERE id = $1
		`
	// Выполнение запроса
	row := s.DB.QueryRow(context.Background(), query, id)

	// Обработка результата запроса
	post := &Post{}
	if err := row.Scan(&post.ID, &post.Title, &post.Content, &post.Author, &post.AllowComments, &post.CreatedAt); err != nil {
		return nil, fmt.Errorf("could not get post: %w", err)
	}

	return post, nil
}

// Обновление разрешения на комментарии к посту
func (s *Service) UpdatePostCommentsPermission(postID string, allowComments bool) (*Post, error) {
	query := `
		UPDATE posts
		SET allow_comments = $1
		WHERE id = $2
		RETURNING id, title, content, author, allow_comments, created_at
		`
	// Выполнение запроса
	row := s.DB.QueryRow(context.Background(), query, allowComments, postID)

	// Обработка результата запроса
	post := &Post{}
	if err := row.Scan(&post.ID, &post.Title, &post.Content, &post.Author, &post.AllowComments, &post.CreatedAt); err != nil {
		return nil, fmt.Errorf("could not update post: %w", err)
	}

	return post, nil
}

// Создание комментария
// parentID — ID родительского комментария
func (s *Service) CreateComment(id, postID string, parentID *string, content, author string) (*Comment, error) {
	query := `
		INSERT INTO comments (id, post_id, parent_id, content, author, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, post_id, parent_id, content, author, created_at
		`
	// Выполнение запроса
	row := s.DB.QueryRow(context.Background(), query, id, postID, parentID, content, author)

	// Обработка результата запроса
	comment := &Comment{}
	if err := row.Scan(&comment.ID, &comment.PostID, &comment.ParentID, &comment.Content, &comment.Author, &comment.CreatedAt); err != nil {
		return nil, fmt.Errorf("could not create comment: %w", err)
	}

	return comment, nil
}

// Получение комментариев к посту с пагинацией
func (s *Service) GetCommentsByPostID(postID string, page, pageSize int) ([]*Comment, error) {
	query := `
		SELECT id, post_id, parent_id, content, author, created_at
		FROM comments
		WHERE post_id = $1 AND parent_id IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
		`
	// Выполнение запроса
	rows, err := s.DB.Query(context.Background(), query, postID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("could not get comments: %w", err)
	}
	defer rows.Close()

	// Обработка результатов запроса
	comments := make([]*Comment, 0)
	for rows.Next() {
		comment := &Comment{}
		if err := rows.Scan(&comment.ID, &comment.PostID, &comment.ParentID, &comment.Content, &comment.Author, &comment.CreatedAt); err != nil {
			return nil, fmt.Errorf("could not read comment: %w", err)
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

// Получение ответов на комментарий с пагинацией
func (s *Service) GetCommentsByPostIDAndParentID(postID string, parentID *string, page, pageSize int) ([]*Comment, error) {
	query := `
		SELECT id, post_id, parent_id, content, author, created_at
		FROM comments
		WHERE post_id = $1 AND parent_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
		`
	// Выполнение запроса
	rows, err := s.DB.Query(context.Background(), query, postID, parentID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("could not get comments: %w", err)
	}
	defer rows.Close()

	// Обработка результатов запроса
	comments := make([]*Comment, 0)
	for rows.Next() {
		comment := &Comment{}
		if err := rows.Scan(&comment.ID, &comment.PostID, &comment.ParentID, &comment.Content, &comment.Author, &comment.CreatedAt); err != nil {
			return nil, fmt.Errorf("could not read comment: %w", err)
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

// Subscribe — добавляет подписчика на новые комментарии к посту
func (s *Service) Subscribe(postID string) (<-chan *Comment, func()) {
	ch := make(chan *Comment)
	// Создание контекста для отмены подписки
	ctx, cancel := context.WithCancel(context.Background())
	var once sync.Once

	// Подключение к базе данных
	conn, err := s.DB.Acquire(ctx)
	if err != nil {
		log.Printf("could not acquire connection: %v", err)
		cancel()
		return nil, nil
	}

	// Подписка на канал уведомлений
	_, err = conn.Exec(ctx, fmt.Sprintf(`LISTEN "comments_%s"`, postID))
	if err != nil {
		log.Printf("could not listen to channel: %v", err)
		cancel()
		conn.Release()
		return nil, nil
	}

	// Получение уведомлений
	go func() {
		defer func() {
			once.Do(func() {
				close(ch)      // Закрытие канала
				conn.Release() // Освобождение соединения
			})
		}()

		for {
			// Ожидание уведомления
			notification, err := conn.Conn().WaitForNotification(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("error while waiting for notification: %v", err)
				return
			}

			// Разбор уведомления
			comment := &Comment{}
			if err := json.Unmarshal([]byte(notification.Payload), comment); err != nil {
				log.Printf("could not unmarshal comment: %v", err)
				continue
			}

			// Отправляем уведомление с учетом возможности отмены
			select {
			case ch <- comment: // Отправка уведомления
			case <-ctx.Done(): // Отмена подписки
				return
			}
		}
	}()

	unsubscribe := func() {
		// Отмена контекста прервет WaitForNotification
		cancel()
		once.Do(func() {
			close(ch)      // Закрытие канала
			conn.Release() // Освобождение соединения
		})
	}

	return ch, unsubscribe
}

// Publish — публикация комментария
func (s *Service) Publish(comment *Comment) {
	// Размещение комментария в канале уведомлений
	payload, err := json.Marshal(comment)
	if err != nil {
		log.Printf("could not marshal comment: %v", err)
		return
	}

	// Отправка уведомления
	_, err = s.DB.Exec(context.Background(), fmt.Sprintf(`NOTIFY "comments_%s", '%s'`, comment.PostID, payload))
	if err != nil {
		log.Printf("could not notify channel: %v", err)
	}
}
