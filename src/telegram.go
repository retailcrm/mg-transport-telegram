package main

import "github.com/go-telegram-bot-api/telegram-bot-api"

//GetFileIDAndURL function
func GetFileIDAndURL(token string, userID int) (fileID, fileURL string, err error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return
	}

	bot.Debug = config.Debug

	res, err := bot.GetUserProfilePhotos(
		tgbotapi.UserProfilePhotosConfig{
			UserID: userID,
			Limit:  1,
		},
	)
	if err != nil {
		return
	}

	if config.Debug {
		logger.Debugf("GetFileIDAndURL Photos: %v", res.Photos)
	}

	if len(res.Photos) > 0 {
		fileID = res.Photos[0][len(res.Photos[0])-1].FileID
		fileURL, err = bot.GetFileDirectURL(fileID)
	}

	return
}

func getMessageID(data *tgbotapi.Message) string {
	switch {
	case data.Sticker != nil:
		return "sticker"
	case data.Audio != nil:
		return "audio"
	case data.Contact != nil:
		return "contact"
	case data.Document != nil:
		return "document"
	case data.Location != nil:
		return "location"
	case data.Video != nil:
		return "video"
	case data.Voice != nil:
		return "voice"
	case data.Photo != nil:
		return "photo"
	default:
		return "undefined"
	}
}
