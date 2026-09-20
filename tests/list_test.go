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

// setupUserAndBoard bikin 1 user baru dan 1 board baru LEWAT API asli
// (bukan insert langsung ke database). Ini penting karena endpoint create board
// otomatis nyatetin user itu sebagai "owner" di tabel board_members —
// kalau board dibuat manual pakai SQL, baris di board_members ini gak akan ada,
// terus semua request berikutnya bakal ditolak 401 "not a member of this board".
//
// Return: userID (ID user yang baru dibuat), boardID (ID board yang baru dibuat), token (JWT buat auth)
func setupUserAndBoard(router http.Handler) (int, int, string) {
	// step 1: register user baru
	registerRequest := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/auth/register",
		strings.NewReader(`{"name":"Yurico","email":"yurico@test.com","password":"password123"}`))
	registerRequest.Header.Add("Content-Type", "application/json")
	registerRecorder := httptest.NewRecorder()
	router.ServeHTTP(registerRecorder, registerRequest)

	// ambil token dan userID dari response register
	registerBody, _ := io.ReadAll(registerRecorder.Result().Body)
	var registerResponse map[string]interface{}
	json.Unmarshal(registerBody, &registerResponse)
	data := registerResponse["data"].(map[string]interface{})
	token := data["token"].(string)
	userID := int(data["user"].(map[string]interface{})["id"].(float64))

	// step 2: bikin board pakai token yang baru didapat
	// (endpoint ini otomatis bikin row di board_members dengan role "owner")
	boardRequest := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/boards",
		strings.NewReader(`{"name":"Test Board"}`))
	boardRequest.Header.Add("Content-Type", "application/json")
	boardRequest.Header.Add("Authorization", "Bearer "+token)
	boardRecorder := httptest.NewRecorder()
	router.ServeHTTP(boardRecorder, boardRequest)

	// ambil boardID dari response create board
	boardBody, _ := io.ReadAll(boardRecorder.Result().Body)
	var boardResponse map[string]interface{}
	json.Unmarshal(boardBody, &boardResponse)
	boardID := int(boardResponse["data"].(map[string]interface{})["id"].(float64))

	return userID, boardID, token
}

// Test: bikin list baru di board yang valid, dengan token yang valid → harus sukses (201)
func TestCreateListSuccess(t *testing.T) {
	db := setupTestDB()      // konek ke database test (task_manager_test)
	defer db.Close()         // tutup koneksi setelah test selesai
	truncateAll(db)          // bersihin semua tabel biar test mulai dari kondisi kosong
	router := setupRouter(db) // bikin instance router (Gin) yang lengkap sama semua handler

	_, boardID, token := setupUserAndBoard(router) // siapin user + board dulu

	// kirim request bikin list baru
	requestBody := strings.NewReader(`{"name":"To Do"}`)
	request := httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/api/boards/%d/lists", boardID), requestBody)
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request) // eksekusi request tanpa perlu server beneran jalan

	// cek response-nya
	response := recorder.Result()
	assert.Equal(t, 201, response.StatusCode) // harus 201 Created

	body, _ := io.ReadAll(response.Body)
	var responseBody map[string]interface{}
	json.Unmarshal(body, &responseBody)

	data := responseBody["data"].(map[string]interface{})
	assert.Equal(t, "To Do", data["name"])              // nama list harus sesuai yang dikirim
	assert.Equal(t, float64(boardID), data["board_id"]) // list harus nempel ke board yang benar
}

// Test: bikin list TANPA header Authorization → harus ditolak (401)
func TestCreateListFailed_NoAuth(t *testing.T) {
	db := setupTestDB()
	defer db.Close()
	truncateAll(db)
	router := setupRouter(db)

	requestBody := strings.NewReader(`{"name":"To Do"}`)
	request := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/boards/1/lists", requestBody)
	request.Header.Add("Content-Type", "application/json")
	// sengaja TIDAK nambahin header Authorization

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	assert.Equal(t, 401, recorder.Result().StatusCode) // harus ditolak karena gak ada token
}

// Test: ambil semua list dalam 1 board → harus balikin list yang udah dibuat sebelumnya
func TestGetListsByBoardSuccess(t *testing.T) {
	db := setupTestDB()
	defer db.Close()
	truncateAll(db)
	router := setupRouter(db)

	_, boardID, token := setupUserAndBoard(router)

	// bikin 1 list dulu, biar ada data buat di-fetch
	createRequest := httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/api/boards/%d/lists", boardID),
		strings.NewReader(`{"name":"To Do"}`))
	createRequest.Header.Add("Content-Type", "application/json")
	createRequest.Header.Add("Authorization", "Bearer "+token)
	router.ServeHTTP(httptest.NewRecorder(), createRequest) // recorder-nya gak dipakai, cuma buat side-effect

	// sekarang fetch semua list di board ini
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
	assert.Len(t, lists, 1) // harus ada tepat 1 list (yang barusan dibuat)
}

// Test: hapus list dengan ID yang gak ada di database → harus 404 Not Found
func TestDeleteListFailed_NotFound(t *testing.T) {
	db := setupTestDB()
	defer db.Close()
	truncateAll(db)
	router := setupRouter(db)

	_, _, token := setupUserAndBoard(router) // butuh token valid, walau list-nya gak ada

	request := httptest.NewRequest(http.MethodDelete, "http://localhost:8080/api/lists/9999", nil) // ID 9999 sengaja gak ada
	request.Header.Add("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	assert.Equal(t, 404, recorder.Result().StatusCode)
}