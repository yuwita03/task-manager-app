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

// helper: bikin user + board, return boardID sama token
func setupUserAndBoardOnly(t *testing.T) (boardID int, token string) {
	db := setupTestDB()
	defer db.Close()
	truncateAll(db)
	router := setupRouter(db)

	// register
	registerReq := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/auth/register",
		strings.NewReader(`{"name":"Yurico","email":"yurico@test.com","password":"password123"}`))
	registerReq.Header.Add("Content-Type", "application/json")
	registerRec := httptest.NewRecorder()
	router.ServeHTTP(registerRec, registerReq)

	body, _ := io.ReadAll(registerRec.Result().Body)
	var registerRes map[string]interface{}
	json.Unmarshal(body, &registerRes)
	token = registerRes["data"].(map[string]interface{})["token"].(string)

	// create board
	boardReq := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/boards",
		strings.NewReader(`{"name":"Test Board"}`))
	boardReq.Header.Add("Content-Type", "application/json")
	boardReq.Header.Add("Authorization", "Bearer "+token)
	boardRec := httptest.NewRecorder()
	router.ServeHTTP(boardRec, boardReq)

	boardBody, _ := io.ReadAll(boardRec.Result().Body)
	var boardRes map[string]interface{}
	json.Unmarshal(boardBody, &boardRes)
	boardID = int(boardRes["data"].(map[string]interface{})["id"].(float64))

	return boardID, token
}

// test: bikin label baru di board, harus sukses
func TestCreateLabelSuccess(t *testing.T) {
	boardID, token := setupUserAndBoardOnly(t)

	db := setupTestDB()
	defer db.Close()
	router := setupRouter(db)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/api/boards/%d/labels", boardID),
		strings.NewReader(`{"name":"Urgent","color":"#FF0000"}`))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, 201, rec.Result().StatusCode)

	body, _ := io.ReadAll(rec.Result().Body)
	var res map[string]interface{}
	json.Unmarshal(body, &res)
	data := res["data"].(map[string]interface{})

	assert.Equal(t, "Urgent", data["name"])
	assert.Equal(t, "#FF0000", data["color"])
}

// test: bikin label dengan warna salah format (bukan 7 karakter hex), harus gagal 400
func TestCreateLabelFailed_InvalidColor(t *testing.T) {
	boardID, token := setupUserAndBoardOnly(t)

	db := setupTestDB()
	defer db.Close()
	router := setupRouter(db)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/api/boards/%d/labels", boardID),
		strings.NewReader(`{"name":"Urgent","color":"red"}`)) // "red" bukan format hex
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, 400, rec.Result().StatusCode)
}

// test: get semua label di board, harus muncul label yang udah dibuat
func TestGetLabelsByBoardSuccess(t *testing.T) {
	boardID, token := setupUserAndBoardOnly(t)

	db := setupTestDB()
	defer db.Close()
	router := setupRouter(db)

	// bikin 1 label dulu
	createReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/api/boards/%d/labels", boardID),
		strings.NewReader(`{"name":"Urgent","color":"#FF0000"}`))
	createReq.Header.Add("Content-Type", "application/json")
	createReq.Header.Add("Authorization", "Bearer "+token)
	router.ServeHTTP(httptest.NewRecorder(), createReq)

	// ambil semua label di board
	getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:8080/api/boards/%d/labels", boardID), nil)
	getReq.Header.Add("Authorization", "Bearer "+token)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)

	assert.Equal(t, 200, getRec.Result().StatusCode)

	body, _ := io.ReadAll(getRec.Result().Body)
	var res map[string]interface{}
	json.Unmarshal(body, &res)
	labels := res["data"].([]interface{})

	assert.Len(t, labels, 1) // harus ada 1 label
}