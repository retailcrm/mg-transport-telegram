package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/getsentry/raven-go"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/retailcrm/api-client-go/v5"
	"github.com/retailcrm/mg-transport-api-client-go/v1"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

var (
	templates = template.Must(template.ParseFiles("templates/layout.html", "templates/form.html", "templates/home.html"))
	validPath = regexp.MustCompile(`^/(save|settings|telegram)/([a-zA-Z0-9-:_+]+)$`)
	localizer *i18n.Localizer
	bundle    = &i18n.Bundle{DefaultLanguage: language.English}
	matcher   = language.NewMatcher([]language.Tag{
		language.English,
		language.Russian,
		language.Spanish,
	})
)

func init() {
	bundle.RegisterUnmarshalFunc("yml", yaml.Unmarshal)
	files, err := ioutil.ReadDir("translate")
	if err != nil {
		logger.Error(err)
	}
	for _, f := range files {
		if !f.IsDir() {
			bundle.MustLoadMessageFile("translate/" + f.Name())
		}
	}
}

func setLocale(al string) {
	tag, _ := language.MatchStrings(matcher, al)
	localizer = i18n.NewLocalizer(bundle, tag.String())
}

// Response struct
type Response struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func setWrapperRoutes() {
	http.HandleFunc("/", connectHandler)
	http.HandleFunc("/settings/", makeHandler(settingsHandler))
	http.HandleFunc("/save/", saveHandler)
	http.HandleFunc("/create/", createHandler)
	http.HandleFunc("/actions/activity", activityHandler)
	http.HandleFunc("/add-bot/", addBotHandler)
	http.HandleFunc("/activity-bot/", activityBotHandler)
}

func renderTemplate(w http.ResponseWriter, tmpl string, c interface{}) {
	tm, err := template.ParseFiles("templates/layout.html", "templates/"+tmpl+".html")
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tm.Execute(w, &c)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func connectHandler(w http.ResponseWriter, r *http.Request) {
	setLocale(r.Header.Get("Accept-Language"))
	p := Connection{}

	res := struct {
		Conn   *Connection
		Locale map[string]interface{}
	}{
		&p,
		map[string]interface{}{
			"ButtonSave": localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "button_save"}),
			"ApiKey":     localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "api_key"}),
			"Title":      localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "title"}),
		},
	}
	renderTemplate(w, "home", &res)
}

func addBotHandler(w http.ResponseWriter, r *http.Request) {
	setLocale(r.Header.Get("Accept-Language"))
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		http.Error(w, localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "error_adding_bot"}), http.StatusInternalServerError)
		return
	}

	var b Bot

	err = json.Unmarshal(body, &b)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		http.Error(w, localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "error_adding_bot"}), http.StatusInternalServerError)
		return
	}

	if b.Token == "" {
		http.Error(w, localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "no_bot_token"}), http.StatusBadRequest)
		return
	}

	c := getConnection(b.ClientID)
	if c.MGURL == "" || c.MGToken == "" {
		http.Error(w, localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "not_found_account"}), http.StatusBadRequest)
		logger.Error(b.ClientID, "MGURL or MGToken is empty")
		return
	}

	cl := getBotByToken(b.Token)
	if cl.ID != 0 {
		http.Error(w, localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "bot_already_created"}), http.StatusBadRequest)
		return
	}

	bot, err := GetBotInfo(b.Token)
	if err != nil {
		logger.Error(b.Token, err.Error())
		http.Error(w, localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "incorrect_token"}), http.StatusBadRequest)
		return
	}

	bot.Debug = false

	wr, err := bot.SetWebhook(tgbotapi.NewWebhook("https://" + config.HTTPServer.Host + "/telegram/" + bot.Token))
	if err != nil {
		logger.Error(b.Token, err.Error())
		http.Error(w, localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "error_creating_webhook"}), http.StatusBadRequest)
		return
	}

	if !wr.Ok {
		logger.Error(b.Token, wr.ErrorCode, wr.Result)
		http.Error(w, localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "error_creating_webhook"}), http.StatusBadRequest)
		return
	}

	_, err = bot.GetWebhookInfo()
	if err != nil {
		logger.Error(b.Token, err.Error())
		http.Error(w, localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "error_creating_webhook"}), http.StatusBadRequest)
		return
	}

	b.Name = GetBotName(bot)

	ch := v1.Channel{
		Type: "telegram",
		Events: []string{
			"message_sent",
			"message_updated",
			"message_deleted",
			"message_read",
		},
	}

	var client = v1.New(c.MGURL, c.MGToken)
	data, status, err := client.ActivateTransportChannel(ch)
	if status != http.StatusCreated {
		http.Error(w, localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "error_activating_channel"}), http.StatusBadRequest)
		logger.Error(c.APIURL, status, err.Error(), data)
		return
	}

	b.Channel = data.ChannelID
	b.Active = true

	err = b.createBot()
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		http.Error(w, localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "error_adding_bot"}), http.StatusInternalServerError)
		logger.Error(c.APIURL, err.Error())
		return
	}

	jsonString, err := json.Marshal(b)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		http.Error(w, localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "error_adding_bot"}), http.StatusInternalServerError)
		logger.Error(c.APIURL, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(jsonString)
}

func activityBotHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		http.Error(w, localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "error_save"}), http.StatusInternalServerError)
		return
	}

	var b Bot

	err = json.Unmarshal(body, &b)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		http.Error(w, localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "error_save"}), http.StatusInternalServerError)
		return
	}

	c := getConnection(b.ClientID)
	if c.MGURL == "" || c.MGToken == "" {
		http.Error(w, localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "not_found_account"}), http.StatusBadRequest)
		logger.Error(b.ClientID, "MGURL or MGToken is empty")
		return
	}

	ch := v1.Channel{
		ID:   getBotChannelByToken(b.Token),
		Type: "telegram",
		Events: []string{
			"message_sent",
			"message_updated",
			"message_deleted",
			"message_read",
		},
	}

	var client = v1.New(c.MGURL, c.MGToken)

	if b.Active {
		data, status, err := client.DeactivateTransportChannel(ch.ID)
		if status > http.StatusOK {
			http.Error(w, localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "error_deactivating_channel"}), http.StatusBadRequest)
			logger.Error(b.ClientID, status, err.Error(), data)
			return
		}
	} else {
		data, status, err := client.ActivateTransportChannel(ch)
		if status > http.StatusCreated {
			http.Error(w, localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "error_activating_channel"}), http.StatusBadRequest)
			logger.Error(b.ClientID, status, err.Error(), data)
			return
		}
	}

	err = b.setBotActivity()
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		logger.Error(b.ClientID, err.Error())
		http.Error(w, localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "error_save"}), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func settingsHandler(w http.ResponseWriter, r *http.Request, uid string) {
	setLocale(r.Header.Get("Accept-Language"))

	p := getConnection(uid)
	if p.ID == 0 {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	bots := Bots{}
	err := bots.getBotsByClientID(uid)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		http.Error(w, localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "error_save"}), http.StatusInternalServerError)
		return
	}

	res := struct {
		Conn   *Connection
		Bots   Bots
		Locale map[string]interface{}
	}{
		p,
		bots,
		map[string]interface{}{
			"ButtonSave":    localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "button_save"}),
			"ApiKey":        localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "api_key"}),
			"TabSettings":   localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "tab_settings"}),
			"TabBots":       localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "tab_bots"}),
			"TableName":     localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "table_name"}),
			"TableToken":    localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "table_token"}),
			"AddBot":        localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "add_bot"}),
			"TableActivity": localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "table_activity"}),
			"Title":         localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "title"}),
		},
	}

	renderTemplate(w, "form", res)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	setLocale(r.Header.Get("Accept-Language"))

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		http.Error(w, localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "error_save"}), http.StatusInternalServerError)
		return
	}

	var c Connection

	err = json.Unmarshal(body, &c)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		http.Error(w, localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "error_save"}), http.StatusInternalServerError)
		return
	}

	err = validateCrmSettings(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		logger.Error(c.APIURL, err.Error())
		return
	}

	err = c.saveConnection()
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		http.Error(w, localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "error_save"}), http.StatusInternalServerError)
		logger.Error(c.APIURL, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "successful"})))
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	setLocale(r.Header.Get("Accept-Language"))

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		http.Error(w, localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "error_save"}), http.StatusInternalServerError)
		return
	}

	var c Connection

	err = json.Unmarshal(body, &c)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		http.Error(w, localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "error_save"}), http.StatusInternalServerError)
		return
	}

	c.ClientID = GenerateToken()

	err = validateCrmSettings(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		logger.Error(c.APIURL, err.Error())
		return
	}

	cl := getConnectionByURL(c.APIURL)
	if cl.ID != 0 {
		http.Error(w, localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "connection_already_created"}), http.StatusBadRequest)
		return
	}

	client := v5.New(c.APIURL, c.APIKEY)

	cr, status, errr := client.APICredentials()
	if errr.RuntimeErr != nil {
		raven.CaptureErrorAndWait(errr.RuntimeErr, nil)
		http.Error(w, localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "not_found_account"}), http.StatusInternalServerError)
		logger.Error(c.APIURL, status, errr.RuntimeErr, cr)
		return
	}

	if !cr.Success {
		http.Error(w, localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "incorrect_url_key"}), http.StatusBadRequest)
		logger.Error(c.APIURL, status, errr.ApiErr, cr)
		return
	}

	//TODO: проверка на необходимые методы cr.Credentials

	integration := v5.IntegrationModule{
		Code:            transport,
		IntegrationCode: transport,
		Active:          true,
		Name:            "Telegram",
		ClientID:        c.ClientID,
		Logo: fmt.Sprintf(
			"https://%s/web/telegram_logo.svg",
			config.HTTPServer.Host,
		),
		BaseURL: fmt.Sprintf(
			"https://%s",
			config.HTTPServer.Host,
		),
		AccountURL: fmt.Sprintf(
			"https://%s/settings/%s",
			config.HTTPServer.Host,
			c.ClientID,
		),
		Actions: map[string]string{"activity": "/actions/activity"},
		Integrations: &v5.Integrations{
			MgTransport: &v5.MgTransport{
				WebhookUrl: fmt.Sprintf(
					"https://%s/webhook",
					config.HTTPServer.Host,
				),
			},
		},
	}

	data, status, errr := client.IntegrationModuleEdit(integration)
	if errr.RuntimeErr != nil {
		raven.CaptureErrorAndWait(errr.RuntimeErr, nil)
		http.Error(w, localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "error_creating_integration"}), http.StatusInternalServerError)
		logger.Error(c.APIURL, status, errr.RuntimeErr, data)
		return
	}

	if status >= http.StatusBadRequest {
		http.Error(w, errr.ApiErr, http.StatusBadRequest)
		logger.Error(c.APIURL, status, errr.ApiErr, data)
		return
	}

	c.MGURL = data.Info["baseUrl"]
	c.MGToken = data.Info["token"]
	c.Active = true

	err = c.createConnection()
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		http.Error(w, localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "error_creating_connection"}), http.StatusInternalServerError)
		return
	}

	res := struct {
		Url     string
		Message string
	}{
		Url:     "/settings/" + c.ClientID,
		Message: localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "successful"}),
	}

	jss, err := json.Marshal(res)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		http.Error(w, localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "error_creating_connection"}), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusFound)
	w.Write(jss)
}

func activityHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	res := Response{Success: false}

	if r.Method != http.MethodPost {
		res.Error = localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "set_method"})
		jsonString, err := json.Marshal(res)
		if err != nil {
			raven.CaptureErrorAndWait(err, nil)
			logger.Error(err)
			return
		}
		w.Write(jsonString)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		res.Error = localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "incorrect_data"})
		jsonString, err := json.Marshal(res)
		if err != nil {
			raven.CaptureErrorAndWait(err, nil)
			logger.Error(err)
			return
		}
		w.Write(jsonString)
		return
	}

	var rec Connection

	err = json.Unmarshal(body, &rec)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		res.Error = localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "incorrect_data"})
		jsonString, err := json.Marshal(res)
		if err != nil {
			raven.CaptureErrorAndWait(err, nil)
			logger.Error(err)
			return
		}
		w.Write(jsonString)
		return
	}

	if err := rec.setConnectionActivity(); err != nil {
		raven.CaptureErrorAndWait(err, nil)
		res.Error = localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "incorrect_data"})
		jsonString, err := json.Marshal(res)
		if err != nil {
			raven.CaptureErrorAndWait(err, nil)
			logger.Error(err)
			return
		}
		w.Write(jsonString)
		return
	}

	res.Success = true
	jsonString, err := json.Marshal(res)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		logger.Error(err)
		return
	}

	w.Write(jsonString)
}

func validateCrmSettings(c Connection) error {
	if c.APIURL == "" || c.APIKEY == "" {
		return errors.New(localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "missing_url_key"}))
	}

	if res, _ := regexp.MatchString(`https://?[\da-z\.-]+\.(retailcrm\.(ru|pro)|ecomlogic\.com)`, c.APIURL); !res {
		return errors.New(localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "incorrect_url"}))
	}

	return nil
}
