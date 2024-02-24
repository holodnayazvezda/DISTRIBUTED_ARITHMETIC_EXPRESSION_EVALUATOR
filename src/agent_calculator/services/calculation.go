package services

import (
	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/PostgreSQL"
	"encoding/json"
	"fmt"
	"time"

	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/agent_calculator/calc_database_worker"
	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/agent_calculator/config"
	"github.com/Knetic/govaluate"
)

type AnswerDataStruct struct {
	Expression string `json:"ex"`
	Answer     string `json:"answer"`
	Err        string `json:"err"`
}

type JSONdataStruct struct {
	Id       string        `json:"id"`
	Task     string        `json:"task"`
	WaitTime time.Duration `json:"wait_time"`
}

// Запускает горутины (обработчики)
func LaunchWorkers(max int, task chan []byte) {
	for i := 0; i < max; i++ {
		go func() {
			config.Log.Info("Workers are starting")
			var data = &JSONdataStruct{}
			for v := range task {
				json.Unmarshal(v, data)
				config.Log.Info("Start do - " + data.Task)
				calRes, err := TaskCalculation(fmt.Sprintf("%s", data.Task))
				if err != nil {
					config.Log.Error(err)
				}
				time.Sleep(data.WaitTime)
				go calc_database_worker.UpdateCalculationRequest(fmt.Sprintf("%v", data.Id), calRes.Expression, calRes.Answer, calRes.Err)
			}
		}()
	}
}

// Добавляет задание для обработчиков, которые были созданы в LaunchWorkers
func AddTask(task chan []byte) {
	for {
		keys, err := PostgreSQL.DB.Query("SELECT ID FROM " + PostgreSQL.PostgresDBConf.TasksDbTable)
		if err != nil {
			config.Log.Error(err)
		}
		if keys == nil {
			return
		}
		for keys.Next() {
			var key string
			err = keys.Scan(&key)
			if err != nil {
				keys.Close()
				config.Log.Error(err)
				continue
			}
			val := PostgreSQL.DB.QueryRow("SELECT data FROM "+PostgreSQL.PostgresDBConf.TasksDbTable+" WHERE ID = $1", key)
			if val.Err() != nil {
				keys.Close()
				config.Log.Error(val.Err())
				continue
			}
			var jsonByte []byte
			err = val.Scan(&jsonByte)
			if err != nil {
				keys.Close()
				config.Log.Error(err)
				continue
			}

			task <- jsonByte
			keys.Close()
			_, err = PostgreSQL.DB.Exec("DELETE FROM "+PostgreSQL.PostgresDBConf.TasksDbTable+" WHERE ID = $1", key)
			if err != nil {
				config.Log.Error(err)
				continue
			}

		}
		keys.Close()
	}
}

// Вычисляет выражение
func TaskCalculation(data string) (AnswerDataStruct, error) {
	answerDataStruct := AnswerDataStruct{}
	expression, err := govaluate.NewEvaluableExpression(data)
	if err != nil {
		config.Log.WithField("err", "Ошибка при создании выражения").Error(err)
		answerDataStruct.Err = fmt.Sprintf("%v", err)
		return answerDataStruct, err
	}

	result, err := expression.Evaluate(nil)
	if err != nil {
		config.Log.WithField("err", "Ошибка при вычислении выражения").Error(err)
		answerDataStruct.Err = fmt.Sprintf("%v", err)
		return answerDataStruct, err
	}

	answerDataStruct.Expression = data
	answerDataStruct.Answer = fmt.Sprintf("%v", result)

	return answerDataStruct, nil
}

// Проверяет какие задание не выполнены и запускает их выполнение (стартует при запуске агента)
func CheckNotReadyTasks(task chan []byte) {
	var jsonData = &JSONdataStruct{}
	if data, ok := calc_database_worker.GetAllCalculationRequests(); ok {
		for _, v := range data {
			if v.Res == "" && v.Err == "" {
				jsonData.Id = v.RId
				jsonData.Task = v.Expression
				jsonData.WaitTime = time.Second * time.Duration(v.ToDoTime)
				byteData, err := json.Marshal(jsonData)
				if err != nil {
					config.Log.Error(err)
					continue
				}
				task <- byteData
			}
		}
	}
}
