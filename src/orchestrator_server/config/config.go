package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type config struct {
	Port string
}

var OrchestratorConf config

// Создает config, записывая в нее данные из .env
func init() {
	RellPath := "src/orchestrator_server/config/.env"
	CurrDir, err := filepath.Abs(".")
	if err != nil {
		Log.WithField("err", "Ошибка при получении текущей директории").Error(err)
		return
	}
	AbsPath := filepath.Join(CurrDir, RellPath)

	err = godotenv.Load(AbsPath)
	if err != nil {
		Log.Error(err)
		panic(err)
	}

	OrchestratorConf.Port = os.Getenv("port")
}
