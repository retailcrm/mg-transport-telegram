package main

import (
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
	logger = newLogger()
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
	http.ListenAndServe(config.HTTPServer.Listen, nil)
}
