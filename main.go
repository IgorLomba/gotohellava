package main

import (
	"gotohellava/cmd"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors:               true,
		ForceQuote:                true,
		EnvironmentOverrideColors: false,
		FullTimestamp:             true,
		TimestampFormat:           "2006-01-02 15:04:05",
	})
	cmd.Execute()
}
