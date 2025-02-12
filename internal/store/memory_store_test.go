package store_test

import (
	"testing"
	"time"

	"github.com/SobolevTim/t-graphql/internal/store"
	"github.com/stretchr/testify/assert"
)

func TestCreatePost(t *testing.T) {
	memStore := store.NewMemoryStore()
	post, err := memStore.CreatePost("1", "Test Title", "Test Content", "Author", true)
	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.Equal(t, "1", post.ID)
	assert.Equal(t, "Test Title", post.Title)
	assert.Equal(t, "Test Content", post.Content)
	assert.Equal(t, "Author", post.Author)
	assert.True(t, post.AllowComments)
}

func TestGetPosts(t *testing.T) {
	memStore := store.NewMemoryStore()
	memStore.CreatePost("1", "Test Title 1", "Test Content 1", "Author 1", true)
	memStore.CreatePost("2", "Test Title 2", "Test Content 2", "Author 2", true)

	posts, err := memStore.GetPosts(1, 10)
	assert.NoError(t, err)
	assert.Len(t, posts, 2)
}

func TestGetPostByID(t *testing.T) {
	memStore := store.NewMemoryStore()
	memStore.CreatePost("1", "Test Title", "Test Content", "Author", true)

	post, err := memStore.GetPostByID("1")
	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.Equal(t, "1", post.ID)
}

func TestUpdatePostCommentsPermission(t *testing.T) {
	memStore := store.NewMemoryStore()
	memStore.CreatePost("1", "Test Title", "Test Content", "Author", true)

	post, err := memStore.UpdatePostCommentsPermission("1", false)
	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.False(t, post.AllowComments)
}

func TestCreateComment(t *testing.T) {
	memStore := store.NewMemoryStore()
	memStore.CreatePost("1", "Test Title", "Test Content", "Author", true)

	comment, err := memStore.CreateComment("1", "1", nil, "Test Comment", "Comment Author")
	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, "1", comment.ID)
	assert.Equal(t, "1", comment.PostID)
	assert.Nil(t, comment.ParentID)
	assert.Equal(t, "Test Comment", comment.Content)
	assert.Equal(t, "Comment Author", comment.Author)
}

func TestGetCommentsByPostID(t *testing.T) {
	memStore := store.NewMemoryStore()
	memStore.CreatePost("1", "Test Title", "Test Content", "Author", true)
	memStore.CreateComment("1", "1", nil, "Test Comment 1", "Comment Author 1")
	memStore.CreateComment("2", "1", nil, "Test Comment 2", "Comment Author 2")

	page := 1
	pageSize := 10
	comments, err := memStore.GetCommentsByPostID("1", &page, &pageSize)
	assert.NoError(t, err)
	assert.Len(t, comments, 2)
}

func TestGetCommentsByPostIDAndParentID(t *testing.T) {
	memStore := store.NewMemoryStore()
	memStore.CreatePost("1", "Test Title", "Test Content", "Author", true)
	parentID := "1"
	memStore.CreateComment("1", "1", nil, "Test Comment 1", "Comment Author 1")
	memStore.CreateComment("2", "1", &parentID, "Test Reply", "Reply Author")

	page := 1
	pageSize := 10
	comments, err := memStore.GetCommentsByPostIDAndParentID("1", &parentID, &page, &pageSize)
	assert.NoError(t, err)
	assert.Len(t, comments, 1)
}

func TestSubscribeAndPublish(t *testing.T) {
	memStore := store.NewMemoryStore()
	memStore.CreatePost("1", "Test Title", "Test Content", "Author", true)

	ch, unsubscribe := memStore.Subscribe("1")
	defer unsubscribe()

	comment := &store.Comment{
		ID:        "1",
		PostID:    "1",
		Content:   "Test Comment",
		Author:    "Comment Author",
		CreatedAt: time.Now(),
	}

	go memStore.Publish(comment)

	receivedComment := <-ch
	assert.NotNil(t, receivedComment)
	assert.Equal(t, "1", receivedComment.ID)
	assert.Equal(t, "Test Comment", receivedComment.Content)
}
