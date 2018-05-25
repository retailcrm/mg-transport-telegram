package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/h2non/gock"
)

func init() {
	c := Connection{
		ClientID: "123123",
		APIKEY:   "test",
		APIURL:   "https://test.retailcrm.ru",
		MGURL:    "https://test.retailcrm.pro",
		MGToken:  "test-token",
		Active:   true,
	}

	c.createConnection()
}
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

	p := url.Values{"url": {"https://test.com/telegram/123123:Qwerty"}}

	gock.New("https://api.telegram.org").
		Post("/bot123123:Qwerty/getMe").
		Reply(200).
		BodyString(`{"ok":true,"result":{"id":123,"is_bot":true,"first_name":"Test","username":"TestBot"}}`)

	gock.New("https://api.telegram.org").
		Post("/bot123123:Qwerty/setWebhook").
		MatchType("url").
		BodyString(p.Encode()).
		Reply(201).
		BodyString(`{"ok":true}`)

	gock.New("https://api.telegram.org").
		Post("/bot123123:Qwerty/getWebhookInfo").
		Reply(200).
		BodyString(`{"ok":true,"result":{"url":"https://test.com/telegram/123123:Qwerty","has_custom_certificate":false,"pending_update_count":0}}`)

	gock.New("https://test.retailcrm.pro").
		Post("/api/v1/transport/channels").
		BodyString(`{"ID":0,"Type":"telegram","Events":["message_sent","message_updated","message_deleted","message_read"]}`).
		MatchHeader("Content-Type", "application/json").
		MatchHeader("X-Transport-Token", "test-token").
		Reply(201).
		BodyString(`{"id": 1}`)

	req, err := http.NewRequest("POST", "/add-bot/", strings.NewReader(`{"token": "123123:Qwerty", "clientId": "123123"}`))
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

func TestRouting_activityBotHandler(t *testing.T) {
	defer gock.Off()

	gock.New("https://test.retailcrm.pro").
		Post("/api/v1/transport/channels").
		BodyString(`{"ID":1,"Type":"telegram","Events":["message_sent","message_updated","message_deleted","message_read"]}`).
		MatchHeader("Content-Type", "application/json").
		MatchHeader("X-Transport-Token", "123123").
		Reply(200).
		BodyString(`{"id": 1}`)

	req, err := http.NewRequest("POST", "/activity-bot/", strings.NewReader(`{"token": "123123:Qwerty", "active": false, "clientId": "123123"}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(activityBotHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusOK)
	}
}

func TestRouting_settingsHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/settings/123123", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(makeHandler(settingsHandler))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusOK)
	}
}

func TestRouting_saveHandler(t *testing.T) {
	req, err := http.NewRequest("POST", "/save/",
		strings.NewReader(
			`{"clientId": "123123", 
			"api_url": "https://test.retailcrm.ru",
			"api_key": "test"}`,
		))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(saveHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusOK)
	}
}

//
//func TestRouting_createHandler(t *testing.T) {
//	defer gock.Off()
//
//	gock.New("https://test2.retailcrm.ru").
//		Get("/api/credentials").
//		Reply(200).
//		BodyString(`{"success": true}`)
//
//	integrationModule := v5.IntegrationModule{
//		Code:            transport,
//		IntegrationCode: transport,
//		Active:          true,
//		Name:            "Telegram",
//		ClientID:        "123",
//		Logo: fmt.Sprintf(
//			"https://%s/web/telegram_logo.svg",
//			config.HTTPServer.Host,
//		),
//		BaseURL: fmt.Sprintf(
//			"https://%s",
//			config.HTTPServer.Host,
//		),
//		AccountURL: fmt.Sprintf(
//			"https://%s/settings/%s",
//			config.HTTPServer.Host,
//			"123",
//		),
//		Actions: map[string]string{"activity": "/actions/activity"},
//		Integrations: &v5.Integrations{
//			MgTransport: &v5.MgTransport{
//				WebhookUrl: fmt.Sprintf(
//					"https://%s/webhook",
//					config.HTTPServer.Host,
//				),
//			},
//		},
//	}
//
//	updateJSON, _ := json.Marshal(&integrationModule)
//	p := url.Values{"integrationModule": {string(updateJSON[:])}}
//
//	gock.New("https://test2.retailcrm.ru").
//		Post(fmt.Sprintf("/api/v5/integration-modules/%s/edit", integrationModule.Code)).
//		MatchType("url").
//		BodyString(p.Encode()).
//		MatchHeader("X-API-KEY", "test").
//		Reply(201).
//		BodyString(`{"success": true, "info": {"baseUrl": "http://test.te", "token": "test"}}`)
//
//	req, err := http.NewRequest("POST", "/create/",
//		strings.NewReader(
//			`{"api_url": "https://test2.retailcrm.ru","api_key": "test"}`,
//		))
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	rr := httptest.NewRecorder()
//	handler := http.HandlerFunc(createHandler)
//	handler.ServeHTTP(rr, req)
//
//	if rr.Code != http.StatusFound {
//		t.Errorf("handler returned wrong status code: got %v want %v",
//			rr.Code, http.StatusFound)
//	}
//}

func TestRouting_activityHandler(t *testing.T) {
	req, err := http.NewRequest("POST", "/actions/activity",
		strings.NewReader(
			`{"clientId": "123123","activity": {"active": true}}`,
		))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(activityHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusOK)
	}
}
