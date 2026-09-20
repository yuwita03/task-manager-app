package test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

// helper: register user + insert board manual, return (userID, boardID, token)
func setupUserAndBoard(db *pgxpool.Pool, router http.Handler) (int, int, string) {
	registerRequest := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/auth/register",
		strings.NewReader(`{"name":"Yurico","email":"yurico@test.com","password":"password123"}`))
	registerRequest.Header.Add("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, registerRequest)

	body, _ := io.ReadAll(recorder.Result().Body)
	var responseBody map[string]interface{}
	json.Unmarshal(body, &responseBody)
	data := responseBody["data"].(map[string]interface{})
	token := data["token"].(string)
	userID := int(data["user"].(map[string]interface{})["id"].(float64))

	var boardID int
	err := db.QueryRow(context.Background(),
		`INSERT INTO boards (name, owner_id) VALUES ($1, $2) RETURNING id`,
		"Test Board", userID,
	).Scan(&boardID)
	if err != nil {
		panic(fmt.Sprintf("gagal insert board: %v", err))
	}

	return userID, boardID, token
}

func TestCreateListSuccess(t *testing.T) {
	db := setupTestDB()
	defer db.Close()
	truncateAll(db)
	router := setupRouter(db)

	_, boardID, token := setupUserAndBoard(db, router)

	requestBody := strings.NewReader(`{"name":"To Do"}`)
	request := httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/api/boards/%d/lists", boardID), requestBody)
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	response := recorder.Result()
	assert.Equal(t, 201, response.StatusCode)

	body, _ := io.ReadAll(response.Body)
	var responseBody map[string]interface{}
	json.Unmarshal(body, &responseBody)

	data := responseBody["data"].(map[string]interface{})
	assert.Equal(t, "To Do", data["name"])
	assert.Equal(t, float64(boardID), data["board_id"])
}

func TestCreateListFailed_NoAuth(t *testing.T) {
	db := setupTestDB()
	defer db.Close()
	truncateAll(db)
	router := setupRouter(db)

	requestBody := strings.NewReader(`{"name":"To Do"}`)
	request := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/boards/1/lists", requestBody)
	request.Header.Add("Content-Type", "application/json")
	// tanpa header Authorization

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	assert.Equal(t, 401, recorder.Result().StatusCode)
}

func TestGetListsByBoardSuccess(t *testing.T) {
	db := setupTestDB()
	defer db.Close()
	truncateAll(db)
	router := setupRouter(db)

	_, boardID, token := setupUserAndBoard(db, router)

	createRequest := httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/api/boards/%d/lists", boardID),
		strings.NewReader(`{"name":"To Do"}`))
	createRequest.Header.Add("Content-Type", "application/json")
	createRequest.Header.Add("Authorization", "Bearer "+token)
	router.ServeHTTP(httptest.NewRecorder(), createRequest)

	getRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:8080/api/boards/%d/lists", boardID), nil)
	getRequest.Header.Add("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, getRequest)

	response := recorder.Result()
	assert.Equal(t, 200, response.StatusCode)

	body, _ := io.ReadAll(response.Body)
	var responseBody map[string]interface{}
	json.Unmarshal(body, &responseBody)

	lists := responseBody["data"].([]interface{})
	assert.Len(t, lists, 1)
}

func TestDeleteListFailed_NotFound(t *testing.T) {
	db := setupTestDB()
	defer db.Close()
	truncateAll(db)
	router := setupRouter(db)

	_, _, token := setupUserAndBoard(db, router)

	request := httptest.NewRequest(http.MethodDelete, "http://localhost:8080/api/lists/9999", nil)
	request.Header.Add("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	assert.Equal(t, 404, recorder.Result().StatusCode)
}