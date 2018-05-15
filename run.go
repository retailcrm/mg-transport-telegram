package main

import (
	"fmt"
	"net/http"

	"os"
	"os/signal"
	"syscall"

	_ "github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
)

var (
	config = LoadConfig("config.yml")
	orm    = NewDb(config)
)

func init() {
	parser.AddCommand("run",
		"Run Message Gateway manager",
		"Run Message Gateway manager.",
		&RunCommand{})
}

type RunCommand struct{}

func (x *RunCommand) Execute(args []string) error {
	go start()

	c := make(chan os.Signal, 1)
	signal.Notify(c)
	for sig := range c {
		switch sig {
		case os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM:
			orm.DB.Close()
			return nil
		default:
		}
	}

	return nil
}

func start() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/settings/", makeHandler(settingsHandler))
	http.HandleFunc("/save/", saveHandler)
	http.HandleFunc("/create/", createHandler)
	http.HandleFunc("/actions/activity", actionActivityHandler)

	SetMsgHandler()

	fmt.Println(http.ListenAndServe(config.HttpServer.Listen, nil))
}
