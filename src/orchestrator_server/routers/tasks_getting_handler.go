package routers

import (
	"encoding/json"
	"errors"
	"net/http"

	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/orchestrator_server/calc_database_worker"

	"github.com/gorilla/mux"
)

// получить задание (весь пример)
func GetOneTask(w http.ResponseWriter, r *http.Request) {
	var answer Answer

	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		answer.Err = errors.New("id не обнаружен")
		answer.Info = "/task/{id}, то как должен выглядеть путь"
		w.WriteHeader(400)
		jsonResp, _ := json.Marshal(answer)
		w.Write(jsonResp)
	}
	task, ok := calc_database_worker.GetTaskFromDb(id)
	if ok {
		answer.Data = task
		w.WriteHeader(200)
		jsonResp, _ := json.Marshal(answer)
		w.Write(jsonResp)
		return
	}

	answer.Info = "Не удалось найти запись"
	w.WriteHeader(400)
	jsonResp, _ := json.Marshal(answer)
	w.Write(jsonResp)

}

// получить все задания (примеры)
func GetAllTask(w http.ResponseWriter, r *http.Request) {
	var answer Answer

	if task, ok := calc_database_worker.GetAllTasksFromDb(); ok {
		answer.Data = task
		w.WriteHeader(200)
		jsonResp, _ := json.Marshal(answer)
		w.Write(jsonResp)
		return
	}

	answer.Info = "Нет записей"
	w.WriteHeader(400)
	jsonResp, _ := json.Marshal(answer)
	w.Write(jsonResp)
}
