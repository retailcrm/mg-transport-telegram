package main

import (
	"os"

	"github.com/getsentry/raven-go"
	"github.com/jessevdk/go-flags"
	"github.com/op/go-logging"
)

type Options struct {
	Config string `short:"c" long:"config" default:"config.yml" description:"Path to configuration file"`
}

const app = "mg-telegram"

var options Options
var parser = flags.NewParser(&options, flags.Default)

func init() {
	raven.SetDSN(config.SentryDSN)
}

var logFormat = logging.MustStringFormatter(
	`%{time:2006-01-02 15:04:05.000} %{level:.4s} => %{message}`,
)

var logger = NewLogger()

func NewLogger() *logging.Logger {
	logger := logging.MustGetLogger(app)
	stdout := logging.NewLogBackend(os.Stdout, "", 0)
	formatBackend := logging.NewBackendFormatter(stdout, logFormat)
	levelBackend := logging.AddModuleLevel(formatBackend)
	levelBackend.SetLevel(config.LogLevel, "")
	logging.SetBackend(levelBackend)

	return logger
}

func main() {
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}
}
