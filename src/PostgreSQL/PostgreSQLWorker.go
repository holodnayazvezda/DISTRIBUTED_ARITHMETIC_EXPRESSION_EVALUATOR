package PostgreSQL

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func init() {
	connStr := fmt.Sprintf("host=DB user=%s password=%s dbname=postgres sslmode=disable", PostgresDBConf.DbUser, PostgresDBConf.DbPass)
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	_, err = DB.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (r_id TEXT PRIMARY KEY, expression TEXT, res TEXT, err TEXT, to_do_time INTEGER)", PostgresDBConf.CalcDbTable))
	if err != nil {
		panic(err)
	}

	_, err = DB.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (ID TEXT PRIMARY KEY, data BYTEA)", PostgresDBConf.ServerDataDbTable))
	if err != nil {
		panic(err)
	}

	_, err = DB.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (ID TEXT PRIMARY KEY, data BYTEA)", PostgresDBConf.TasksDbTable))
	if err != nil {
		panic(err)
	}
}
