package main

import (
	"os"

	"github.com/jessevdk/go-flags"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/op/go-logging"
	"golang.org/x/text/language"
)

// Options struct
type Options struct {
	Config string `short:"c" long:"config" default:"config.yml" description:"Path to configuration file"`
}

const transport = "mg-telegram"

var (
	config       *TransportConfig
	orm          *Orm
	logger       *logging.Logger
	options      Options
	parser       = flags.NewParser(&options, flags.Default)
	tokenCounter uint32
	localizer    *i18n.Localizer
	bundle       = &i18n.Bundle{DefaultLanguage: language.English}
	matcher      = language.NewMatcher([]language.Tag{
		language.English,
		language.Russian,
		language.Spanish,
	})
)

func main() {
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}
}
