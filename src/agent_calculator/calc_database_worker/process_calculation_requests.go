package calc_database_worker

import (
	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/PostgreSQL"
	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/agent_calculator/config"
	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/agent_calculator/structures"
	"fmt"
)

// Поучает задание по ID
func GetCalculationRequest(rId string) (structures.CalculationRequestDTO, bool) {
	var calculationRequestDTO structures.CalculationRequestDTO
	err := PostgreSQL.DB.QueryRow("SELECT * FROM "+PostgreSQL.PostgresDBConf.CalcDbTable+" WHERE r_id = $1", rId).Scan(&calculationRequestDTO.RId, &calculationRequestDTO.Expression, &calculationRequestDTO.Res, &calculationRequestDTO.Err, &calculationRequestDTO.ToDoTime)
	if err != nil {
		return calculationRequestDTO, false
	}
	return calculationRequestDTO, calculationRequestDTO.Res != ""
}

// Выдает все имеющиеся задания
func GetAllCalculationRequests() ([]structures.CalculationRequestDTO, bool) {
	rows, err := PostgreSQL.DB.Query("SELECT * FROM " + PostgreSQL.PostgresDBConf.CalcDbTable)
	if err != nil {
		return nil, false
	}
	defer rows.Close()
	var calculationRequestDTO structures.CalculationRequestDTO
	var res []structures.CalculationRequestDTO
	for rows.Next() {
		err = rows.Scan(&calculationRequestDTO.RId, &calculationRequestDTO.Expression, &calculationRequestDTO.Res, &calculationRequestDTO.Err, &calculationRequestDTO.ToDoTime)
		if err != nil {
			return nil, false
		}
		res = append(res, calculationRequestDTO)
	}
	return res, true
}

// AddCalculationRequest Добавляет задание (то что нада посчитать) в БД
func AddCalculationRequest(id, ex string, time int) {
	task := structures.CalculationRequestDTO{
		RId:        id,
		Expression: ex,
		Res:        "",
		Err:        "",
		ToDoTime:   time,
	}
	result, err := PostgreSQL.DB.Exec("INSERT INTO "+PostgreSQL.PostgresDBConf.CalcDbTable+" (r_id, expression, res, err, to_do_time) VALUES ($1, $2, $3, $4, $5)", task.RId, task.Expression, task.Res, task.Err, task.ToDoTime)
	if err != nil {
		config.Log.Error(result)
	}
}

// UpdateCalculationRequest Обновляет данные о задании в БД
func UpdateCalculationRequest(id, ex, res, err string) {
	_, dbErr := PostgreSQL.DB.Exec("UPDATE "+PostgreSQL.PostgresDBConf.CalcDbTable+" SET expression = $1, res = $2, err = $3 WHERE r_id = $4", ex, res, err, id)
	if dbErr != nil {
		config.Log.Error(err)
	}
	fmt.Println(id, ex, res)
}
