package main

import (
	"os"

	"github.com/jessevdk/go-flags"
)

type Options struct {
	Config string `short:"c" long:"config" default:"config.yml" description:"Path to configuration file"`
}

var options Options
var parser = flags.NewParser(&options, flags.Default)

func main() {
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}
}
