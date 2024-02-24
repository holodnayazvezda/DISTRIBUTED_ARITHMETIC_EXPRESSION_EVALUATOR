package services

import (
	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/PostgreSQL"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/orchestrator_server/calc_database_worker"
	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/orchestrator_server/config"

	"github.com/sirupsen/logrus"
)

type DataForRequest struct {
	Id       string        `json:"id"`
	Task     string        `json:"task"`
	WaitTime time.Duration `json:"wait_time"`
}

// Перемешивает слайс
func shuffleSlice(inputSlice []string) {
	rndSource := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(rndSource)
	n := len(inputSlice)

	for i := n - 1; i > 0; i-- {
		// Генерация случайного индекса от 0 до i (включительно)
		j := rng.Intn(i + 1)
		// Обмен значениями между i-м и случайно выбранным индексом
		inputSlice[i], inputSlice[j] = inputSlice[j], inputSlice[i]
	}
}

// Проверка на содержание букв
func containsLetters(inputString string) bool {
	re := regexp.MustCompile("[a-zA-Z]")
	return re.MatchString(inputString)
}

// Удаление объекта из салйса по его ID
func removeFromSlice(slice []string, id int) []string {
	if id >= 0 && id < len(slice) {
		slice = append(slice[:id], slice[id+1:]...)
	} else {
		fmt.Println("Index out of bounds")
	}
	return slice
}

// Выдает любой сервер(агент)
func getRandomAgent() (Server, error) {
	rows, err := PostgreSQL.DB.Query(fmt.Sprintf("SELECT ID FROM %s;", PostgreSQL.PostgresDBConf.ServerDataDbTable))
	if err != nil {
		config.Log.Error(err)
		return Server{}, err
	}
	var keys []string
	for rows.Next() {
		var key string
		err = rows.Scan(&key)
		if err != nil {
			config.Log.Error(err)
			continue
		}
		keys = append(keys, key)
	}
	rows.Close()
	shuffleSlice(keys)

	for _, key := range keys {
		info := PostgreSQL.DB.QueryRow(fmt.Sprintf("SELECT data FROM %s WHERE ID = '%s';", PostgreSQL.PostgresDBConf.ServerDataDbTable, key))

		var dataByte []byte
		err := info.Scan(&dataByte)
		if err != nil {
			config.Log.Error(err)
			continue
		}
		var Serv Server
		json.Unmarshal(dataByte, &Serv)

		if Serv.Status == 1 {
			return Serv, nil
		}
	}
	return Server{}, errors.New("нету доступных серверов")
}

// Отправляет к обработчику (агенту) запрос на выполнение
func sendToCalculation(serv Server, subst string, add, sub, mult, div int) (map[string]interface{}, error) {
	var dataForRequest DataForRequest

	dataForRequest.Id = HashSome(subst)
	dataForRequest.Task = subst
	dataForRequest.WaitTime = CalcWaitTime(subst, add, sub, mult, div)

	jsonData, err := json.Marshal(dataForRequest)
	if err != nil {
		config.Log.Error(err)
		return map[string]interface{}{}, err
	}

	req, err := http.NewRequest("POST", serv.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		config.Log.Error(err)
		return map[string]interface{}{}, err
	}

	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		config.Log.Error(err)
		return map[string]interface{}{}, err
	}
	defer resp.Body.Close()

	var a map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&a)
	if err != nil {
		config.Log.Error(err)
		return map[string]interface{}{}, err
	}

	return a, nil
}

// Забирает результаты обработчки выражения из БД и меняет структуру выражения для последкющей записи итогового выражения в бд
func takeCalRes(ids []string, val string) ([]string, string, string) {
	for {
		if len(ids) == 0 {
			break
		}
		for i, id := range ids {
			if calRes, ok := calc_database_worker.GetCalculationRequest(id); ok && calRes.Res != "" {
				if calRes.Err != "" {
					return []string{}, "", calRes.Err
				}
				val = strings.Replace(val, calRes.Expression, calRes.Res, 1)
				ids = removeFromSlice(ids, i)
				config.Log.WithFields(logrus.Fields{
					"val":        val,
					"Expression": calRes.Expression,
					"Res":        calRes.Res,
					"AllId":      ids,
				}).Info("OK")
			}
		}
		time.Sleep(time.Second)
	}

	return ids, val, ""
}

// Выбирает подвыражения из выражения (разбивает арифметически пример на "действия")
func makeTask(mathExpression string, ids []string) ([]string, []string, error) {
	var fullTask []string
	var resT []string
	var resId []string
	for _, subexpression := range findSubExpressions(mathExpression) {
		res, ok := keepSignBetweenNumbers(subexpression)
		if ok != nil {
			config.Log.Error(ok)
			return []string{}, []string{}, ok
		} else if res == "+" || res == "-" || res == "*" || res == "/" || res == "**" {
			fullTask = append(fullTask, subexpression)
			ids = append(ids, HashSome(subexpression))
		} else {
			resT, resId, ok = makeTask(subexpression, ids)
			if ok != nil {
				config.Log.Error(ok)
				return []string{}, []string{}, ok
			}
			fullTask = append(fullTask, resT...)
			ids = append(ids, resId...)
		}
	}
	return fullTask, ids, nil
}

// ProcessMathExpression Обрабатывает выражение, сначала разбвая его на части (по действиям), а потом отправляя их выполняться на агенты (обработчки)
func ProcessMathExpression(value, id string, addition, subtraction, multiplying, division int) (string, error) {
	tempVal := value
	if containsLetters(value) {
		config.Log.WithField("ex", value).Error("Выражение содержит буквы")
		go calc_database_worker.UpdateTaskInDb(id, "", "Выражение содержит буквы")
		return "", errors.New("выражение содержит буквы")
	}

	randomAgent, err := getRandomAgent()
	if err != nil {
		config.Log.Error(err)
		return "", err
	}

	allIds := []string{}

	var errTake string
	var allTasks []string
	for {
		value, err = extractEverything(value)
		if err != nil {
			config.Log.Error(err)
			return "", err
		}

		allTasks, allIds, _ = makeTask(value, allIds)

		for _, v := range allTasks {
			go sendToCalculation(randomAgent, v, addition, subtraction, multiplying, division)
		}

		allIds, value, errTake = takeCalRes(allIds, value)
		if errTake != "" {
			go calc_database_worker.UpdateTaskInDb(id, "", errTake)
			config.Log.Error(errTake)
			return "", errors.New(errTake)
		}

		value, err = removeParenthesesAroundNumbers(value)
		if err != nil {
			config.Log.Error(err)
			return "", err
		}

		res, ok := keepSignBetweenNumbers(value)
		if ok != nil {
			config.Log.Error(ok)
			return "", ok
		} else if res == "+" || res == "-" || res == "*" || res == "/" || res == "**" {
			go sendToCalculation(randomAgent, value, addition, subtraction, multiplying, division)
			allIds = append(allIds, HashSome(value))
			_, value, errTake = takeCalRes(allIds, value)
			if errTake != "" {
				go calc_database_worker.UpdateTaskInDb(id, "", errTake)
				config.Log.Error(errTake)
				return "", errors.New(errTake)
			}
			go calc_database_worker.UpdateTaskInDb(id, value, "")
			config.Log.Info("Готово - ", tempVal)
			return value, nil
		} else if res == "" || strings.Replace(res, "*", "", 1) == "-e" || strings.Replace(res, "*", "", 1) == "e" || strings.Replace(res, "*", "", 1) == "e-" || strings.Replace(res, "*", "", 1) == "+e" || strings.Replace(res, "*", "", 1) == "e+" {
			go calc_database_worker.UpdateTaskInDb(id, value, "")
			config.Log.Info("Готово - ", tempVal)
			return value, nil
		}
	}
}

// Ищет подвыражения в скобках
func findSubExpressions(expression string) []string {
	var subExpressions []string
	stack := 0
	start := -1

	for i, char := range expression {
		if char == '(' {
			if stack == 0 {
				start = i
			}
			stack++
		} else if char == ')' {
			stack--
			if stack == 0 && start != -1 {
				subexpression := expression[start+1 : i]
				subExpressions = append(subExpressions, subexpression)
				start = -1
			}
		}
	}

	return subExpressions
}

// Считает время ожидания для выражения
func CalcWaitTime(value string, addition, subtraction, multiplying, division int) time.Duration {
	res := 0
	res += strings.Count(value, "+") * addition
	res += strings.Count(value, "-") * subtraction
	res += strings.Count(value, "*") * multiplying
	res += strings.Count(value, "/") * division

	return time.Duration(res) * time.Second
}

// Функция для выделения частей математического выражения в скобки в порядке выполнения (умножение, деление)
func extractMultiplyingAndDivision(mathExpression string) (string, error) {
	// Используем регулярное выражение для поиска сумм и разностей
	matches := regexp.MustCompile(`(\d+)\s*[\*\/]\s*(\d+)`).FindAllString(mathExpression, -1)
	// Заменяем найденные суммы и разности на скобки
	for _, match := range matches {
		mathExpression = strings.Replace(mathExpression, match, "("+match+")", 1)
	}

	return mathExpression, nil
}

// Функция для выделения частей математического выражения в скобки в порядке выполнения (сложение, вычитание)
func extractAdditionAndSubtraction(mathExpression string) (string, error) {
	// Используем регулярное выражение для поиска сумм и разностей
	compile := regexp.MustCompile(`(\d+)\s*[\+\-]\s*(\d+)`)
	matches := compile.FindAllString(mathExpression, -1)
	// Заменяем найденные суммы и разности на скобки
	for _, match := range matches {
		mathExpression = strings.Replace(mathExpression, match, "("+match+")", 1)
	}
	return mathExpression, nil
}

// Функция для выделения частей математического выражения в скобки в порядке выполнения (степень)
func extractDegree(mathExpression string) (string, error) {
	// Используем регулярное выражение для поиска степеней
	compile := regexp.MustCompile(`(\d+)\s*\*\*\s*(\d+)`)
	matches := compile.FindAllString(mathExpression, -1)
	// Заменяем найденные степени на скобки
	for _, match := range matches {
		mathExpression = strings.Replace(mathExpression, match, "("+match+")", 1)
	}
	return mathExpression, nil
}

// Выделяет все части мат выражения
func extractEverything(mathExpression string) (string, error) {
	mathExpression, err := extractDegree(mathExpression)
	if err != nil {
		return "", err
	}
	mathExpression, err = extractMultiplyingAndDivision(mathExpression)
	if err != nil {
		return "", err
	}
	mathExpression, err = extractAdditionAndSubtraction(mathExpression)
	if err != nil {
		return "", err
	}

	return mathExpression, nil
}

// Функция для удаления скобок вокруг чисел
func removeParenthesesAroundNumbers(mathExpression string) (string, error) {
	// Используем регулярное выражение для поиска скобок вокруг чисел
	compile := regexp.MustCompile(`\((\-?\d+)\)`)
	matches := compile.FindAllStringSubmatch(mathExpression, -1)
	// Удаляем найденные скобки вокруг чисел
	for _, match := range matches {
		mathExpression = strings.Replace(mathExpression, match[0], match[1], 1)
	}

	return mathExpression, nil
}

// Функция для удаления всего, кроме знака между двумя числами
func keepSignBetweenNumbers(mathExpression string) (string, error) {
	// Используем регулярное выражение для замены всего, кроме знаков между числами
	compile := regexp.MustCompile(`-?\d+(\.\d+)?`)
	result := compile.ReplaceAllString(mathExpression, "$2")
	compile = regexp.MustCompile(`[()]`)
	result = compile.ReplaceAllString(result, "")

	return strings.Replace(result, " ", "", -1), nil
}

// Проверяет какие задание не выполнены и запускает их выполнение (Запускается при старте)
func CheckNotReadyEx() {
	if data, ok := calc_database_worker.GetAllTasksFromDb(); ok {
		for _, v := range data {
			if v.Res == "" && v.Err == "" {
				config.Log.Info("Начата обработка не завершённой зодачий - " + v.Expression)
				go ProcessMathExpression(v.Expression, v.Reqid, 0, 0, 0, 0)
			}
		}
	}
}
