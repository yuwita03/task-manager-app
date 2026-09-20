// tests/task_test.go
package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// helper: bikin user + board + list, terus return listID sama token
// dipakai berulang di semua test task biar gak nulis ulang tiap kali
func setupUserBoardAndList(t *testing.T) (listID int, token string) {
	db := setupTestDB()
	defer db.Close()
	truncateAll(db) // bersihin data lama biar test gak numpuk
	router := setupRouter(db)

	// step 1: register user
	registerReq := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/auth/register",
		strings.NewReader(`{"name":"Yurico","email":"yurico@test.com","password":"password123"}`))
	registerReq.Header.Add("Content-Type", "application/json")
	registerRec := httptest.NewRecorder()
	router.ServeHTTP(registerRec, registerReq)

	body, _ := io.ReadAll(registerRec.Result().Body)
	var registerRes map[string]interface{}
	json.Unmarshal(body, &registerRes)
	data := registerRes["data"].(map[string]interface{})
	token = data["token"].(string)

	// step 2: create board pakai token tadi
	boardReq := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/boards",
		strings.NewReader(`{"name":"Test Board"}`))
	boardReq.Header.Add("Content-Type", "application/json")
	boardReq.Header.Add("Authorization", "Bearer "+token)
	boardRec := httptest.NewRecorder()
	router.ServeHTTP(boardRec, boardReq)

	boardBody, _ := io.ReadAll(boardRec.Result().Body)
	var boardRes map[string]interface{}
	json.Unmarshal(boardBody, &boardRes)
	boardData := boardRes["data"].(map[string]interface{})
	boardID := int(boardData["id"].(float64))

	// step 3: create list di dalam board tadi
	listReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/api/boards/%d/lists", boardID),
		strings.NewReader(`{"name":"To Do"}`))
	listReq.Header.Add("Content-Type", "application/json")
	listReq.Header.Add("Authorization", "Bearer "+token)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)

	listBody, _ := io.ReadAll(listRec.Result().Body)
	var listRes map[string]interface{}
	json.Unmarshal(listBody, &listRes)
	listData := listRes["data"].(map[string]interface{})
	listID = int(listData["id"].(float64))

	return listID, token
}

// test: bikin task baru, harus sukses dengan status 201
func TestCreateTaskSuccess(t *testing.T) {
	listID, token := setupUserBoardAndList(t)

	db := setupTestDB()
	defer db.Close()
	router := setupRouter(db)

	// kirim request create task
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/api/lists/%d/tasks", listID),
		strings.NewReader(`{"title":"Belajar Golang"}`))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// cek response
	assert.Equal(t, 201, rec.Result().StatusCode)

	body, _ := io.ReadAll(rec.Result().Body)
	var res map[string]interface{}
	json.Unmarshal(body, &res)
	data := res["data"].(map[string]interface{})

	// pastikan data task sesuai yang dikirim
	assert.Equal(t, "Belajar Golang", data["title"])
	assert.Equal(t, "todo", data["status"]) // default status harus "todo"
}

// test: create task tanpa title, harus gagal validasi (400)
func TestCreateTaskFailed_EmptyTitle(t *testing.T) {
	listID, token := setupUserBoardAndList(t)

	db := setupTestDB()
	defer db.Close()
	router := setupRouter(db)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/api/lists/%d/tasks", listID),
		strings.NewReader(`{"title":""}`)) // title kosong
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, 400, rec.Result().StatusCode)
}

// test: update task, ganti title dan status jadi "done"
func TestUpdateTaskSuccess(t *testing.T) {
	listID, token := setupUserBoardAndList(t)

	db := setupTestDB()
	defer db.Close()
	router := setupRouter(db)

	// bikin task dulu
	createReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/api/lists/%d/tasks", listID),
		strings.NewReader(`{"title":"Belajar Golang"}`))
	createReq.Header.Add("Content-Type", "application/json")
	createReq.Header.Add("Authorization", "Bearer "+token)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)

	createBody, _ := io.ReadAll(createRec.Result().Body)
	var createRes map[string]interface{}
	json.Unmarshal(createBody, &createRes)
	taskID := int(createRes["data"].(map[string]interface{})["id"].(float64))

	// update task-nya
	updateReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("http://localhost:8080/api/tasks/%d", taskID),
		strings.NewReader(`{"title":"Belajar Golang Lanjutan","description":"","status":"done"}`))
	updateReq.Header.Add("Content-Type", "application/json")
	updateReq.Header.Add("Authorization", "Bearer "+token)
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)

	assert.Equal(t, 200, updateRec.Result().StatusCode)

	updateBody, _ := io.ReadAll(updateRec.Result().Body)
	var updateRes map[string]interface{}
	json.Unmarshal(updateBody, &updateRes)
	data := updateRes["data"].(map[string]interface{})

	assert.Equal(t, "Belajar Golang Lanjutan", data["title"])
	assert.Equal(t, "done", data["status"])
}

// test: hapus task yang gak ada, harus 404
func TestDeleteTaskFailed_NotFound(t *testing.T) {
	_, token := setupUserBoardAndList(t)

	db := setupTestDB()
	defer db.Close()
	router := setupRouter(db)

	req := httptest.NewRequest(http.MethodDelete, "http://localhost:8080/api/tasks/9999", nil)
	req.Header.Add("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, 404, rec.Result().StatusCode)
}