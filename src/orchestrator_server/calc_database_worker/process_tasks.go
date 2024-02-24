package calc_database_worker

import (
	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/PostgreSQL"
	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/orchestrator_server/config"
	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/orchestrator_server/structures"
)

// AddTaskToDb Добавляет задание в БД
func AddTaskToDb(ex, reqId string, time int) {
	taskDTO := structures.TaskDTO{
		Expression: ex,
		Reqid:      reqId,
		Status:     false,
		ToDoTime:   time,
		Res:        "",
		Err:        "",
	}
	_, err := PostgreSQL.DB.Exec("INSERT INTO "+PostgreSQL.PostgresDBConf.CalcDbTable+" (expression, r_id, to_do_time, res, err) VALUES ($1, $2, $3, $4, $5)", taskDTO.Expression, taskDTO.Reqid, taskDTO.ToDoTime, taskDTO.Res, taskDTO.Err)
	if err != nil {
		config.Log.Error(err)
	}

}

// UpdateTaskInDb Обновляют данные o задании
func UpdateTaskInDb(reqId, res, error string) {
	_, err := PostgreSQL.DB.Exec("UPDATE "+PostgreSQL.PostgresDBConf.CalcDbTable+" SET res = $1, err = $2 WHERE r_id = $3", res, error, reqId)
	if err != nil {
		config.Log.Error(err)
	}
}

// проверяет готовность задания к выдаче (задаие можно будет получить по id и найти в списке всех заданий только если оно имеет ответ (то есть посчиталось) или содержит ошибку)
func CheckTask(t structures.TaskDTO) bool {
	return (t.Res != "" && t.Reqid != "" && t.Expression != "") || t.Err != ""
}

// GetTaskFromDb Выдает задание по его ID
func GetTaskFromDb(reqId string) (structures.TaskDTO, bool) {
	task, err := PostgreSQL.DB.Query("SELECT * FROM "+PostgreSQL.PostgresDBConf.CalcDbTable+" WHERE r_id = $1", reqId)
	if err != nil {
		config.Log.Error(err)
	}
	defer task.Close()
	if task == nil {
		return structures.TaskDTO{}, false
	}
	var taskDTO structures.TaskDTO
	for task.Next() {
		err := task.Scan(&taskDTO.Reqid, &taskDTO.Expression, &taskDTO.Res, &taskDTO.Err, &taskDTO.ToDoTime)
		if err != nil {
			config.Log.Error(err)
		}
	}
	return taskDTO, CheckTask(taskDTO)
}

// GetAllTasksFromDb Списком выдает все задания
func GetAllTasksFromDb() ([]structures.TaskDTO, bool) {
	db := PostgreSQL.DB
	task, err := db.Query("SELECT * FROM " + PostgreSQL.PostgresDBConf.CalcDbTable)
	if err != nil {
		config.Log.Error(err)
	}
	if task == nil {
		return nil, false
	}
	defer task.Close()
	var taskDTOS []structures.TaskDTO
	for task.Next() {
		var t structures.TaskDTO
		err := task.Scan(&t.Reqid, &t.Expression, &t.Res, &t.Err, &t.ToDoTime)
		if err != nil {
			config.Log.Error(err)
		}
		taskDTOS = append(taskDTOS, t)
	}
	return taskDTOS, true
}
