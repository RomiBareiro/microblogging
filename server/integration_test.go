//go:build integration
// +build integration

package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"

	"microblogging/model"
	"microblogging/repository"
	"microblogging/server"
	"microblogging/service"
)

type integrationServer interface {
	CreatePostHandler(http.ResponseWriter, *http.Request)
	GetTimelineHandler(http.ResponseWriter, *http.Request)
}

func setupPostgresContainer(ctx context.Context, t *testing.T) (testcontainers.Container, string) {
	t.Helper()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		Env:          map[string]string{"POSTGRES_USER": "test", "POSTGRES_PASSWORD": "test", "POSTGRES_DB": "testdb"},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp").WithStartupTimeout(2 * time.Minute),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	assert.NoError(t, err)

	host, err := container.Host(ctx)
	assert.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	assert.NoError(t, err)

	dsn := fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", host, port.Port())
	return container, dsn
}

func runSQLFile(t *testing.T, db *sqlx.DB, path string) {
	t.Helper()

	content, err := os.ReadFile(path)
	assert.NoError(t, err)

	_, err = db.Exec(string(content))
	assert.NoError(t, err)
}

func newTestServer(t *testing.T, dsn string) integrationServer {
	t.Helper()

	db, err := sqlx.Connect("postgres", dsn)
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	repo := &repository.DBConnector{DB: db, Logger: zap.NewNop()}
	svc := service.NewBlogService(repo)
	return server.NewServer(context.Background(), svc)
}

func TestCreatePostAndTimelineIntegration(t *testing.T) {
	ctx := context.Background()
	container, dsn := setupPostgresContainer(ctx, t)
	defer func() {
		err := container.Terminate(ctx)
		assert.NoError(t, err)
	}()

	db, err := sqlx.Connect("postgres", dsn)
	assert.NoError(t, err)
	defer db.Close()

	createScript := filepath.Join("..", "config", "db_creation", "1-table_creation.sql")
	runSQLFile(t, db, createScript)

	srv := newTestServer(t, dsn)

	userID := "11111111-1111-1111-1111-111111111111"
	_, err = db.Exec(
		`INSERT INTO users (id, user_name, email, password, created_at, updated_at) VALUES ($1, $2, $3, $4, now(), now())`,
		userID, "alice", "alice@example.com", "password123",
	)
	assert.NoError(t, err)

	postReq := model.CreatePostRequest{UserID: userID, Content: "Integration test post"}
	body, err := json.Marshal(postReq)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/V1/post", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.CreatePostHandler(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var createResp struct {
		Message string `json:"message"`
		Data    struct {
			UserID string `json:"user_id"`
			PostID string `json:"post_id"`
		} `json:"data"`
	}
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&createResp))
	assert.Equal(t, userID, createResp.Data.UserID)
	assert.NotEmpty(t, createResp.Data.PostID)

	timelineURL := fmt.Sprintf("/V1/timeline?user_id=%s&limit=10&before=2000-01-01T00:00:00Z", userID)
	timelineReq := httptest.NewRequest(http.MethodGet, timelineURL, nil)
	w = httptest.NewRecorder()
	srv.GetTimelineHandler(w, timelineReq)
	assert.Equal(t, http.StatusOK, w.Code)

	var timelineResp struct {
		Message string `json:"message"`
		Data    struct {
			UserID string                 `json:"user_id"`
			Posts  model.TimelineResponse `json:"posts"`
		} `json:"data"`
	}
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&timelineResp))
	assert.Equal(t, userID, timelineResp.Data.UserID)
	assert.Len(t, timelineResp.Data.Posts.Posts, 1)
	assert.Equal(t, "Integration test post", timelineResp.Data.Posts.Posts[0].Content)
}
