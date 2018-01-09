package main

import (
	"os"

	"github.com/orange-lightsaber/psb-rotatord/cmd"
)

var (
	// VERSION is set during build
	VERSION = "0.1.0"
)

func main() {
	cmd.Exec(VERSION)
	os.Exit(1)
}
