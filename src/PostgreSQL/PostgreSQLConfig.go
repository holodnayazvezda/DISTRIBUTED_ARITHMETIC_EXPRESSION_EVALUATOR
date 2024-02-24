package PostgreSQL

type config struct {
	DbUser            string
	DbPass            string
	CalcDbTable       string
	TasksDbTable      string
	ServerDataDbTable string
	DbHost            string
}

var PostgresDBConf config

func init() {
	PostgresDBConf.DbUser = "postgres"
	PostgresDBConf.DbPass = "1339"
	PostgresDBConf.CalcDbTable = "calc"
	PostgresDBConf.TasksDbTable = "tasks"
	PostgresDBConf.ServerDataDbTable = "server_data"
	PostgresDBConf.DbHost = "db"
}
