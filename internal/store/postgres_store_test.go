package store

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testStore *Service

// TestMain выполняет инициализацию подключения к тестовой базе,
// создаёт необходимые таблицы и по завершении тестов их удаляет.
func TestMain(m *testing.M) {
	// Получаем URL подключения к тестовой БД из переменной окружения
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		fmt.Println("TEST_DATABASE_URL is not set")
		os.Exit(1)
	}

	var err error
	testStore, err = NewPostgresStore(dbURL)
	if err != nil {
		fmt.Printf("Failed to connect to test database: %v\n", err)
		os.Exit(1)
	}

	// Создаём таблицы для тестов
	if err := setupTables(testStore.DB); err != nil {
		fmt.Printf("Failed to setup tables: %v\n", err)
		os.Exit(1)
	}

	// Запускаем тесты
	code := m.Run()

	// Удаляем таблицы после тестирования
	if err := teardownTables(testStore.DB); err != nil {
		fmt.Printf("Failed to teardown tables: %v\n", err)
	}
	testStore.Close()
	os.Exit(code)
}

// setupTables создаёт таблицы posts и comments, если их ещё нет.
func setupTables(db *pgxpool.Pool) error {
	ctx := context.Background()
	queries := []string{
		`CREATE TABLE IF NOT EXISTS posts (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			author TEXT NOT NULL,
			allow_comments BOOLEAN NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS comments (
			id TEXT PRIMARY KEY,
			post_id TEXT NOT NULL,
			parent_id TEXT,
			content TEXT NOT NULL,
			author TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		);`,
	}
	for _, q := range queries {
		if _, err := db.Exec(ctx, q); err != nil {
			return err
		}
	}
	return nil
}

// teardownTables удаляет таблицы posts и comments.
func teardownTables(db *pgxpool.Pool) error {
	ctx := context.Background()
	queries := []string{
		`DROP TABLE IF EXISTS comments;`,
		`DROP TABLE IF EXISTS posts;`,
	}
	for _, q := range queries {
		if _, err := db.Exec(ctx, q); err != nil {
			return err
		}
	}
	return nil
}

// cleanTables очищает данные из таблиц posts и comments.
func cleanTables(s *Service) error {
	ctx := context.Background()
	_, err := s.DB.Exec(ctx, "TRUNCATE TABLE comments, posts;")
	return err
}

// TestCreatePost проверяет создание поста.
func TestCreatePost(t *testing.T) {
	if err := cleanTables(testStore); err != nil {
		t.Fatalf("failed to clean tables: %v", err)
	}

	postID := uuid.NewString()
	title := "Test Post"
	content := "This is a test post."
	author := "tester"
	allowComments := true

	post, err := testStore.CreatePost(postID, title, content, author, allowComments)
	if err != nil {
		t.Fatalf("CreatePost failed: %v", err)
	}

	if post.ID != postID || post.Title != title || post.Content != content ||
		post.Author != author || post.AllowComments != allowComments {
		t.Errorf("created post does not match input: %+v", post)
	}
}

// TestGetPosts проверяет выборку постов с пагинацией.
func TestGetPosts(t *testing.T) {
	if err := cleanTables(testStore); err != nil {
		t.Fatalf("failed to clean tables: %v", err)
	}

	post1ID := uuid.NewString()
	post2ID := uuid.NewString()
	post3ID := uuid.NewString()

	// Создаём несколько постов
	postsData := []struct {
		id    string
		title string
	}{
		{post1ID, "Post 1"},
		{post2ID, "Post 2"},
		{post3ID, "Post 3"},
	}
	for _, pd := range postsData {
		if _, err := testStore.CreatePost(pd.id, pd.title, "Content", "Author", true); err != nil {
			t.Fatalf("failed to create post %s: %v", pd.id, err)
		}
		// Небольшая задержка для гарантии различного времени создания
		time.Sleep(10 * time.Millisecond)
	}

	// Запрашиваем первую страницу с 2 записями (сортировка по created_at DESC)
	posts, err := testStore.GetPosts(1, 2)
	if err != nil {
		t.Fatalf("GetPosts failed: %v", err)
	}
	if len(posts) != 2 {
		t.Errorf("expected 2 posts, got %d", len(posts))
	}
	// При сортировке DESC первым должен идти последний созданный пост ("3")
	if posts[0].ID != post3ID || posts[1].ID != post2ID {
		t.Errorf("unexpected order of posts: got %v", []string{posts[0].ID, posts[1].ID})
	}
}

// TestGetPostByID проверяет получение поста по его идентификатору.
func TestGetPostByID(t *testing.T) {
	if err := cleanTables(testStore); err != nil {
		t.Fatalf("failed to clean tables: %v", err)
	}

	postID := uuid.NewString()
	title := "Test Post"
	content := "This is a test post."
	author := "tester"

	createdPost, err := testStore.CreatePost(postID, title, content, author, true)
	if err != nil {
		t.Fatalf("CreatePost failed: %v", err)
	}

	fetchedPost, err := testStore.GetPostByID(postID)
	if err != nil {
		t.Fatalf("GetPostByID failed: %v", err)
	}
	if fetchedPost.ID != createdPost.ID || fetchedPost.Title != createdPost.Title {
		t.Errorf("fetched post does not match created post: %+v vs %+v", fetchedPost, createdPost)
	}
}

// TestUpdatePostCommentsPermission проверяет обновление разрешения на комментарии.
func TestUpdatePostCommentsPermission(t *testing.T) {
	if err := cleanTables(testStore); err != nil {
		t.Fatalf("failed to clean tables: %v", err)
	}

	postID := uuid.NewString()
	_, err := testStore.CreatePost(postID, "Test Post", "This is a test post.", "tester", false)
	if err != nil {
		t.Fatalf("CreatePost failed: %v", err)
	}

	updatedPost, err := testStore.UpdatePostCommentsPermission(postID, true)
	if err != nil {
		t.Fatalf("UpdatePostCommentsPermission failed: %v", err)
	}
	if !updatedPost.AllowComments {
		t.Errorf("expected allowComments to be true, got false")
	}
}

// TestCreateComment проверяет создание комментария.
func TestCreateComment(t *testing.T) {
	if err := cleanTables(testStore); err != nil {
		t.Fatalf("failed to clean tables: %v", err)
	}

	// Создаём пост для комментария
	postID := uuid.NewString()
	if _, err := testStore.CreatePost(postID, "Test Post", "Content", "tester", true); err != nil {
		t.Fatalf("CreatePost failed: %v", err)
	}

	commentID := uuid.NewString()
	content := "Nice post!"
	author := "commenter"
	// parentID == nil означает верхний уровень
	comment, err := testStore.CreateComment(commentID, postID, nil, content, author)
	if err != nil {
		t.Fatalf("CreateComment failed: %v", err)
	}
	if comment.ID != commentID || comment.Content != content || comment.Author != author {
		t.Errorf("created comment does not match input: %+v", comment)
	}
}

// TestGetCommentsByPostID проверяет выборку верхнеуровневых комментариев для поста.
func TestGetCommentsByPostID(t *testing.T) {
	if err := cleanTables(testStore); err != nil {
		t.Fatalf("failed to clean tables: %v", err)
	}

	// Создаём пост
	postID := uuid.NewString()
	if _, err := testStore.CreatePost(postID, "Test Post", "Content", "tester", true); err != nil {
		t.Fatalf("CreatePost failed: %v", err)
	}

	// Создаём два верхнеуровневых комментария
	comment1ID := uuid.NewString()
	comment2ID := uuid.NewString()

	commentIDs := []string{comment1ID, comment2ID}
	for _, cid := range commentIDs {
		if _, err := testStore.CreateComment(cid, postID, nil, "Comment "+cid, "commenter"); err != nil {
			t.Fatalf("CreateComment %s failed: %v", cid, err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	// Создаём ответ на комментарий c1 (не должен попадать в выборку верхнеуровневых)
	parentID := uuid.NewString()
	comment3ID := uuid.NewString()
	if _, err := testStore.CreateComment(comment3ID, postID, &parentID, "Reply to c1", "replyer"); err != nil {
		t.Fatalf("CreateComment reply failed: %v", err)
	}

	page := 1
	pageSize := 10
	comments, err := testStore.GetCommentsByPostID(postID, page, pageSize)
	if err != nil {
		t.Fatalf("GetCommentsByPostID failed: %v", err)
	}
	// Ожидаем получить только два верхнеуровневых комментария
	if len(comments) != 2 {
		t.Errorf("expected 2 top-level comments, got %d", len(comments))
	}
}

// TestGetCommentsByPostIDAndParentID проверяет выборку ответов (reply) на конкретный комментарий.
func TestGetCommentsByPostIDAndParentID(t *testing.T) {
	if err := cleanTables(testStore); err != nil {
		t.Fatalf("failed to clean tables: %v", err)
	}

	// Создаём пост
	postID := uuid.NewString()
	if _, err := testStore.CreatePost(postID, "Test Post", "Content", "tester", true); err != nil {
		t.Fatalf("CreatePost failed: %v", err)
	}

	// Создаём верхнеуровневый комментарий
	parentCommentID := uuid.NewString()
	if _, err := testStore.CreateComment(parentCommentID, postID, nil, "Top-level comment", "commenter"); err != nil {
		t.Fatalf("CreateComment failed: %v", err)
	}
	// Создаём два ответа на комментарий c1
	commentC1ID := uuid.NewString()
	commentC2ID := uuid.NewString()
	replyIDs := []string{commentC1ID, commentC2ID}
	for _, rid := range replyIDs {
		if _, err := testStore.CreateComment(rid, postID, &parentCommentID, "Reply "+rid, "replyer"); err != nil {
			t.Fatalf("CreateComment reply %s failed: %v", rid, err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	page := 1
	pageSize := 10
	replies, err := testStore.GetCommentsByPostIDAndParentID(postID, &parentCommentID, page, pageSize)
	if err != nil {
		t.Fatalf("GetCommentsByPostIDAndParentID failed: %v", err)
	}
	if len(replies) != 2 {
		t.Errorf("expected 2 replies, got %d", len(replies))
	}
}

// TestSubscribeAndPublish проверяет работу подписки на уведомления о новых комментариях и их публикацию.
func TestSubscribeAndPublish(t *testing.T) {
	if err := cleanTables(testStore); err != nil {
		t.Fatalf("failed to clean tables: %v", err)
	}

	// Создаём пост
	postID := uuid.NewString()
	if _, err := testStore.CreatePost(postID, "Test Post", "Content", "tester", true); err != nil {
		t.Fatalf("CreatePost failed: %v", err)
	}

	commentID := uuid.NewString()
	content := "Subscription comment"
	author := "subscriber"
	commentToPublish := &Comment{
		ID:        commentID,
		PostID:    postID,
		ParentID:  nil,
		Content:   content,
		Author:    author,
		CreatedAt: time.Now(),
	}

	// Подписываемся на уведомления для поста
	ch, unsubscribe := testStore.Subscribe(postID)
	if ch == nil || unsubscribe == nil {
		t.Fatalf("Subscribe returned nil channel or unsubscribe function")
	}
	defer unsubscribe()

	// Небольшая задержка для установления подписки
	time.Sleep(100 * time.Millisecond)

	// Публикуем комментарий
	testStore.Publish(commentToPublish)

	// Ждём уведомления с таймаутом
	select {
	case receivedComment := <-ch:
		// Если уведомление получено, проверяем его содержимое.
		// Обратите внимание: Publish использует JSON, поэтому могут быть неточности с формированием времени.
		// Здесь проверяем только основные поля.
		if receivedComment.ID != commentID || receivedComment.Content != content || receivedComment.Author != author {
			t.Errorf("received comment does not match published comment: %+v", receivedComment)
		}
	case <-time.After(2 * time.Second):
		t.Errorf("timed out waiting for published comment")
	}
}
