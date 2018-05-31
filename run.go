package main

import (
	"net/http"

	"os"
	"os/signal"
	"syscall"

	"github.com/getsentry/raven-go"
	_ "github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
)

func init() {
	parser.AddCommand("run",
		"Run mg-telegram",
		"Run mg-telegram.",
		&RunCommand{},
	)
}

// RunCommand struct
type RunCommand struct{}

// Execute command
func (x *RunCommand) Execute(args []string) error {
	config = LoadConfig(options.Config)
	orm = NewDb(config)
	logger = newLogger()
	raven.SetDSN(config.SentryDSN)

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
	setWrapperRoutes()
	setTransportRoutes()
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.ListenAndServe(config.HTTPServer.Listen, nil)
}
