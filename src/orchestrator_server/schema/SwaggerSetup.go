package schema

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/orchestrator_server/routers"
	_ "DISTRIBUTED_ARITHMETIC_EXPRESSION_EVALUATOR/src/orchestrator_server/schema/schema" // Это импорт, который содержит аннотации Swagger

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Создает swagger
type TaskStruct struct {
	MathExpression     string `json:"math_expression"`
	AdditionTime       string `json:"addition_time"`
	SubtractionTime    string `json:"subtraction_time"`
	MultiplicationTime string `json:"multiplication_time"`
	DivisionTime       string `json:"division_time"`
}

// для отправки здадания на выполнение
func AddTask(c *gin.Context) {

	// Construct the DELETE request
	apiURL := "http://localhost:4000/add_task"

	var requestData TaskStruct
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Marshal the requestData struct to JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("Content-Type", "application/json")

	// Send the DELETE request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Respond to the client with the API response
	var data routers.Answer
	json.Unmarshal(body, &data)
	c.JSON(http.StatusOK, data)
}

// для получения задания по id
func GetTask(c *gin.Context) {
	// Send a request to your API

	taskID := c.Param("id")
	apiURL := "http://localhost:4000/task/" + taskID // Update with your actual API URL

	resp, err := http.Get(apiURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Respond to the client with the API response
	var data routers.Answer
	json.Unmarshal(body, &data)
	c.JSON(http.StatusOK, data)
}

// для получения списка всех заданий
func GetAllTask(c *gin.Context) {
	// Send a request to your API

	apiURL := "http://localhost:4000/get_tasks"

	resp, err := http.Get(apiURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Respond to the client with the API response
	var data routers.Answer
	json.Unmarshal(body, &data)
	c.JSON(http.StatusOK, data)
}

// запускает swagger
func SetupSwagger() {
	r := gin.New()

	r.POST("/add_task", AddTask)
	r.GET("/task/:id", GetTask)
	r.GET("/get_tasks", GetAllTask)
	// Swagger routes
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.Run(":8080")
}
