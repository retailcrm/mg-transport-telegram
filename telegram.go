package main

import (
	"net/http"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func SetMsgHandler() {
	http.HandleFunc("/add-bot/", addBotHandler)
	http.HandleFunc("/del-bot/", deleteBotHandler)
	http.HandleFunc("/map-bot/", mappingBotHandler)
}

func GetBotInfo(token string) (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return bot, nil
}

func GetNameBot(bot *tgbotapi.BotAPI) string {
	return bot.Self.FirstName
}
