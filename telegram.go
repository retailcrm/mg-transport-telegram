package main

import (
	"net/http"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func setTransportRoutes() {
	http.HandleFunc("/add-bot/", addBotHandler)
	http.HandleFunc("/activity-bot/", activityBotHandler)
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
