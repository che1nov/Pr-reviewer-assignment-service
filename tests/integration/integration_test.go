package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/che1nov/Pr-reviewer-assignment-service/config"
	"github.com/che1nov/Pr-reviewer-assignment-service/internal/adapters/postgresql"
	"github.com/che1nov/Pr-reviewer-assignment-service/internal/app"
	"github.com/che1nov/Pr-reviewer-assignment-service/pkg/logger"
)

const (
	testDBURL  = "postgres://app:app@localhost:5432/pr_service_test?sslmode=disable"
	adminToken = "test-admin-token"
	userToken  = "test-user-token"
)

type testServer struct {
	server *httptest.Server
	db     *sqlx.DB
}

func setupTestServer(t *testing.T) *testServer {
	t.Helper()

	db, err := sqlx.Connect("postgres", testDBURL)
	if err != nil {
		t.Skipf("Не удалось подключиться к тестовой БД: %v. Запустите: docker run -d -p 5432:5432 -e POSTGRES_DB=pr_service_test -e POSTGRES_USER=app -e POSTGRES_PASSWORD=app postgres:17", err)
		return nil
	}

	cleanupDB(t, db)

	if err := postgresql.RunMigrations(db.DB, logger.New(logger.Config{Level: "error"})); err != nil {
		t.Fatalf("Не удалось применить миграции: %v", err)
	}

	cfg := config.Config{
		HTTPPort:    "0",
		LogLevel:    "error",
		AdminToken:  adminToken,
		UserToken:   userToken,
		DatabaseURL: testDBURL,
	}

	log := logger.New(logger.Config{Level: cfg.LogLevel})
	application, err := app.New(cfg, log)
	if err != nil {
		t.Fatalf("Не удалось создать приложение: %v", err)
	}

	handler := application.Handler()
	server := httptest.NewServer(handler)

	return &testServer{
		server: server,
		db:     db,
	}
}

func (ts *testServer) Close() {
	ts.server.Close()
	_ = ts.db.Close()
}

func cleanupDB(t *testing.T, db *sqlx.DB) {
	t.Helper()
	_, _ = db.Exec("DROP TABLE IF EXISTS pull_requests, teams, users, schema_migrations")
}

func TestFullWorkflow(t *testing.T) {
	ts := setupTestServer(t)
	if ts == nil {
		return
	}
	defer ts.Close()

	team := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "user1", "username": "Alice", "is_active": true},
			{"user_id": "user2", "username": "Bob", "is_active": true},
			{"user_id": "user3", "username": "Charlie", "is_active": true},
			{"user_id": "user4", "username": "Dave", "is_active": true},
		},
	}

	resp := makeRequest(t, ts, "POST", "/team/add", team, adminToken)
	assertEqual(t, http.StatusCreated, resp.StatusCode, "Создание команды")
	defer closeResponseBody(t, resp)

	var teamResp map[string]interface{}
	mustDecodeJSON(t, resp, &teamResp)

	resp = makeRequest(t, ts, "GET", "/team/get?team_name=backend", nil, userToken)
	assertEqual(t, http.StatusOK, resp.StatusCode, "Получение команды")
	defer closeResponseBody(t, resp)

	pr := map[string]interface{}{
		"pull_request_id":   "pr-1",
		"pull_request_name": "Add feature",
		"author_id":         "user1",
	}

	resp = makeRequest(t, ts, "POST", "/pullRequest/create", pr, adminToken)
	assertEqual(t, http.StatusCreated, resp.StatusCode, "Создание PR")
	defer closeResponseBody(t, resp)

	var prResp map[string]interface{}
	mustDecodeJSON(t, resp, &prResp)
	prData := prResp["pr"].(map[string]interface{})

	reviewers := prData["assigned_reviewers"].([]interface{})
	if len(reviewers) != 2 {
		t.Errorf("Ожидалось 2 ревьювера, получено %d", len(reviewers))
	}

	assertEqual(t, "OPEN", prData["status"], "Статус PR должен быть OPEN")

	reviewer := reviewers[0].(string)
	resp = makeRequest(t, ts, "GET", fmt.Sprintf("/users/getReview?user_id=%s", reviewer), nil, userToken)
	assertEqual(t, http.StatusOK, resp.StatusCode, "Получение PR для ревьювера")
	defer closeResponseBody(t, resp)

	var reviewerPRs map[string]interface{}
	mustDecodeJSON(t, resp, &reviewerPRs)
	prs := reviewerPRs["pull_requests"].([]interface{})
	if len(prs) != 1 {
		t.Errorf("Ожидался 1 PR для ревьювера, получено %d", len(prs))
	}

	reassign := map[string]interface{}{
		"pull_request_id": "pr-1",
		"old_user_id":     reviewer,
	}

	resp = makeRequest(t, ts, "POST", "/pullRequest/reassign", reassign, adminToken)
	assertEqual(t, http.StatusOK, resp.StatusCode, "Переназначение ревьювера")
	defer closeResponseBody(t, resp)

	merge := map[string]interface{}{
		"pull_request_id": "pr-1",
	}

	resp = makeRequest(t, ts, "POST", "/pullRequest/merge", merge, adminToken)
	assertEqual(t, http.StatusOK, resp.StatusCode, "Merge PR")
	defer closeResponseBody(t, resp)

	var mergeResp map[string]interface{}
	mustDecodeJSON(t, resp, &mergeResp)
	mergedPR := mergeResp["pr"].(map[string]interface{})
	assertEqual(t, "MERGED", mergedPR["status"], "Статус должен быть MERGED")

	resp = makeRequest(t, ts, "POST", "/pullRequest/merge", merge, adminToken)
	assertEqual(t, http.StatusOK, resp.StatusCode, "Повторный merge должен быть успешным")
	defer closeResponseBody(t, resp)

	resp = makeRequest(t, ts, "POST", "/pullRequest/reassign", reassign, adminToken)
	assertEqual(t, http.StatusConflict, resp.StatusCode, "Переназначение после merge должно быть запрещено")
	defer closeResponseBody(t, resp)
}

func TestStatistics(t *testing.T) {
	ts := setupTestServer(t)
	if ts == nil {
		return
	}
	defer ts.Close()

	team := map[string]interface{}{
		"team_name": "frontend",
		"members": []map[string]interface{}{
			{"user_id": "dev1", "username": "Dev1", "is_active": true},
			{"user_id": "dev2", "username": "Dev2", "is_active": true},
		},
	}
	resp := makeRequest(t, ts, "POST", "/team/add", team, adminToken)
	defer closeResponseBody(t, resp)

	for i := 1; i <= 3; i++ {
		pr := map[string]interface{}{
			"pull_request_id":   fmt.Sprintf("pr-%d", i),
			"pull_request_name": fmt.Sprintf("Feature %d", i),
			"author_id":         "dev1",
		}
		resp := makeRequest(t, ts, "POST", "/pullRequest/create", pr, adminToken)
		defer closeResponseBody(t, resp)
	}

	merge := map[string]interface{}{"pull_request_id": "pr-1"}
	resp = makeRequest(t, ts, "POST", "/pullRequest/merge", merge, adminToken)
	defer closeResponseBody(t, resp)

	resp = makeRequest(t, ts, "GET", "/stats", nil, userToken)
	assertEqual(t, http.StatusOK, resp.StatusCode, "Получение статистики")
	defer closeResponseBody(t, resp)

	var stats map[string]interface{}
	mustDecodeJSON(t, resp, &stats)

	prStats := stats["pr_stats"].(map[string]interface{})
	assertEqual(t, 3.0, prStats["total_prs"], "Всего PR")
	assertEqual(t, 2.0, prStats["open_prs"], "Открытых PR")
	assertEqual(t, 1.0, prStats["merged_prs"], "Merged PR")
}

func TestDeactivateTeamUsers(t *testing.T) {
	ts := setupTestServer(t)
	if ts == nil {
		return
	}
	defer ts.Close()

	team1 := map[string]interface{}{
		"team_name": "team-a",
		"members": []map[string]interface{}{
			{"user_id": "a1", "username": "A1", "is_active": true},
			{"user_id": "a2", "username": "A2", "is_active": true},
		},
	}
	resp := makeRequest(t, ts, "POST", "/team/add", team1, adminToken)
	defer closeResponseBody(t, resp)

	team2 := map[string]interface{}{
		"team_name": "team-b",
		"members": []map[string]interface{}{
			{"user_id": "b1", "username": "B1", "is_active": true},
			{"user_id": "b2", "username": "B2", "is_active": true},
		},
	}
	resp = makeRequest(t, ts, "POST", "/team/add", team2, adminToken)
	defer closeResponseBody(t, resp)

	pr := map[string]interface{}{
		"pull_request_id":   "pr-test",
		"pull_request_name": "Test PR",
		"author_id":         "b1",
	}
	resp = makeRequest(t, ts, "POST", "/pullRequest/create", pr, adminToken)
	defer closeResponseBody(t, resp)

	deactivate := map[string]interface{}{"team_name": "team-a"}
	resp = makeRequest(t, ts, "POST", "/team/deactivateUsers", deactivate, adminToken)
	assertEqual(t, http.StatusOK, resp.StatusCode, "Деактивация команды")
	defer closeResponseBody(t, resp)

	var result map[string]interface{}
	mustDecodeJSON(t, resp, &result)

	assertEqual(t, 2.0, result["deactivated_count"], "Деактивировано пользователей")

	setActive := map[string]interface{}{"user_id": "a1", "is_active": false}
	resp = makeRequest(t, ts, "POST", "/users/setIsActive", setActive, adminToken)
	assertEqual(t, http.StatusOK, resp.StatusCode, "Пользователь уже неактивен")
	defer closeResponseBody(t, resp)
}

func TestUserActivation(t *testing.T) {
	ts := setupTestServer(t)
	if ts == nil {
		return
	}
	defer ts.Close()

	team := map[string]interface{}{
		"team_name": "test-team",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "User1", "is_active": true},
		},
	}
	resp := makeRequest(t, ts, "POST", "/team/add", team, adminToken)
	defer closeResponseBody(t, resp)

	setActive := map[string]interface{}{"user_id": "u1", "is_active": false}
	resp = makeRequest(t, ts, "POST", "/users/setIsActive", setActive, adminToken)
	assertEqual(t, http.StatusOK, resp.StatusCode, "Деактивация пользователя")
	defer closeResponseBody(t, resp)

	setActive = map[string]interface{}{"user_id": "u1", "is_active": true}
	resp = makeRequest(t, ts, "POST", "/users/setIsActive", setActive, adminToken)
	assertEqual(t, http.StatusOK, resp.StatusCode, "Активация пользователя")
	defer closeResponseBody(t, resp)
}

func makeRequest(t *testing.T, ts *testServer, method, path string, body interface{}, token string) *http.Response {
	t.Helper()

	var reqBody *bytes.Buffer
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Ошибка маршалинга JSON: %v", err)
		}
		reqBody = bytes.NewBuffer(data)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, ts.server.URL+path, reqBody)
	if err != nil {
		t.Fatalf("Ошибка создания запроса: %v", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Ошибка выполнения запроса: %v", err)
	}

	return resp
}

func mustDecodeJSON(t *testing.T, resp *http.Response, v interface{}) {
	t.Helper()
	defer closeResponseBody(t, resp)

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("Ошибка декодирования JSON: %v", err)
	}
}

func closeResponseBody(t *testing.T, resp *http.Response) {
	t.Helper()
	if err := resp.Body.Close(); err != nil {
		t.Logf("Ошибка закрытия response body: %v", err)
	}
}

func assertEqual(t *testing.T, expected, actual interface{}, msg string) {
	t.Helper()
	if expected != actual {
		t.Errorf("%s: ожидалось %v, получено %v", msg, expected, actual)
	}
}
