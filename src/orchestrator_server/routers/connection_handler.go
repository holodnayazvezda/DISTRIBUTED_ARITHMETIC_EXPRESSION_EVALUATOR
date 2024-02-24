package routers

import (
	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/PostgreSQL"
	"encoding/json"
	"errors"
	"net/http"

	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/orchestrator_server/config"
	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/orchestrator_server/services"

	"github.com/gorilla/mux"
)

type Answer struct {
	Err  error       `json:"err"`
	Data interface{} `json:"data"`
	Info string      `json:"info"`
}

// пытается подключиться к агенту (обработчику)
func Connect(w http.ResponseWriter, r *http.Request) {
	answer := Answer{}

	fromURL := services.GetIP(r)
	if fromURL == "" {
		answer.Err = errors.New("не удалось получить хост")
		answer.Info = "Хост передается в Header X-Forwarded-For"
		data, _ := json.Marshal(answer)
		w.Write(data)
		return
	}

	err := services.AddAgent(services.HashSome(fromURL), fromURL)
	if err != nil {
		w.WriteHeader(400)
		answer.Err = err
		data, _ := json.Marshal(answer)
		w.Write(data)
		return
	}

	data, _ := json.Marshal(answer)
	w.Write(data)
}

// списком выдает все агенты (вычислители)
func AllServ(w http.ResponseWriter, r *http.Request) {
	allAgents := services.ReturnAllAgents()

	a := Answer{
		Err:  nil,
		Data: allAgents,
	}

	data, _ := json.Marshal(a)
	w.Write(data)
}

// удаляет агент (вычислитель)
func DeleteServer(w http.ResponseWriter, r *http.Request) {
	answer := Answer{}

	vars := mux.Vars(r)
	servId, ok := vars["id"]
	if !ok {
		config.Log.WithField("err", "Не удалось найти id").Error(ok)
		w.WriteHeader(400)
		answer.Err = errors.New("не удалось найти id")
		data, _ := json.Marshal(answer)
		w.Write(data)
		return
	}

	_, err := PostgreSQL.DB.Exec("DELETE FROM server_data WHERE ID = $1", servId)
	if err != nil {
		config.Log.WithField("err", "Не удалось удалить сервер").Error(err)
		w.WriteHeader(400)
		answer.Err = err
		data, _ := json.Marshal(answer)
		w.Write(data)
		return
	}

	answer.Data = servId
	answer.Info = "Successful delete"

	data, _ := json.Marshal(answer)
	w.Write(data)
}

func AddWorkerFor(w http.ResponseWriter, r *http.Request) {
	answer := Answer{}
	var server services.Server

	vars := mux.Vars(r)
	servId := vars["id"]
	maxAdd := vars["add"]

	info := PostgreSQL.DB.QueryRow("SELECT data FROM servers WHERE ID = $1", servId)
	if info.Err() != nil {
		config.Log.WithField("err", "Не удалось найти").Error(info.Err())
		w.WriteHeader(400)
		answer.Err = info.Err()
		answer.Info = "Не удалось найти"
		data, _ := json.Marshal(answer)
		w.Write(data)
		return
	}

	var data []byte
	err := info.Scan(&data)
	if err != nil {
		config.Log.WithField("err", "Не удалось прочитать данные").Error(err)
		w.WriteHeader(500)
		answer.Err = err
		answer.Info = "Не удалось прочитать данные"
		data, _ := json.Marshal(answer)
		w.Write(data)
		return
	}

	err = json.Unmarshal(data, &server)
	if err != nil {
		config.Log.WithField("err", "Не удалось декодировать данные сервера").Error(err)
		w.WriteHeader(500)
		answer.Err = err
		answer.Info = "Не удалось декодировать данные сервера"
		data, _ := json.Marshal(answer)
		w.Write(data)
		return
	}
	if err != nil {
		config.Log.WithField("err", "Не удалось декодировать данные сервера").Error(err)
		w.WriteHeader(500)
		answer.Err = err
		answer.Info = "Не удалось декодировать данные сервера"
		data, _ = json.Marshal(answer)
		w.Write(data)
		return
	}

	fullUrl := server.URL + "add/" + maxAdd

	req, err := http.NewRequest("POST", fullUrl, nil)
	if err != nil {
		config.Log.Error(err)
		config.Log.WithField("err", "Не удалось создать запрос").Error(err)
		w.WriteHeader(500)
		answer.Err = err
		answer.Info = "Не удалось создать запрос"
		data, _ = json.Marshal(answer)
		w.Write(data)
		return
	}

	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		config.Log.Error(err)
		config.Log.WithField("err", "Не удалось отправить запрос").Error(err)
		w.WriteHeader(500)
		answer.Err = err
		answer.Info = "Не удалось отправить запрос"
		data, _ = json.Marshal(answer)
		w.Write(data)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(200)
	answer.Data = maxAdd
	answer.Info = "Воркеры добавлены"
	data, _ = json.Marshal(answer)
	w.Write(data)
}
