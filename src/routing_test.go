package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/h2non/gock"
	"github.com/retailcrm/mg-transport-api-client-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var router *gin.Engine

func init() {
	os.Chdir("../")
	config = LoadConfig("config_test.yml")
	orm = NewDb(config)
	logger = newLogger()
	router = setup()
	c := Connection{
		ID:       1,
		ClientID: "123123",
		APIKEY:   "test",
		APIURL:   "https://test.retailcrm.ru",
		MGURL:    "https://test.retailcrm.pro",
		MGToken:  "test-token",
		Active:   true,
	}

	c.createConnection()
	orm.DB.Delete(Bot{}, "token = ?", "123123:Qwerty")
}

func TestRouting_connectHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code,
		fmt.Sprintf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK))
}

func TestRouting_addBotHandler(t *testing.T) {
	defer gock.Off()

	ch := v1.Channel{
		Type: "telegram",
		Settings: v1.ChannelSettings{
			SpamAllowed: false,
			Status: v1.Status{
				Delivered: v1.ChannelFeatureSend,
				Read:      v1.ChannelFeatureNone,
			},
			Text: v1.ChannelSettingsText{
				Creating: v1.ChannelFeatureBoth,
				Editing:  v1.ChannelFeatureBoth,
				Quoting:  v1.ChannelFeatureBoth,
				Deleting: v1.ChannelFeatureReceive,
			},
			Product: v1.Product{
				Creating: v1.ChannelFeatureReceive,
				Editing:  v1.ChannelFeatureReceive,
			},
			Order: v1.Order{
				Creating: v1.ChannelFeatureReceive,
				Editing:  v1.ChannelFeatureReceive,
			},
		},
	}

	outgoing, _ := json.Marshal(&ch)
	p := url.Values{"url": {"https://" + config.HTTPServer.Host + "/telegram/123123:Qwerty"}}

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
		BodyString(`{"ok":true,"result":{"url":"https://` + config.HTTPServer.Host + `/telegram/123123:Qwerty","has_custom_certificate":false,"pending_update_count":0}}`)

	gock.New("https://test.retailcrm.pro").
		Post("/api/transport/v1/channels").
		JSON([]byte(outgoing)).
		MatchHeader("Content-Type", "application/json").
		MatchHeader("X-Transport-Token", "test-token").
		Reply(201).
		BodyString(`{"id": 1}`)

	req, err := http.NewRequest("POST", "/add-bot/", strings.NewReader(`{"token": "123123:Qwerty", "connectionId": 1}`))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusCreated, rr.Code,
		fmt.Sprintf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusCreated))

	bytes, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		t.Fatal(err)
	}

	var res map[string]interface{}

	err = json.Unmarshal(bytes, &res)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "123123:Qwerty", res["token"])
}

func TestRouting_deleteBotHandler(t *testing.T) {
	defer gock.Off()

	gock.New("https://test.retailcrm.pro").
		Post("/api/transport/v1/channels").
		BodyString(`{"ID":1,"Type":"telegram","Events":["message_sent","message_updated","message_deleted","message_read"]}`).
		MatchHeader("Content-Type", "application/json").
		MatchHeader("X-Transport-Token", "123123").
		Reply(200).
		BodyString(`{"id": 1}`)

	req, err := http.NewRequest("POST", "/delete-bot/", strings.NewReader(`{"token": "123123:Qwerty", "active": false, "connectionId": 1}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code,
		fmt.Sprintf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK))
}

func TestRouting_settingsHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/settings/123123", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code,
		fmt.Sprintf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK))
}

func TestRouting_saveHandler(t *testing.T) {
	defer gock.Off()

	gock.New("https://test.retailcrm.ru").
		Get("/api/credentials").
		Reply(200).
		BodyString(`{"success": true, "credentials": ["/api/integration-modules/{code}", "/api/integration-modules/{code}/edit"]}`)

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
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code,
		fmt.Sprintf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK))
}

func TestRouting_activityHandler(t *testing.T) {
	data := url.Values{}
	data.Set("clientId", "123123")
	data.Set("activity", `{"active": true, "freeze": false}`)

	req, err := http.NewRequest("POST", "/actions/activity", strings.NewReader(data.Encode()))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code,
		fmt.Sprintf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK))
}
