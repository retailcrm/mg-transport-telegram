package main

import (
	"os"

	"github.com/jessevdk/go-flags"

	"github.com/op/go-logging"
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
