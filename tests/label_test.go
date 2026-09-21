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

func setupUserAndBoardOnly(router http.Handler) (int, string) {
	registerReq := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/auth/register",
		strings.NewReader(`{"name":"Yurico","email":"yurico@test.com","password":"password123"}`))
	registerReq.Header.Add("Content-Type", "application/json")
	registerRec := httptest.NewRecorder()
	router.ServeHTTP(registerRec, registerReq)

	body, _ := io.ReadAll(registerRec.Result().Body)
	var registerRes map[string]interface{}
	json.Unmarshal(body, &registerRes)
	token := registerRes["data"].(map[string]interface{})["token"].(string)

	boardReq := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/boards",
		strings.NewReader(`{"name":"Test Board"}`))
	boardReq.Header.Add("Content-Type", "application/json")
	boardReq.Header.Add("Authorization", "Bearer "+token)
	boardRec := httptest.NewRecorder()
	router.ServeHTTP(boardRec, boardReq)

	boardBody, _ := io.ReadAll(boardRec.Result().Body)
	var boardRes map[string]interface{}
	json.Unmarshal(boardBody, &boardRes)
	boardID := int(boardRes["data"].(map[string]interface{})["id"].(float64))

	return boardID, token
}

func TestCreateLabelSuccess(t *testing.T) {
	db := setupTestDB()      // GANTI: db+router dibuat DULU
	defer db.Close()
	truncateAll(db)          // TAMBAH: ini juga sempat kelewat di semua test file ini
	router := setupRouter(db)

	boardID, token := setupUserAndBoardOnly(router) // GANTI: kirim router, bukan t

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

func TestCreateLabelFailed_InvalidColor(t *testing.T) {
	db := setupTestDB()
	defer db.Close()
	truncateAll(db)
	router := setupRouter(db)

	boardID, token := setupUserAndBoardOnly(router)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/api/boards/%d/labels", boardID),
		strings.NewReader(`{"name":"Urgent","color":"red"}`))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, 400, rec.Result().StatusCode)
}

func TestGetLabelsByBoardSuccess(t *testing.T) {
	db := setupTestDB()
	defer db.Close()
	truncateAll(db)
	router := setupRouter(db)

	boardID, token := setupUserAndBoardOnly(router) // GANTI: kirim router, bukan t

	createReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/api/boards/%d/labels", boardID),
		strings.NewReader(`{"name":"Urgent","color":"#FF0000"}`))
	createReq.Header.Add("Content-Type", "application/json")
	createReq.Header.Add("Authorization", "Bearer "+token)
	router.ServeHTTP(httptest.NewRecorder(), createReq)

	getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:8080/api/boards/%d/labels", boardID), nil)
	getReq.Header.Add("Authorization", "Bearer "+token)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)

	assert.Equal(t, 200, getRec.Result().StatusCode)

	body, _ := io.ReadAll(getRec.Result().Body)
	var res map[string]interface{}
	json.Unmarshal(body, &res)
	labels := res["data"].([]interface{})

	assert.Len(t, labels, 1)
}