package structures

type CalculationRequestDTO struct {
	Rid        string
	Expression string
	Res        string
	Err        string
	ToDoTime   int
}

type TaskDTO struct {
	Expression string
	Reqid      string
	Status     bool
	ToDoTime   int
	Res        string
	Err        string
}
