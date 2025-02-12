package service_test

import (
	"errors"
	"testing"

	"github.com/SobolevTim/t-graphql/internal/service"
	"github.com/SobolevTim/t-graphql/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockStore struct {
	mock.Mock
}

func (m *MockStore) CreatePost(id, title, content, author string, allowComments bool) (*store.Post, error) {
	args := m.Called(id, title, content, author, allowComments)
	return args.Get(0).(*store.Post), args.Error(1)
}

func (m *MockStore) GetPosts(page, pageSize int) ([]*store.Post, error) {
	args := m.Called(page, pageSize)
	return args.Get(0).([]*store.Post), args.Error(1)
}

func (m *MockStore) UpdatePostCommentsPermission(postID string, allowComments bool) (*store.Post, error) {
	args := m.Called(postID, allowComments)
	return args.Get(0).(*store.Post), args.Error(1)
}

func (m *MockStore) GetPostByID(postID string) (*store.Post, error) {
	args := m.Called(postID)
	return args.Get(0).(*store.Post), args.Error(1)
}

func (m *MockStore) CreateComment(id, postID string, parentID *string, content, author string) (*store.Comment, error) {
	args := m.Called(id, postID, parentID, content, author)
	return args.Get(0).(*store.Comment), args.Error(1)
}

func (m *MockStore) GetCommentsByPostID(postID string, page, pageSize *int) ([]*store.Comment, error) {
	args := m.Called(postID, page, pageSize)
	return args.Get(0).([]*store.Comment), args.Error(1)
}

func (m *MockStore) GetCommentsByPostIDAndParentID(postID string, parentID *string, page, pageSize *int) ([]*store.Comment, error) {
	args := m.Called(postID, parentID, page, pageSize)
	return args.Get(0).([]*store.Comment), args.Error(1)
}

func (m *MockStore) Subscribe(postID string) (<-chan *store.Comment, func()) {
	args := m.Called(postID)
	return args.Get(0).(<-chan *store.Comment), args.Get(1).(func())
}

func (m *MockStore) Publish(comment *store.Comment) {
	m.Called(comment)
}

func TestAddComment(t *testing.T) {
	mockStore := new(MockStore)
	commentService := service.NewCommentService(mockStore)

	postID := "post1"
	content := "This is a comment"
	author := "author1"
	parentID := "parent1"
	commentID := "comment1"

	mockStore.On("GetPostByID", postID).Return(&store.Post{ID: postID, AllowComments: true}, nil)
	mockStore.On("CreateComment", mock.Anything, postID, &parentID, content, author).Return(&store.Comment{ID: commentID}, nil)

	comment, err := commentService.AddComment(postID, content, author, &parentID)
	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, commentID, comment.ID)

	mockStore.AssertExpectations(t)
}

func TestAddComment_PostNotFound(t *testing.T) {
	mockStore := new(MockStore)
	commentService := service.NewCommentService(mockStore)

	postID := "post1"
	content := "This is a comment"
	author := "author1"
	parentID := "parent1"

	mockStore.On("GetPostByID", postID).Return(&store.Post{}, errors.New("post not found"))

	comment, err := commentService.AddComment(postID, content, author, &parentID)
	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, "failed to get post: post not found", err.Error())

	mockStore.AssertExpectations(t)
}

func TestAddComment_CommentTooLong(t *testing.T) {
	mockStore := new(MockStore)
	commentService := service.NewCommentService(mockStore)

	postID := "post1"
	content := make([]byte, 2001)
	author := "author1"
	parentID := "parent1"

	mockStore.On("GetPostByID", postID).Return(&store.Post{ID: postID, AllowComments: true}, nil)

	comment, err := commentService.AddComment(postID, string(content), author, &parentID)
	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, "comment is too long", err.Error())

	mockStore.AssertExpectations(t)
}

func TestAddComment_CommentsNotAllowed(t *testing.T) {
	mockStore := new(MockStore)
	commentService := service.NewCommentService(mockStore)

	postID := "post1"
	content := "This is a comment"
	author := "author1"
	parentID := "parent1"

	mockStore.On("GetPostByID", postID).Return(&store.Post{ID: postID, AllowComments: false}, nil)

	comment, err := commentService.AddComment(postID, content, author, &parentID)
	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, "comments are not allowed", err.Error())

	mockStore.AssertExpectations(t)
}

func TestGetCommentsByPostID(t *testing.T) {
	mockStore := new(MockStore)
	commentService := service.NewCommentService(mockStore)

	postID := "post1"
	page := 1
	pageSize := 10
	comments := []*store.Comment{{ID: "comment1"}, {ID: "comment2"}}

	mockStore.On("GetCommentsByPostID", postID, &page, &pageSize).Return(comments, nil)

	result, err := commentService.GetCommentsByPostID(postID, page, pageSize)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, comments, result)

	mockStore.AssertExpectations(t)
}

func TestGetCommentsByPostIDAndParentID(t *testing.T) {
	mockStore := new(MockStore)
	commentService := service.NewCommentService(mockStore)

	postID := "post1"
	parentID := "parent1"
	page := 1
	pageSize := 10
	comments := []*store.Comment{{ID: "comment1"}, {ID: "comment2"}}

	mockStore.On("GetCommentsByPostIDAndParentID", postID, &parentID, &page, &pageSize).Return(comments, nil)

	result, err := commentService.GetCommentsByPostIDAndParentID(postID, &parentID, page, pageSize)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, comments, result)

	mockStore.AssertExpectations(t)
}

func TestCreatePost(t *testing.T) {
	mockStore := new(MockStore)
	postService := service.NewPostService(mockStore)

	title := "Test Title"
	content := "Test Content"
	author := "Test Author"
	allowComments := true
	postID := "post1"

	mockStore.On("CreatePost", mock.Anything, title, content, author, allowComments).Return(&store.Post{ID: postID}, nil)

	post, err := postService.CreatePost(title, content, author, allowComments)
	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.Equal(t, postID, post.ID)

	mockStore.AssertExpectations(t)
}

func TestCreatePost_MissingFields(t *testing.T) {
	mockStore := new(MockStore)
	postService := service.NewPostService(mockStore)

	_, err := postService.CreatePost("", "content", "author", true)
	assert.Error(t, err)
	assert.Equal(t, "title is required", err.Error())

	_, err = postService.CreatePost("title", "", "author", true)
	assert.Error(t, err)
	assert.Equal(t, "content is required", err.Error())

	_, err = postService.CreatePost("title", "content", "", true)
	assert.Error(t, err)
	assert.Equal(t, "author is required", err.Error())
}

func TestGetPosts(t *testing.T) {
	mockStore := new(MockStore)
	postService := service.NewPostService(mockStore)

	page := 1
	pageSize := 10
	posts := []*store.Post{{ID: "post1"}, {ID: "post2"}}

	mockStore.On("GetPosts", page, pageSize).Return(posts, nil)

	result, err := postService.GetPosts(page, pageSize)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, posts, result)

	mockStore.AssertExpectations(t)
}

func TestGetPostByID(t *testing.T) {
	mockStore := new(MockStore)
	postService := service.NewPostService(mockStore)

	postID := "post1"
	post := &store.Post{ID: postID}

	mockStore.On("GetPostByID", postID).Return(post, nil)

	result, err := postService.GetPostByID(postID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, post, result)

	mockStore.AssertExpectations(t)
}

func TestUpdatePostCommentsPermission(t *testing.T) {
	mockStore := new(MockStore)
	postService := service.NewPostService(mockStore)

	postID := "post1"
	allowComments := true
	post := &store.Post{ID: postID, AllowComments: allowComments}

	mockStore.On("UpdatePostCommentsPermission", postID, allowComments).Return(post, nil)

	result, err := postService.UpdatePostCommentsPermission(postID, allowComments)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, post, result)

	mockStore.AssertExpectations(t)
}

func TestSubscribe(t *testing.T) {
	mockStore := new(MockStore)
	subscriptionService := service.NewSubscriptionService(mockStore)

	postID := "post1"
	Chan := make(chan *store.Comment)
	unsubscribeFunc := func() {}

	mockStore.On("Subscribe", postID).Return((<-chan *store.Comment)(Chan), unsubscribeFunc)

	resultChan, resultFunc := subscriptionService.Subscribe(postID)
	assert.NotNil(t, resultChan)
	assert.NotNil(t, resultFunc)

	mockStore.AssertExpectations(t)
}

func TestPublish(t *testing.T) {
	mockStore := new(MockStore)
	subscriptionService := service.NewSubscriptionService(mockStore)

	comment := &store.Comment{ID: "comment1"}

	mockStore.On("Publish", comment).Return()

	subscriptionService.Publish(comment)

	mockStore.AssertExpectations(t)
}
