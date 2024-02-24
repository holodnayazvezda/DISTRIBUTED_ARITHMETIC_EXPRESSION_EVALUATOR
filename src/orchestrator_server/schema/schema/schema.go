package schema

import "github.com/swaggo/swag"

const schemaTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "title": "{{.Title}}",
        "version": "{{.Version}}"
    },
    "paths": {
        "/add_task": {
            "post": {
                "description": "Добавить задания для вычисления",
                "consumes": [
                    "application/json"
                ],
                "tags": [
                    "Calculation"
                ],
                "summary": "Создать задание",
                "parameters": [
                    {
                        "description": "Тело запроса в формате JSON (время для сложения, вычитания, умножения, деления и само выражение)",
                        "name": "TaskDTO",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/schema.CalculationType"
                        }
                    }
                ],
                "responses": {}
            }
        },
        "/task/{id}": {
            "get": {
                "description": "Получить задание(выражение) по ID",
                "tags": [
                    "MathExpression"
                ],
                "summary": "Получить задание по ID",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Task id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {}
            }
        },
        "/get_tasks": {
            "get": {
                "description": "Получить все задания(выражения), которые когда-либо отправлялись на подсчет",
                "tags": [
                    "MathExpression"
                ],
                "summary": "Получить все задания",
                "responses": {}
            }
        }
    },
    "definitions": {
        "schema.CalculationType": {
            "type": "object",
            "properties": {
				"math_expression": {
                    "type": "string",
					"example": "2+2*2"
                },
                "addition_time": {
                    "type": "string",
					"example": "1"
                },
				"subtraction_time": {
                    "type": "string",
					"example": "1"
                },
				"multiplication_time": {
                    "type": "string",
					"example": "1"
                },
                "division_time": {
                    "type": "string",
					"example": "1"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1",
	Schemes:          []string{},
	Title:            "Распределенный вычислитель арифметических выражений",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  schemaTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
