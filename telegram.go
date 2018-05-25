package main

import (
	"net/http"

	"encoding/json"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/getsentry/raven-go"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/retailcrm/mg-transport-api-client-go/v1"
)

func setTransportRoutes() {
	http.HandleFunc("/telegram/", makeHandler(telegramWebhookHandler))
}

// GetBotInfo function
func GetBotInfo(token string) (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return bot, nil
}

// GetBotName function
func GetBotName(bot *tgbotapi.BotAPI) string {
	return bot.Self.FirstName
}

func telegramWebhookHandler(w http.ResponseWriter, r *http.Request, token string) {
	b := getBotByToken(token)
	if b.ID == 0 {
		logger.Error(token, "missing")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !b.Active {
		logger.Error(token, "deactivated")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	c := getConnection(b.ClientID)
	if c.MGURL == "" || c.MGToken == "" {
		logger.Error(token, "MGURL or MGToken is empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var update tgbotapi.Update

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error(token, err)
		return
	}

	err = json.Unmarshal(bytes, &update)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		logger.Error(token, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var client = v1.New(c.MGURL, c.MGToken)

	if update.Message != nil {
		snd := v1.SendData{
			Message: v1.SendMessage{
				Message: v1.Message{
					ExternalID: strconv.Itoa(update.Message.MessageID),
					Type:       "text",
					Text:       update.Message.Text,
				},
				SentAt: time.Now(),
			},
			User: v1.User{
				ExternalID: strconv.Itoa(update.Message.From.ID),
				Nickname:   update.Message.From.UserName,
				Firstname:  update.Message.From.FirstName,
				Lastname:   update.Message.From.LastName,
				Language:   update.Message.From.LanguageCode,
			},
			Channel: b.Channel,
		}

		data, st, err := client.Messages(snd)
		if err != nil {
			logger.Error(token, err.Error(), st, data)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	if update.EditedMessage != nil {
		snd := v1.UpdateData{
			Message: v1.UpdateMessage{
				Message: v1.Message{
					ExternalID: strconv.Itoa(update.EditedMessage.MessageID),
					Type:       "text",
					Text:       update.EditedMessage.Text,
				},
			},
			Channel: b.Channel,
		}

		data, st, err := client.UpdateMessages(snd)
		if err != nil {
			logger.Error(token, err.Error(), st, data)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
