package application

import (
	"net/http"

	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/orchestrator_server/config"
	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/orchestrator_server/routers"
	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/orchestrator_server/schema"
	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/orchestrator_server/services"

	"github.com/gorilla/mux"
)

func RunOrchestrator() {
	// запуск горутин для создания gui и попытки подключению а агентам (воркерам)
	go schema.SetupSwagger()
	go services.CheckAgent()
	go services.CheckNotReadyEx()
	// настройка севера
	router := mux.NewRouter()
	router.HandleFunc("/server/new_connection", routers.Connect).Methods("POST")
	router.HandleFunc("/add_task", routers.CALCULATE).Methods("POST")
	router.HandleFunc("/task/{id}", routers.GetOneTask).Methods("GET")
	router.HandleFunc("/get_tasks", routers.GetAllTask).Methods("GET")
	err := http.ListenAndServe(":"+config.OrchestratorConf.Port, router)
	// обработка ошибок
	if err != nil {
		panic(err)
	}
}
