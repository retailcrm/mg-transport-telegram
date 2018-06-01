package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/getsentry/raven-go"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/retailcrm/mg-transport-api-client-go/v1"
)

func setTransportRoutes() {
	http.HandleFunc("/telegram/", makeHandler(telegramWebhookHandler))
	http.HandleFunc("/webhook/", mgWebhookHandler)
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
	if config.Debug {
		var v interface{}
		json.NewDecoder(r.Body).Decode(&v)
		logger.Debugf("token: %v mgWebhookHandler: %v", token, v)
	}
	b, err := getBotByToken(token)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		logger.Error(token, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

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

	c := getConnectionById(b.ConnectionID)
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
			raven.CaptureErrorAndWait(err, nil)
			logger.Error(token, err.Error(), st, data)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if config.Debug {
			logger.Debugf("Bot: %v, Message: %v, Response: %v", b.ID, snd, data)
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
			raven.CaptureErrorAndWait(err, nil)
			logger.Error(token, err.Error(), st, data)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if config.Debug {
			logger.Debugf("Bot: %v, Message: %v, Response: %v", b.ID, snd, data)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func mgWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if config.Debug {
		var v interface{}
		json.NewDecoder(r.Body).Decode(&v)
		logger.Debugf("mgWebhookHandler: %v", v)
	}
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		logger.Error(err)
		return
	}

	var msg v1.WebhookRequest
	err = json.Unmarshal(bytes, &msg)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		logger.Error(err)
		return
	}

	uid, _ := strconv.Atoi(msg.Data.ExternalMessageID)

	b := getBotByChannel(msg.Data.ChannelID)
	if b.ID == 0 {
		logger.Error(msg.Data.ChannelID, "missing")
		return
	}

	if !b.Active {
		logger.Error(msg.Data.ChannelID, "deactivated")
		return
	}

	bot, err := tgbotapi.NewBotAPI(b.Token)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		logger.Error(err)
		return
	}

	if msg.Type == "message_sent" {
		msg, err := bot.Send(tgbotapi.NewMessage(msg.Data.ExternalChatID, msg.Data.Content))
		if err != nil {
			raven.CaptureErrorAndWait(err, nil)
			logger.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if config.Debug {
			logger.Debugf("%v", msg)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Message sent"))
	}

	if msg.Type == "message_updated" {
		msg, err := bot.Send(tgbotapi.NewEditMessageText(msg.Data.ExternalChatID, uid, msg.Data.Content))
		if err != nil {
			raven.CaptureErrorAndWait(err, nil)
			logger.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if config.Debug {
			logger.Debugf("%v", msg)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Message updated"))
	}

	if msg.Type == "message_deleted" {
		msg, err := bot.Send(tgbotapi.NewDeleteMessage(msg.Data.ExternalChatID, uid))
		if err != nil {
			raven.CaptureErrorAndWait(err, nil)
			logger.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if config.Debug {
			logger.Debugf("%v", msg)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Message deleted"))
	}
}
