package routers

import (
	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/PostgreSQL"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/agent_calculator/calc_database_worker"
	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/agent_calculator/config"
	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/agent_calculator/services"
	"github.com/gorilla/mux"
)

// Принимает задание и добавляет его в бд
func AddCalculationRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Println("new call!")
	jsonData := &services.JSONdataStruct{}
	err := json.NewDecoder(r.Body).Decode(jsonData)
	if err != nil {
		config.Log.Error(err)
		w.WriteHeader(400)
		resp, _ := json.Marshal(`{"err":"Не удалось декодировать JSON"}`)
		w.Write(resp)
		return
	}

	if _, ok := calc_database_worker.GetCalculationRequest(jsonData.Id); !ok {
		calc_database_worker.AddCalculationRequest(jsonData.Id, jsonData.Task, int(jsonData.WaitTime.Seconds()))

		jsonByte, errJson := json.Marshal(jsonData)
		if errJson != nil {
			config.Log.Error(err)
			return
		}
		_, err = PostgreSQL.DB.Exec("INSERT INTO "+PostgreSQL.PostgresDBConf.TasksDbTable+" (ID, data) VALUES ($1, $2)", jsonData.Id, jsonByte)
		if err != nil {
			config.Log.Error(err)
			return
		}
		config.Log.Info("Add task - " + jsonData.Task)
	}
}

// Добавляет обработчик выражений
func AddWorkers(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	maxAdd := v["add"]
	val, _ := strconv.Atoi(maxAdd)
	go services.LaunchWorkers(val, config.TasksChan)
	config.Log.Info("Add workers - " + maxAdd)
}
