package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUsers(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users", nil)
	router.ServeHTTP(w, req)

	var expectedUsers = []user{
		{Id: "2", Name: "Fernando Migliorini Tenório", Email: "dotenorio@gmail.com"},
	}

	var response []user
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, response[0].Id, expectedUsers[0].Id)
	assert.Equal(t, response[0].Name, expectedUsers[0].Name)
	assert.Equal(t, response[0].Email, expectedUsers[0].Email)
	assert.NotNil(t, response[0].CreatedAt)
}
