package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	loadEnv()

	dbName := os.Getenv("DATABASE_NAME")
	os.Setenv("DATABASE_NAME", fmt.Sprintf("%s-test", dbName))

	setupDb()

	code := m.Run()
	os.Exit(code)
}

func TestGetUsers(t *testing.T) {
	router := setupRouter()

	var expectedUsers = []User{
		{Name: "Foo Bar", Email: "foo@bar.com"},
	}

	db.Create(&expectedUsers[0])

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users", nil)
	router.ServeHTTP(w, req)

	var response []User
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.Nil(t, err)
	assert.Equal(t, w.Code, http.StatusOK)
	assert.Equal(t, response[0].Id, expectedUsers[0].Id)
	assert.Equal(t, response[0].Name, expectedUsers[0].Name)
	assert.Equal(t, response[0].Email, expectedUsers[0].Email)
	assert.NotNil(t, response[0].CreatedAt)

	db.Unscoped().Delete(&expectedUsers[0])
}

func TestGetUserById(t *testing.T) {
	router := setupRouter()

	var expectedUser = User{Name: "Foo Bar", Email: "foo@bar.com"}

	db.Create(&expectedUser)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/users/%s", expectedUser.Id), nil)
	router.ServeHTTP(w, req)

	var response User
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.Nil(t, err)
	assert.Equal(t, w.Code, http.StatusOK)
	assert.Equal(t, response.Id, expectedUser.Id)
	assert.Equal(t, response.Name, expectedUser.Name)
	assert.Equal(t, response.Email, expectedUser.Email)
	assert.NotNil(t, response.CreatedAt)

	db.Unscoped().Delete(&expectedUser)
}

func TestCreateUser(t *testing.T) {
	router := setupRouter()

	var expectedUser = User{Name: "Foo Bar", Email: "foo@bar.com"}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(
		"POST", "/users", strings.NewReader(fmt.Sprintf(
			`{"name": "%s", "email": "%s"}`,
			expectedUser.Name, expectedUser.Email,
		),
		),
	)
	router.ServeHTTP(w, req)

	var response User
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.Nil(t, err)
	assert.Equal(t, w.Code, http.StatusCreated)
	assert.NotNil(t, response.Id)
	assert.Equal(t, response.Name, expectedUser.Name)
	assert.Equal(t, response.Email, expectedUser.Email)
	assert.NotNil(t, response.CreatedAt)
	assert.Nil(t, response.UpdatedAt)
	assert.Nil(t, response.DeletedAt)

	db.Unscoped().Delete(&User{Id: response.Id})
}

func TestCreateUserIncorrectBody(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/users", nil)
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.Nil(t, err)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.NotNil(t, response["message"])
	assert.Equal(t, response["message"], "invalid request body")
}

func TestCreateUserAlreadyExists(t *testing.T) {
	router := setupRouter()

	var expectedUser = User{Name: "Foo Bar", Email: "foo@bar.com"}

	db.Create(&expectedUser)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(
		"POST", "/users", strings.NewReader(fmt.Sprintf(
			`{"name": "%s", "email": "%s"}`,
			expectedUser.Name, expectedUser.Email,
		),
		),
	)
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.Nil(t, err)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.NotNil(t, response["message"])
	assert.Equal(t, response["message"], "user already exists")

	db.Unscoped().Delete(&expectedUser)
}

func TestEditUser(t *testing.T) {
	router := setupRouter()

	var expectedUser = User{Name: "Foo Bar", Email: "foo@bar.com"}

	db.Create(&expectedUser)

	newName := fmt.Sprintf("%s (edited)", expectedUser.Name)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(
		"PATCH", fmt.Sprintf("/users/%s", expectedUser.Id), strings.NewReader(fmt.Sprintf(
			`{"name": "%s"}`,
			newName,
		),
		),
	)
	router.ServeHTTP(w, req)

	var response User
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.Nil(t, err)
	assert.Equal(t, w.Code, http.StatusOK)
	assert.NotNil(t, response.Id)
	assert.Equal(t, response.Name, newName)
	assert.Equal(t, response.Email, expectedUser.Email)
	assert.NotNil(t, response.CreatedAt)
	assert.NotNil(t, response.UpdatedAt)
	assert.Nil(t, response.DeletedAt)

	db.Unscoped().Delete(&expectedUser)
}

func TestEditUserIncorrectBody(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/users/38720f78-9268-445e-8cbb-25be090b8ab4", nil)
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.Nil(t, err)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.NotNil(t, response["message"])
	assert.Equal(t, response["message"], "invalid request body")
}

func TestEditUserNotFound(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/users/38720f78-9268-445e-8cbb-25be090b8ab4", strings.NewReader(
		`{"name": "not found"}`,
	),
	)
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.Nil(t, err)
	assert.Equal(t, w.Code, http.StatusNotFound)
	assert.NotNil(t, response["message"])
	assert.Equal(t, response["message"], "user not found")
}

func TestDeleteUser(t *testing.T) {
	router := setupRouter()

	var expectedUser = User{Name: "Foo Bar", Email: "foo@bar.com"}

	db.Create(&expectedUser)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/users/%s", expectedUser.Id), nil)
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.Nil(t, err)
	assert.Equal(t, w.Code, http.StatusOK)
	assert.NotNil(t, response["message"])
	assert.Equal(t, response["message"], "user deleted")

	db.Unscoped().Delete(&expectedUser)
}

func TestDeleteUserNotFound(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/users/38720f78-9268-445e-8cbb-25be090b8ab4", nil)
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.Nil(t, err)
	assert.Equal(t, w.Code, http.StatusNotFound)
	assert.NotNil(t, response["message"])
	assert.Equal(t, response["message"], "user not found")
}
