package main

import (
	"net/http"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"net/url"
)

func setTransportRoutes() {
	http.HandleFunc("/add-bot/", addBotHandler)
	http.HandleFunc("/activity-bot/", activityBotHandler)
	http.HandleFunc("/map-bot/", mappingHandler)
}

// GetBotInfo function
func GetBotInfo(token string) (*tgbotapi.BotAPI, error) {
	proxyUrl, err := url.Parse("http://201.132.155.10:8080")
	httpClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}

	bot, err := tgbotapi.NewBotAPIWithClient(token, httpClient)
	//bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return bot, nil
}

// GetBotName function
func GetBotName(bot *tgbotapi.BotAPI) string {
	return bot.Self.FirstName
}
