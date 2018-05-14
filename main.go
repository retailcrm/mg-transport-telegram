package main

import (
	"fmt"
	"net/http"

	"github.com/golang-migrate/migrate"
	_ "github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
)

var (
	config = LoadConfig("config.yml")
	orm    = NewDb(config)
)

func main() {
	m, err := migrate.New("file://"+config.Migration.Dir, config.Database.Connection)
	if err != nil {
		fmt.Printf("Migrations path %s does not exist or permission denied\n", config.Migration.Dir)
	}
	m.Up()
	m.Close()

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/settings/", makeHandler(settingsHandler))
	http.HandleFunc("/save/", saveHandler)
	http.HandleFunc("/create/", createHandler)
	http.HandleFunc("/actions/activity", actionActivityHandler)

	SetMsgHandler()

	fmt.Println(http.ListenAndServe(config.HttpServer.Listen, nil))
}
