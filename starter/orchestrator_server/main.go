package main

import "DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/orchestrator_server/application"

func StartOrchestratorServer() {
	application.RunOrchestrator()
}

func main() {
	StartOrchestratorServer()
}
