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

func TestRegisterSuccess(t *testing.T) {
	db := setupTestDB()
	defer db.Close()
	truncateAll(db)
	router := setupRouter(db)

	requestBody := strings.NewReader(`{"name":"Yurico","email":"yurico@test.com","password":"password123"}`)
	request := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/auth/register", requestBody)
	request.Header.Add("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	response := recorder.Result()
	assert.Equal(t, 201, response.StatusCode)

	body, _ := io.ReadAll(response.Body)
	var responseBody map[string]interface{}
	json.Unmarshal(body, &responseBody)
	fmt.Println(responseBody)

	assert.Equal(t, 201, int(responseBody["code"].(float64)))
	assert.Equal(t, "CREATED", responseBody["status"])

	data := responseBody["data"].(map[string]interface{})
	assert.NotEmpty(t, data["token"])
	user := data["user"].(map[string]interface{})
	assert.Equal(t, "Yurico", user["name"])
	assert.Equal(t, "yurico@test.com", user["email"])
}

func TestRegisterFailed_ValidationError(t *testing.T) {
	db := setupTestDB()
	defer db.Close()
	truncateAll(db)
	router := setupRouter(db)

	requestBody := strings.NewReader(`{"name":"","email":"notanemail","password":"123"}`)
	request := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/auth/register", requestBody)
	request.Header.Add("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	response := recorder.Result()
	assert.Equal(t, 400, response.StatusCode)

	body, _ := io.ReadAll(response.Body)
	var responseBody map[string]interface{}
	json.Unmarshal(body, &responseBody)

	assert.Equal(t, "BAD REQUEST", responseBody["status"])
}

func TestRegisterFailed_EmailAlreadyExists(t *testing.T) {
	db := setupTestDB()
	defer db.Close()
	truncateAll(db)
	router := setupRouter(db)

	// register pertama, harus sukses
	firstRequest := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/auth/register",
		strings.NewReader(`{"name":"Yurico","email":"yurico@test.com","password":"password123"}`))
	firstRequest.Header.Add("Content-Type", "application/json")
	firstRecorder := httptest.NewRecorder()
	router.ServeHTTP(firstRecorder, firstRequest)

	// register kedua pakai email sama, harus gagal 409
	secondRequest := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/auth/register",
		strings.NewReader(`{"name":"Yurico2","email":"yurico@test.com","password":"password456"}`))
	secondRequest.Header.Add("Content-Type", "application/json")
	secondRecorder := httptest.NewRecorder()
	router.ServeHTTP(secondRecorder, secondRequest)

	response := secondRecorder.Result()
	assert.Equal(t, 409, response.StatusCode)

	body, _ := io.ReadAll(response.Body)
	var responseBody map[string]interface{}
	json.Unmarshal(body, &responseBody)
	assert.Equal(t, "CONFLICT", responseBody["status"])
}

func TestLoginSuccess(t *testing.T) {
	db := setupTestDB()
	defer db.Close()
	truncateAll(db)
	router := setupRouter(db)

	registerRequest := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/auth/register",
		strings.NewReader(`{"name":"Yurico","email":"yurico@test.com","password":"password123"}`))
	registerRequest.Header.Add("Content-Type", "application/json")
	router.ServeHTTP(httptest.NewRecorder(), registerRequest)

	loginRequest := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/auth/login",
		strings.NewReader(`{"email":"yurico@test.com","password":"password123"}`))
	loginRequest.Header.Add("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, loginRequest)

	response := recorder.Result()
	assert.Equal(t, 200, response.StatusCode)

	body, _ := io.ReadAll(response.Body)
	var responseBody map[string]interface{}
	json.Unmarshal(body, &responseBody)

	data := responseBody["data"].(map[string]interface{})
	assert.NotEmpty(t, data["token"])
}

func TestLoginFailed_WrongPassword(t *testing.T) {
	db := setupTestDB()
	defer db.Close()
	truncateAll(db)
	router := setupRouter(db)

	registerRequest := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/auth/register",
		strings.NewReader(`{"name":"Yurico","email":"yurico@test.com","password":"password123"}`))
	registerRequest.Header.Add("Content-Type", "application/json")
	router.ServeHTTP(httptest.NewRecorder(), registerRequest)

	loginRequest := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/auth/login",
		strings.NewReader(`{"email":"yurico@test.com","password":"wrongpassword"}`))
	loginRequest.Header.Add("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, loginRequest)

	response := recorder.Result()
	assert.Equal(t, 401, response.StatusCode)
}

func TestMeReturnsCurrentUserProfile(t *testing.T) {
	db := setupTestDB()
	defer db.Close()
	truncateAll(db)
	router := setupRouter(db)

	registerRequest := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/auth/register",
		strings.NewReader(`{"name":"Yurico","email":"yurico@test.com","password":"password123"}`))
	registerRequest.Header.Add("Content-Type", "application/json")
	registerRec := httptest.NewRecorder()
	router.ServeHTTP(registerRec, registerRequest)

	body, _ := io.ReadAll(registerRec.Result().Body)
	var registerRes map[string]interface{}
	json.Unmarshal(body, &registerRes)
	token := registerRes["data"].(map[string]interface{})["token"].(string)

	meReq := httptest.NewRequest(http.MethodGet, "http://localhost:8080/api/auth/me", nil)
	meReq.Header.Add("Authorization", "Bearer "+token)
	meRec := httptest.NewRecorder()
	router.ServeHTTP(meRec, meReq)

	assert.Equal(t, 200, meRec.Result().StatusCode)

	meBody, _ := io.ReadAll(meRec.Result().Body)
	var meRes map[string]interface{}
	json.Unmarshal(meBody, &meRes)
	meData := meRes["data"].(map[string]interface{})

	assert.Equal(t, "Yurico", meData["name"])
	assert.Equal(t, "yurico@test.com", meData["email"])
}
