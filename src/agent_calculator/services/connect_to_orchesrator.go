package services

import (
	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/agent_calculator/config"
	"net/http"
	"time"
)

// Говорит оркестратору, что сервер агент доступенн
func PingOrchestrator(urls []string, path, ihost string) {
	for {
		for _, url := range urls {
			fullUrl := url + path

			request, err := http.NewRequest("POST", fullUrl, nil)
			if err != nil {
				config.Log.Error(err)
				continue
			}

			request.Header.Set("X-Forwarded-For", ihost)

			client := http.Client{}

			response, err := client.Do(request)
			if err != nil {
				config.Log.Error(err)
				continue
			}
			response.Body.Close()
		}

		time.Sleep(time.Second * 30)
	}
}
