package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/getsentry/raven-go"
	"github.com/retailcrm/api-client-go/v5"
)

var (
	templates = template.Must(template.ParseFiles("templates/form.html", "templates/home.html"))
	validPath = regexp.MustCompile("^/(save|settings)/([a-zA-Z0-9]+)$")
)

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
}

func renderTemplate(w http.ResponseWriter, tmpl string, c interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", c)
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
	p := Connection{}
	renderTemplate(w, "home", &p)
}

func addBotHandler(w http.ResponseWriter, r *http.Request) {
	b := &Bot{
		ClientID: string([]byte(r.FormValue("clientId"))),
		Token:    string([]byte(r.FormValue("bot_token"))),
		Active:   true,
	}

	if b.Token == "" {
		http.Error(w, "set bot token", http.StatusInternalServerError)
		return
	}

	cl, _ := getBotByToken(b.Token)
	if cl.ID != 0 {
		http.Error(w, "bot already created", http.StatusInternalServerError)
		return
	}

	bot, err := GetBotInfo(b.Token)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b.Name = GetBotName(bot)

	err = b.createBot()
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func deleteBotHandler(w http.ResponseWriter, r *http.Request) {
	b := &Bot{
		Token: string([]byte(r.FormValue("token"))),
	}

	err := b.deleteBot()
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func mappingHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		return
	}

	var rec []SiteBot

	err = json.Unmarshal(body, &rec)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		return
	}

	err = createMapping(rec)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func settingsHandler(w http.ResponseWriter, r *http.Request, uid string) {
	p, err := getConnection(uid)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if p.ID == 0 {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	bots := Bots{}
	bots.getBotsByClientID(uid)

	client := v5.New(p.APIURL, p.APIKEY)
	sites, _, _ := client.Sites()

	res := struct {
		Conn  *Connection
		Bots  Bots
		Sites map[string]v5.Site
	}{
		p,
		bots,
		sites.Sites,
	}

	renderTemplate(w, "form", res)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	c := &Connection{
		ClientID: string([]byte(r.FormValue("clientId"))),
		APIKEY:   string([]byte(r.FormValue("api_key"))),
		APIURL:   string([]byte(r.FormValue("api_url"))),
	}

	erv := validate(r, *c)
	if erv != "" {
		http.Error(w, erv, http.StatusBadRequest)
		logger.Error(erv)
		return
	}

	err := c.saveConnection()
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("/settings/" + r.FormValue("clientId")))
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	c := &Connection{
		ClientID: GenerateToken(),
		APIURL:   string([]byte(r.FormValue("api_url"))),
		APIKEY:   string([]byte(r.FormValue("api_key"))),
	}

	erv := validate(r, *c)
	if erv != "" {
		http.Error(w, erv, http.StatusBadRequest)
		logger.Error(erv)
		return
	}

	cl, _ := getConnectionByURL(c.APIURL)
	if cl.ID != 0 {
		http.Error(w, "connection already created", http.StatusBadRequest)
		return
	}

	client := v5.New(c.APIURL, c.APIKEY)

	cr, _, errors := client.APICredentials()
	if errors.RuntimeErr != nil {
		logger.Error(errors.RuntimeErr)
		return
	}

	if !cr.Success {
		http.Error(w, "set correct crm url or key", http.StatusBadRequest)
		return
	}

	err := c.createConnection()
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	integration := v5.IntegrationModule{
		Code:            config.AppName,
		IntegrationCode: config.AppName,
		Active:          true,
		Name:            config.AppName,
		ClientID:        c.ClientID,
		BaseURL: fmt.Sprintf(
			"https://%s",
			r.Host,
		),
		AccountURL: fmt.Sprintf(
			"https://%s/settings/%s",
			r.Host,
			c.ClientID,
		),
		Actions: map[string]string{"activity": "/actions/activity"},
		Integrations: &v5.Integrations{
			MgTransport: &v5.MgTransport{
				WebhookUrl: fmt.Sprintf(
					"https://%s/webhook",
					r.Host,
				),
			},
		},
	}

	_, status, errors := client.IntegrationModuleEdit(integration)
	if errors.RuntimeErr != nil {
		logger.Error(errors.RuntimeErr)
		return
	}

	if status >= http.StatusBadRequest {
		logger.Error(errors.ApiErr, c.APIURL)
		return
	}

	w.WriteHeader(http.StatusFound)
	w.Write([]byte("/settings/" + c.ClientID))
}

func activityHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	res := Response{Success: false}

	if r.Method != http.MethodPost {
		res.Error = "set POST"
		jsonString, _ := json.Marshal(res)
		w.Write(jsonString)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		res.Error = "incorrect data"
		jsonString, _ := json.Marshal(res)
		w.Write(jsonString)
		return
	}

	var rec Connection

	err = json.Unmarshal(body, &rec)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		res.Error = "incorrect data"
		jsonString, _ := json.Marshal(res)
		w.Write(jsonString)
		return
	}

	if err := rec.setConnectionActivity(); err != nil {
		raven.CaptureErrorAndWait(err, nil)
		res.Error = "incorrect data"
		jsonString, _ := json.Marshal(res)
		w.Write(jsonString)
		return
	}

	res.Success = true
	jsonString, _ := json.Marshal(res)
	w.Write(jsonString)
}

func validate(r *http.Request, c Connection) string {
	r.ParseForm()
	if len(r.Form) == 0 {
		return "set correct crm url"
	}

	if res, _ := regexp.MatchString(`https:\/\/?[\da-z\.-]+\.(retailcrm\.(ru|pro)|ecomlogic\.com)`, c.APIURL); !res {
		return "set correct crm url"
	}

	return ""
}
