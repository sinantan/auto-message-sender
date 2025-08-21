package logger

import (
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

func New() *logrus.Logger {
	logger := logrus.New()

	env := strings.ToLower(os.Getenv("ENVIRONMENT"))
	logLevel := strings.ToLower(os.Getenv("LOG_LEVEL"))
	logFormat := strings.ToLower(os.Getenv("LOG_FORMAT"))

	switch logLevel {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn", "warning":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	case "fatal":
		logger.SetLevel(logrus.FatalLevel)
	case "panic":
		logger.SetLevel(logrus.PanicLevel)
	default:
		if env == "production" {
			logger.SetLevel(logrus.InfoLevel)
		} else {
			logger.SetLevel(logrus.DebugLevel)
		}
	}

	if logFormat == "text" || (logFormat == "" && env == "development") {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
			ForceColors:   true,
		})
	} else {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	}

	logger.SetOutput(os.Stdout)

	return logger
}

type LoggerConfig struct {
	Level       string `json:"level"`
	Format      string `json:"format"`
	Output      string `json:"output"`
	Filename    string `json:"filename,omitempty"`
	ForceColors bool   `json:"force_colors"`
}
