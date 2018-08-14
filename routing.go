package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/retailcrm/api-client-go/v5"
	"github.com/retailcrm/mg-transport-api-client-go/v1"
)

func connectHandler(c *gin.Context) {
	rx := regexp.MustCompile(`/+$`)
	ra := rx.ReplaceAllString(c.Query("account"), ``)
	p := Connection{
		APIURL: ra,
	}

	res := struct {
		Conn   *Connection
		Locale map[string]string
	}{
		&p,
		getLocale(),
	}

	c.HTML(http.StatusOK, "home", &res)
}

func addBotHandler(c *gin.Context) {
	var b Bot

	if err := c.ShouldBindJSON(&b); err != nil {
		c.Error(err)
		return
	}

	if b.Token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": getLocalizedMessage("no_bot_token")})
		return
	}

	cl, err := getBotByToken(b.Token)
	if err != nil {
		c.Error(err)
		return
	}

	if cl.ID != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": getLocalizedMessage("bot_already_created")})
		return
	}

	bot, err := tgbotapi.NewBotAPI(b.Token)
	if err != nil {
		logger.Error(b.Token, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": getLocalizedMessage("incorrect_token")})
		return
	}

	bot.Debug = config.Debug

	wr, err := bot.SetWebhook(tgbotapi.NewWebhook("https://" + config.HTTPServer.Host + "/telegram/" + bot.Token))
	if err != nil {
		logger.Error(b.Token, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": getLocalizedMessage("error_creating_webhook")})
		return
	}

	if !wr.Ok {
		logger.Error(b.Token, wr.ErrorCode, wr.Result)
		c.JSON(http.StatusBadRequest, gin.H{"error": getLocalizedMessage("error_creating_webhook")})
		return
	}

	b.Name = bot.Self.FirstName

	ch := v1.Channel{
		Type: "telegram",
		Events: []string{
			"message_sent",
			"message_updated",
			"message_deleted",
			"message_read",
		},
	}

	conn := getConnectionById(b.ConnectionID)

	var client = v1.New(conn.MGURL, conn.MGToken)
	data, status, err := client.ActivateTransportChannel(ch)
	if status != http.StatusCreated {
		logger.Error(conn.APIURL, status, err.Error(), data)
		c.JSON(http.StatusBadRequest, gin.H{"error": getLocalizedMessage("error_activating_channel")})
		return
	}

	b.Channel = data.ChannelID

	err = conn.createBot(b)
	if err != nil {
		c.Error(err)
		return
	}

	jsonString, err := json.Marshal(b)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, jsonString)
}

func deleteBotHandler(c *gin.Context) {
	var b Bot

	if err := c.ShouldBindJSON(&b); err != nil {
		c.Error(err)
		return
	}

	conn := getConnectionById(b.ConnectionID)
	if conn.MGURL == "" || conn.MGToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": getLocalizedMessage("not_found_account")})
		logger.Error(b.ID, "MGURL or MGToken is empty")
		return
	}

	var client = v1.New(conn.MGURL, conn.MGToken)

	data, status, err := client.DeactivateTransportChannel(getBotChannelByToken(b.Token))
	if status > http.StatusOK {
		c.JSON(http.StatusBadRequest, gin.H{"error": getLocalizedMessage("error_deactivating_channel")})
		logger.Error(b.ID, status, err.Error(), data)
		return
	}

	err = b.deleteBot()
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func settingsHandler(c *gin.Context) {
	uid := c.Param("uid")

	p := getConnection(uid)
	if p.ID == 0 {
		c.Redirect(http.StatusFound, "/")
		return
	}

	bots := p.getBotsByClientID()

	res := struct {
		Conn   *Connection
		Bots   Bots
		Locale map[string]string
	}{
		p,
		bots,
		getLocale(),
	}

	c.HTML(http.StatusOK, "form", &res)
}

func saveHandler(c *gin.Context) {
	var conn Connection

	if err := c.BindJSON(&conn); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": getLocalizedMessage("incorrect_url_key")})
		return
	}

	_, err, code := getAPIClient(conn.APIURL, conn.APIKEY)
	if err != nil {
		c.AbortWithStatusJSON(code, gin.H{"error": err.Error()})
		return
	}

	err = conn.saveConnection()
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": getLocalizedMessage("successful")})
}

func createHandler(c *gin.Context) {
	var conn Connection

	if err := c.BindJSON(&conn); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": getLocalizedMessage("incorrect_url_key")})
		return
	}

	conn.ClientID = GenerateToken()

	cl := getConnectionByURL(conn.APIURL)
	if cl.ID != 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": getLocalizedMessage("connection_already_created")})
		return
	}

	client, err, _ := getAPIClient(conn.APIURL, conn.APIKEY)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	integration := v5.IntegrationModule{
		Code:            transport,
		IntegrationCode: transport,
		Active:          true,
		Name:            "Telegram",
		ClientID:        conn.ClientID,
		Logo: fmt.Sprintf(
			"https://%s/static/telegram_logo.svg",
			config.HTTPServer.Host,
		),
		BaseURL: fmt.Sprintf(
			"https://%s",
			config.HTTPServer.Host,
		),
		AccountURL: fmt.Sprintf(
			"https://%s/settings/%s",
			config.HTTPServer.Host,
			conn.ClientID,
		),
		Actions: map[string]string{"activity": "/actions/activity"},
		Integrations: &v5.Integrations{
			MgTransport: &v5.MgTransport{
				WebhookUrl: fmt.Sprintf(
					"https://%s/webhook/",
					config.HTTPServer.Host,
				),
			},
		},
	}

	data, status, errr := client.IntegrationModuleEdit(integration)
	if errr.RuntimeErr != nil {
		c.Error(errr.RuntimeErr)
		return
	}

	if status >= http.StatusBadRequest {
		c.JSON(http.StatusBadRequest, gin.H{"error": getLocalizedMessage("error_activity_mg")})
		logger.Error(conn.APIURL, status, errr.ApiErr, data)
		return
	}

	conn.MGURL = data.Info["baseUrl"]
	conn.MGToken = data.Info["token"]
	conn.Active = true

	err = conn.createConnection()
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(
		http.StatusCreated,
		gin.H{
			"url":     "/settings/" + conn.ClientID,
			"message": getLocalizedMessage("successful"),
		},
	)
}

func activityHandler(c *gin.Context) {
	var rec v5.ActivityCallback

	if err := c.ShouldBindJSON(&rec); err != nil {
		c.Error(err)
		return
	}

	conn := getConnection(rec.ClientId)
	if conn.ID == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{
				"success": false,
				"error":   "Wrong data",
			},
		)
		return
	}

	conn.Active = rec.Activity.Active && !rec.Activity.Freeze

	if err := conn.setConnectionActivity(); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
