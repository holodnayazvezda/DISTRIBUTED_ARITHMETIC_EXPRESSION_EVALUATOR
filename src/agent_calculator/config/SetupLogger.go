package config

import (
	"github.com/sirupsen/logrus"
	"os"
)

var Log = logrus.New()

// Создает логер logrus
func init() {
	Log.SetLevel(logrus.DebugLevel)
	Log.SetFormatter(&logrus.JSONFormatter{})
	file, err := os.OpenFile("agent.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		Log.SetOutput(file)
		Log.Info("SetOutput OK")
	} else {
		Log.Info("Не удалось открыть файл логов, используется stderr")
	}
}
