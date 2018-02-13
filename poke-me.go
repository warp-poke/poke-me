package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/warp-poke/poke-me/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Panicf("%v", err)
	}
}
