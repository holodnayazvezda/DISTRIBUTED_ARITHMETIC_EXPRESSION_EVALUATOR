package services

import (
	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/PostgreSQL"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/orchestrator_server/config"
	"golang.org/x/text/encoding/unicode"
)

type Server struct {
	Id       string    `json:"id"`
	URL      string    `json:"url"`
	Status   int       `json:"status"`
	LastPing time.Time `json:"last_ping"`
}

// Засовывает сервер(агент) в БД
func serverToDB(s Server) error {
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	_, err = PostgreSQL.DB.Exec(fmt.Sprintf("INSERT INTO %s (ID, data) VALUES ($1, $2) ON CONFLICT (ID) DO UPDATE SET data = EXCLUDED.data;", PostgreSQL.PostgresDBConf.ServerDataDbTable), s.Id, data)
	if err != nil {
		return err
	}
	return nil
}

// RemoveServerFromDB Удаляет сервер(агент) из БД
func RemoveServerFromDB(id string) error {
	_, err := PostgreSQL.DB.Exec(fmt.Sprintf("DELETE FROM %s WHERE ID = $1;", PostgreSQL.PostgresDBConf.ServerDataDbTable), id)
	if err != nil {
		return err
	}
	return nil
}

// HashSome Создает хэш значения
func HashSome(val string) string {
	utf8Encoder := unicode.UTF8.NewEncoder()
	utf8Bytes, _ := utf8Encoder.Bytes([]byte(val))
	hasher := sha256.New()
	hasher.Write(utf8Bytes)
	hashInBytes := hasher.Sum(nil)
	hashString := fmt.Sprintf("%x", hashInBytes)
	return hashString
}

// GetIP Получает IP из запроса
func GetIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		return ip
	}
	return ""
}

// AddAgent Добавляет или обновляет подключение обработчика(агента)
func AddAgent(id, URL string) error {
	info, err := PostgreSQL.DB.Query(fmt.Sprintf("SELECT data FROM %s WHERE ID = $1;", PostgreSQL.PostgresDBConf.ServerDataDbTable), id)
	if err != nil {
		return err
	}
	defer info.Close()
	var dataByte []byte
	if info.Next() {
		err = info.Scan(&dataByte)
		if err != nil {
			return err
		}
	}
	if len(dataByte) > 0 {
		var oldServ Server
		err := json.Unmarshal(dataByte, &oldServ)
		if err != nil {
			return err
		}

		oldServ.LastPing = time.Now()
		oldServ.Status = 1

		err = serverToDB(oldServ)
		if err != nil {
			return err
		}

		config.Log.Info("update connect server " + URL)
		return nil
	}
	server := Server{
		Id:       id,
		URL:      URL,
		Status:   1,
		LastPing: time.Now(),
	}

	err = serverToDB(server)
	if err != nil {
		return err
	}
	config.Log.Info("connect new server " + URL)
	return nil
}

// ReturnAllAgents возвращает все обработчики(агенты)
func ReturnAllAgents() []Server {
	info, err := PostgreSQL.DB.Query(fmt.Sprintf("SELECT ID FROM %s;", PostgreSQL.PostgresDBConf.ServerDataDbTable))
	if err != nil {
		config.Log.Error(err)
		return nil
	}
	var keys []string
	for info.Next() {
		var key string
		err = info.Scan(&key)
		if err != nil {
			config.Log.Error(err)
			continue
		}
		keys = append(keys, key)
	}
	allServers := []Server{}
	info.Close()
	for _, key := range keys {
		info := PostgreSQL.DB.QueryRow(fmt.Sprintf("SELECT data FROM %s WHERE ID = $1;", PostgreSQL.PostgresDBConf.ServerDataDbTable), key)

		var dataByte []byte
		err := info.Scan(&dataByte)
		if err != nil {
			config.Log.Error(err)
		}

		var Serv Server
		err = json.Unmarshal(dataByte, &Serv)
		if err != nil {
			config.Log.Error(err)
			config.Log.Error(err)
		}
		allServers = append(allServers, Serv)
	}
	return allServers
}

// CheckAgent Проверяет последнее подключение к обработчику(агенту)
func CheckAgent() {
	for {
		info, err := PostgreSQL.DB.Query(fmt.Sprintf("SELECT ID FROM %s;", PostgreSQL.PostgresDBConf.ServerDataDbTable))
		if err != nil {
			config.Log.Error(err)
			continue
		}
		var keys []string
		for info.Next() {
			var key string
			err = info.Scan(&key)
			if err != nil {
				config.Log.Error(err)
				continue
			}
			keys = append(keys, key)
		}
		info.Close()
		for _, key := range keys {
			info := PostgreSQL.DB.QueryRow(fmt.Sprintf("SELECT data FROM %s WHERE ID = $1;", PostgreSQL.PostgresDBConf.ServerDataDbTable), key)

			var dataByte []byte
			err := info.Scan(&dataByte)
			if err != nil {
				config.Log.Error(err)
				continue
			}

			var Serv Server
			err = json.Unmarshal(dataByte, &Serv)
			if err != nil {
				config.Log.Error(err)
			}

			if time.Since(Serv.LastPing) >= (time.Minute) {
				Serv.Status = 2
				serverToDB(Serv)
			}
		}
		time.Sleep(time.Second * 20)
	}
}
