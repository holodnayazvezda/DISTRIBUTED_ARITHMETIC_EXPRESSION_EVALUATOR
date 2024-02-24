package calc_database_worker

import (
	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/PostgreSQL"
	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/orchestrator_server/structures"
)

// Выдает выражение по его ID
func GetCalculationRequest(rId string) (structures.CalculationRequestDTO, bool) {
	var calculationRequest structures.CalculationRequestDTO
	err := PostgreSQL.DB.QueryRow("SELECT * FROM "+PostgreSQL.PostgresDBConf.CalcDbTable+" WHERE r_id = $1", rId).Scan(&calculationRequest.Rid, &calculationRequest.Expression, &calculationRequest.Res, &calculationRequest.Err, &calculationRequest.ToDoTime)
	if err != nil {
		return calculationRequest, false
	}
	return calculationRequest, true
}
