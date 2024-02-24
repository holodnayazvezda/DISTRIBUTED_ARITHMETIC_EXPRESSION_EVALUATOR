package routers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/orchestrator_server/calc_database_worker"
	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/orchestrator_server/services"
)

type TaskStruct struct {
	MathExpression     string `json:"math_expression"`
	AdditionTime       string `json:"addition_time"`       // время сложения
	SubtractionTime    string `json:"subtraction_time"`    // время вычитания
	MultiplicationTime string `json:"multiplication_time"` // время умножения
	DivisionTime       string `json:"division_time"`       // время деления
}

// CALCULATE получает арфметическое вырражение для вычисления, разбивает его на задачи, запускает обработку задач
func CALCULATE(w http.ResponseWriter, r *http.Request) {
	answer := Answer{}
	var data TaskStruct

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		answer.Err = err
		answer.Info = "Не удалось декодировать JSON"
		w.WriteHeader(400)
		jsonResp, _ := json.Marshal(answer)
		w.Write(jsonResp)
		return
	}
	reqId := services.HashSome(data.MathExpression)
	additionTime, _ := strconv.Atoi(fmt.Sprintf("%v", data.AdditionTime))
	subtractionTime, _ := strconv.Atoi(fmt.Sprintf("%v", data.SubtractionTime))
	multiplicationTime, _ := strconv.Atoi(fmt.Sprintf("%v", data.MultiplicationTime))
	divisionTime, _ := strconv.Atoi(fmt.Sprintf("%v", data.DivisionTime))
	waitTime := services.CalcWaitTime(data.MathExpression, additionTime, subtractionTime, multiplicationTime, divisionTime)
	_, ok := calc_database_worker.GetTaskFromDb(reqId)
	if !ok {
		go services.ProcessMathExpression(data.MathExpression, reqId, additionTime, subtractionTime, multiplicationTime, divisionTime)
		go calc_database_worker.AddTaskToDb(data.MathExpression, reqId, int(waitTime.Seconds()))
	} else {
		go calc_database_worker.UpdateTaskInDb(reqId, "", "")
	}

	answer.Data = map[string]string{
		"reqID": reqId,
	}

	jsonResp, _ := json.Marshal(answer)
	w.Write(jsonResp)
}
