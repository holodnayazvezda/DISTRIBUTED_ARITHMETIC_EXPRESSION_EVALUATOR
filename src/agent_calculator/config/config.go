package config

import (
	"github.com/joho/godotenv"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type config struct {
	AmountOfWorkers   int
	Port              string
	ConnectToBaseUrls []string
	ConnectPath       string
	SelfHost          string
}

var AgentConf config
var TasksChan = make(chan []byte)

// Создает config, записывая в нее данные из .env
func init() {
	RelPath := "src/agent_calculator/config/.env"
	CurrDir, err := filepath.Abs(".")
	if err != nil {
		Log.WithField("err", "Ошибка при получении текущей директории").Error(err)
		return
	}
	AbsPath := filepath.Join(CurrDir, RelPath)

	err = godotenv.Load(AbsPath)
	if err != nil {
		Log.Error(err)
		panic(err)
	}

	AgentConf.AmountOfWorkers, _ = strconv.Atoi(os.Getenv("worker"))
	AgentConf.Port = os.Getenv("port")
	AgentConf.SelfHost = os.Getenv("i_host")
	AgentConf.ConnectToBaseUrls = strings.Split(os.Getenv("connect_to"), ",")
	AgentConf.ConnectPath = os.Getenv("connect_path")
}
