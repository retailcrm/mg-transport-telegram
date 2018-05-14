package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/retailcrm/api-client-go/v5"
)

var (
	templates = template.Must(template.ParseFiles("templates/form.html", "templates/home.html"))
	validPath = regexp.MustCompile("^/(save|settings)/([a-zA-Z0-9]+)$")
)

type Response struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func renderTemplate(w http.ResponseWriter, tmpl string, c interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", c)
	if err != nil {
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

func homeHandler(w http.ResponseWriter, r *http.Request) {
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
		fmt.Println("set bot token")
		return
	}

	cl, _ := getByToken(b.Token)
	if cl.ID != 0 {
		http.Error(w, "bot already created", http.StatusInternalServerError)
		fmt.Println("bot already created")
		return
	}

	bot, err := GetBotInfo(b.Token)
	b.Name = GetNameBot(bot)

	err = b.createBot()
	if err != nil {
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func mappingBotHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var rec []SiteBot

	err = json.Unmarshal(body, &rec)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = createSiteBots(rec)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func settingsHandler(w http.ResponseWriter, r *http.Request, uid string) {
	p, err := get(uid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bots := Bots{}
	bots.getBotsByClientId(uid)

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
		fmt.Println(erv)
		return
	}

	err := c.save()
	if err != nil {
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
		fmt.Println(erv)
		return
	}

	cl, _ := getByUrlCrm(c.APIURL)
	if cl.ID != 0 {
		http.Error(w, "connection already created", http.StatusBadRequest)
		fmt.Println()
		return
	}

	client := v5.New(c.APIURL, c.APIKEY)

	cr, _, errors := client.APICredentials()
	if errors.RuntimeErr != nil {
		fmt.Println(errors.RuntimeErr)
		return
	}

	if !cr.Success {
		http.Error(w, "set correct crm url or key", http.StatusBadRequest)
		fmt.Println("set correct crm url or key")
		return
	}

	err := c.create()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	integration := v5.IntegrationModule{
		Code:            "test-123",
		IntegrationCode: "test-123",
		Active:          true,
		Name:            "test-telegram",
		ClientID:        c.ClientID,
		BaseURL:         "https://test.te",
		AccountURL: fmt.Sprintf(
			"%s/settings/%s",
			"https://test.te",
			c.ClientID,
		),
		Actions: map[string]string{"activity": "/actions/activity"},
		//Integrations: &v5.Integrations{
		//	MgTransport: &v5.MgTransport{
		//		WebhookUrl: "https://test.te/telegram",
		//	},
		//},
	}

	_, status, errors := client.IntegrationModuleEdit(integration)

	if errors.RuntimeErr != nil {
		fmt.Printf("%v", errors.Error())
	}

	if status >= http.StatusBadRequest {
		fmt.Printf("%v", errors.ApiError())
	}

	w.WriteHeader(http.StatusFound)
	w.Write([]byte("/settings/" + c.ClientID))
}

func actionActivityHandler(w http.ResponseWriter, r *http.Request) {
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
		fmt.Println(err.Error())
		res.Error = "incorrect data"
		jsonString, _ := json.Marshal(res)
		w.Write(jsonString)
		return
	}

	var rec Connection

	err = json.Unmarshal(body, &rec)
	if err != nil {
		fmt.Println(err.Error())
		res.Error = "incorrect data"
		jsonString, _ := json.Marshal(res)
		w.Write(jsonString)
		return
	}

	if err := rec.setActive(); err != nil {
		fmt.Println(err.Error())
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

	if res, _ := regexp.MatchString(`https:\/\/?[\da-z\.-]+\.retailcrm.ru`, c.APIURL); !res {
		return "set correct crm url"
	}

	return ""
}
