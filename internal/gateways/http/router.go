package http

import (
	"github.com/gin-gonic/gin"
)

func setupRouter(r *gin.Engine, us UseCases, ws *WebSocketHandler) {
	handlers := &Handlers{us: us, ws: ws, eb: ws.eb}
	r.Use(ContentLengthMiddleware())

	r.HandleMethodNotAllowed = true
	r.NoMethod(handlers.noMethod)

	r.POST("/users", handlers.requireJSONContentType, handlers.postUsers)
	r.OPTIONS("/users", handlers.optionsHandler("POST,OPTIONS"))

	r.GET("/sensors", handlers.requireJSONAccept, handlers.getSensors)
	r.HEAD("/sensors", handlers.requireJSONAccept, handlers.getSensors)
	r.POST("/sensors", handlers.requireJSONContentType, handlers.postSensors)
	r.OPTIONS("/sensors", handlers.optionsHandler("GET,POST,HEAD,OPTIONS"))

	r.GET("/sensors/:sensor_id", handlers.requireJSONAccept, handlers.getSensorsSID)
	r.HEAD("/sensors/:sensor_id", handlers.requireJSONAccept, handlers.getSensorsSID)
	r.OPTIONS("/sensors/:sensor_id", handlers.optionsHandler("GET,HEAD,OPTIONS"))

	r.GET("/users/:user_id/sensors", handlers.requireJSONAccept, handlers.getUsersUIDSensors)
	r.HEAD("/users/:user_id/sensors", handlers.requireJSONAccept, handlers.getUsersUIDSensors)
	r.POST("/users/:user_id/sensors", handlers.requireJSONContentType, handlers.postUsersUIDSensors)
	r.OPTIONS("/users/:user_id/sensors", handlers.optionsHandler("GET,POST,HEAD,OPTIONS"))

	r.POST("/events", handlers.requireJSONContentType, handlers.postEvent)
	r.OPTIONS("/events", handlers.optionsHandler("POST,OPTIONS"))

	r.GET("/sensors/:sensor_id/events", handlers.getSensorsSIDEvents)

	r.GET("sensors/:sensor_id/history", handlers.getSensorsSIDHistory)
}
