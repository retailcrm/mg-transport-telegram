package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"strings"

	"encoding/json"

	"github.com/h2non/gock"
	"github.com/retailcrm/mg-transport-api-client-go/v1"
)

func TestRouting_connectHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(connectHandler)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusOK)
	}
}

func TestRouting_addBotHandler(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.telegram.org").
		Post("/botbot123:test/getMe").
		Reply(200).
		BodyString(`{"ok":true,"status":200,"name":"TestBot"}`)

	ch := v1.Channel{
		Type: "telegram",
		Events: []string{
			"message_sent",
			"message_read",
		},
	}
	str, _ := json.Marshal(ch)

	gock.New("https://mg-test.com").
		Post("/api/v1/transport/channels").
		JSON(str).
		MatchHeader("Content-Type", "application/json").
		MatchHeader("X-Transport-Token", "test-token").
		Reply(200).
		BodyString(`{"id": 1}`)

	req, err := http.NewRequest("POST", "/add-bot/", strings.NewReader(`{"token": "bot123:test", "clientId": "test"}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(addBotHandler)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusCreated)
	}
}
