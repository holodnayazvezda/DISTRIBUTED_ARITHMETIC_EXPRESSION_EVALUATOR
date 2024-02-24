package application

import (
	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/agent_calculator/config"
	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/agent_calculator/routers"
	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/agent_calculator/services"
	"github.com/gorilla/mux"
	"net/http"
)

func RunAgent() {
	// запускаем горотины для незавершенных вычислений и попвтки соединения с оркестратором.
	go services.LaunchWorkers(config.AgentConf.AmountOfWorkers, config.TasksChan)
	go services.AddTask(config.TasksChan)
	go services.CheckNotReadyTasks(config.TasksChan)
	go services.PingOrchestrator(config.AgentConf.ConnectToBaseUrls, config.AgentConf.ConnectPath, config.AgentConf.SelfHost)
	// собираем и запускаем сервер
	router := mux.NewRouter()
	router.HandleFunc("/", routers.AddCalculationRequest).Methods("POST")
	router.HandleFunc("/add/{add}", routers.AddWorkers).Methods("POST")
	err := http.ListenAndServe(":"+config.AgentConf.Port, router)
	// обработка ошибок
	if err != nil {
		config.Log.Error(err)
		panic(err)
	}
}
