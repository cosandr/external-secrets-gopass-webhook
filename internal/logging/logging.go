package logging

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

func Init() {
	setLogFormat()
	setLogLevel()
}

func setLogFormat() {
	format := os.Getenv("LOG_FORMAT")

	switch strings.ToLower(format) {
	case "", "text":
		log.SetFormatter(&log.TextFormatter{})
	case "json":
		log.SetFormatter(&log.JSONFormatter{})

	default:
		log.SetFormatter(&log.TextFormatter{})
		log.Warnf("unknown log format '%s', using TEXT", format)
	}
}

func setLogLevel() {
	level := os.Getenv("LOG_LEVEL")

	switch strings.ToLower(level) {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "", "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)

	default:
		log.Warnf("unknown log level '%s', using INFO", level)
		log.SetLevel(log.InfoLevel)
	}
}
