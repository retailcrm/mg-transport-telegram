package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/getsentry/raven-go"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/retailcrm/mg-transport-api-client-go/v1"
)

func setTransportRoutes() {
	http.HandleFunc("/telegram/", makeHandler(telegramWebhookHandler))
	http.HandleFunc("/webhook/", mgWebhookHandler)
}

// GetBotName function
func GetBotName(bot *tgbotapi.BotAPI) string {
	return bot.Self.FirstName
}

func telegramWebhookHandler(w http.ResponseWriter, r *http.Request, token string) {
	defer r.Body.Close()
	b, err := getBotByToken(token)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		logger.Error(token, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if b.ID == 0 {
		logger.Error(token, "telegramWebhookHandler: missing or deactivated")
		w.WriteHeader(http.StatusOK)
		return
	}

	c := getConnectionById(b.ConnectionID)
	if !c.Active {
		logger.Error(c.ClientID, "telegramWebhookHandler: connection deactivated")
		w.WriteHeader(http.StatusOK)
		return
	}

	var update tgbotapi.Update

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		logger.Error(token, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if config.Debug {
		logger.Debugf("telegramWebhookHandler: %v", string(bytes))
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
		if update.Message.Text == "" {
			setLocale(update.Message.From.LanguageCode)
			update.Message.Text = localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: getMessageID(update.Message)})
		}

		nickname := update.Message.From.UserName
		user := getUserByExternalID(update.Message.From.ID)

		if update.Message.From.UserName == "" {
			nickname = update.Message.From.FirstName
		}

		if user.Expired(config.UpdateInterval) || user.ID == 0 {
			fileID, fileURL, err := GetFileIDAndURL(b.Token, update.Message.From.ID)
			if err != nil {
				raven.CaptureErrorAndWait(err, nil)
				logger.Error(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if fileID != user.UserPhotoID && fileURL != "" {
				picURL, err := UploadUserAvatar(fileURL)
				if err != nil {
					raven.CaptureErrorAndWait(err, nil)
					logger.Error(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				user.UserPhotoID = fileID
				user.UserPhotoURL = picURL
			}

			if user.ExternalID == 0 {
				user.ExternalID = update.Message.From.ID
			}

			err = user.save()
			if err != nil {
				raven.CaptureErrorAndWait(err, nil)
				logger.Error(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		lang := update.Message.From.LanguageCode

		if len(update.Message.From.LanguageCode) > 2 {
			lang = update.Message.From.LanguageCode[:2]
		}

		if config.Debug {
			logger.Debugf("telegramWebhookHandler user %v", user)
		}

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
				Nickname:   nickname,
				Firstname:  update.Message.From.FirstName,
				Avatar:     user.UserPhotoURL,
				Lastname:   update.Message.From.LastName,
				Language:   lang,
			},
			Channel:        b.Channel,
			ExternalChatID: strconv.FormatInt(update.Message.Chat.ID, 10),
		}

		if update.Message.ReplyToMessage != nil {
			snd.Quote = &v1.SendMessageRequestQuote{ExternalID: strconv.Itoa(update.Message.ReplyToMessage.MessageID)}
		}

		data, st, err := client.Messages(snd)
		if err != nil {
			raven.CaptureErrorAndWait(err, nil)
			logger.Error(token, err.Error(), st, data)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if config.Debug {
			logger.Debugf("telegramWebhookHandler Type: SendMessage, Bot: %v, Message: %v, Response: %v", b.ID, snd, data)
		}
	}

	if update.EditedMessage != nil {
		if update.EditedMessage.Text == "" {
			setLocale(update.EditedMessage.From.LanguageCode)
			update.EditedMessage.Text = localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: getMessageID(update.Message)})
		}

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
			logger.Debugf("telegramWebhookHandler Type: UpdateMessage, Bot: %v, Message: %v, Response: %v", b.ID, snd, data)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func mgWebhookHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	clientID := r.Header.Get("Clientid")
	if clientID == "" {
		logger.Error("mgWebhookHandler clientID is empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	c := getConnection(clientID)
	if !c.Active {
		logger.Error(c.ClientID, "mgWebhookHandler: connection deactivated")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Connection deactivated"))
		return
	}

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if config.Debug {
		logger.Debugf("mgWebhookHandler request: %v", string(bytes))
	}

	var msg v1.WebhookRequest
	err = json.Unmarshal(bytes, &msg)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	uid, _ := strconv.Atoi(msg.Data.ExternalMessageID)
	cid, _ := strconv.ParseInt(msg.Data.ExternalChatID, 10, 64)

	b := getBot(c.ID, msg.Data.ChannelID)
	if b.ID == 0 {
		logger.Error(msg.Data.ChannelID, "mgWebhookHandler: missing or deactivated")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing or deactivated"))
		return
	}

	bot, err := tgbotapi.NewBotAPI(b.Token)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if msg.Type == "message_sent" {
		m := tgbotapi.NewMessage(cid, msg.Data.Content)

		if msg.Data.QuoteExternalID != "" {
			qid, err := strconv.Atoi(msg.Data.QuoteExternalID)
			if err != nil {
				raven.CaptureErrorAndWait(err, nil)
				logger.Error(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			m.ReplyToMessageID = qid
		}

		msg, err := bot.Send(m)
		if err != nil {
			raven.CaptureErrorAndWait(err, nil)
			logger.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if config.Debug {
			logger.Debugf("mgWebhookHandler sent %v", msg)
		}

		rsp, err := json.Marshal(map[string]string{"external_message_id": strconv.Itoa(msg.MessageID)})
		if err != nil {
			raven.CaptureErrorAndWait(err, nil)
			logger.Error(err)
			return
		}

		if config.Debug {
			logger.Debugf("mgWebhookHandler sent response %v", string(rsp))
		}

		w.WriteHeader(http.StatusOK)
		w.Write(rsp)
	}

	if msg.Type == "message_updated" {
		msg, err := bot.Send(tgbotapi.NewEditMessageText(cid, uid, msg.Data.Content))
		if err != nil {
			raven.CaptureErrorAndWait(err, nil)
			logger.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if config.Debug {
			logger.Debugf("mgWebhookHandler update %v", msg)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Message updated"))
	}

	if msg.Type == "message_deleted" {
		msg, err := bot.Send(tgbotapi.NewDeleteMessage(cid, uid))
		if err != nil {
			raven.CaptureErrorAndWait(err, nil)
			logger.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if config.Debug {
			logger.Debugf("mgWebhookHandler delete %v", msg)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Message deleted"))
	}
}

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

//UploadUserAvatar function
func UploadUserAvatar(url string) (picURLs3 string, err error) {
	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(
			config.ConfigAWS.AccessKeyID,
			config.ConfigAWS.SecretAccessKey,
			""),
		Region: aws.String(config.ConfigAWS.Region),
	}

	s := session.Must(session.NewSession(s3Config))
	uploader := s3manager.NewUploader(s)

	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return "", errors.New(fmt.Sprintf("get: %v code: %v", url, resp.StatusCode))
	}

	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(config.ConfigAWS.Bucket),
		Key:         aws.String(fmt.Sprintf("%v/%v.jpg", config.ConfigAWS.FolderName, GenerateToken())),
		Body:        resp.Body,
		ContentType: aws.String(config.ConfigAWS.ContentType),
		ACL:         aws.String("public-read"),
	})
	if err != nil {
		return
	}

	picURLs3 = result.Location

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
