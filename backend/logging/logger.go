package logging

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Logger = logrus.New()

// Init sets up the global logger (call from main)
func Init() error {
	levelStr := os.Getenv("LOG_LEVEL")
	if levelStr == "" {
		levelStr = "info"
	}

	level, err := logrus.ParseLevel(levelStr)
	if err != nil {
		Logger.WithField("level", levelStr).Errorf("Invalid LOG_LEVEL: %v", err)
		return err
	}

	Logger.SetLevel(level)
	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	Logger.Infof("Logger initialized with level: %s", levelStr)
	return nil
}
