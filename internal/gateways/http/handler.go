package http

import (
	"errors"
	"homework/internal/domain"
	"homework/internal/usecase"
	"homework/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-openapi/swag"
)

const (
	ErrInvalidJSONFormat     = "Неверный формат JSON"
	ErrValidation            = "Ошибка при валидации"
	ErrUserNotFound          = "Пользователь не найден"
	ErrUserCreateFailed      = "Не удалось создать пользователя"
	ErrSensorNotFound        = "Сенсор не найден"
	ErrSensorCreateFailed    = "Не удалось создать сенсор"
	ErrSensorAttach          = "Ошибка при привязке сенсора к пользователю"
	ErrEventProcessingFailed = "Ошибка обработки события"
	ErrInvalidIDFormat       = "Некорректный формат ID"
	ErrInvalidDateFormat     = "Некорректный формат даты"
)

type Handlers struct {
	us UseCases
	ws *WebSocketHandler
	eb *EventBroker
}

func (h *Handlers) optionsHandler(allowedMethods string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Allow", allowedMethods)
		c.Status(http.StatusNoContent)
	}
}

func (h *Handlers) requireJSONAccept(c *gin.Context) {
	if c.GetHeader("Accept") != "application/json" {
		c.AbortWithStatus(http.StatusNotAcceptable)
	}
}

func (h *Handlers) requireJSONContentType(c *gin.Context) {
	if c.ContentType() != "application/json" {
		c.AbortWithStatus(http.StatusUnsupportedMediaType)
	}
}

func (h *Handlers) handleError(c *gin.Context, err error, status int, message string) {
	if err != nil {
		modelError := models.Error{
			Reason: swag.String(message),
		}
		err := modelError.Validate(nil)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, models.Error{
				Reason: swag.String(ErrValidation),
			})
			return
		}
		c.AbortWithStatusJSON(status, modelError)
	}
}

func (h *Handlers) parseId(c *gin.Context, key string) int64 {
	id, err := strconv.ParseInt(c.Param(key), 10, 64)
	h.handleError(c, err, http.StatusUnprocessableEntity, ErrInvalidIDFormat)
	return id
}

func (h *Handlers) parseDate(c *gin.Context, key string) time.Time {
	date, err := time.Parse(time.RFC3339Nano, c.Query(key))
	h.handleError(c, err, http.StatusBadRequest, ErrInvalidDateFormat)
	return date
}

func (h *Handlers) noMethod(c *gin.Context) {
	c.Header("Allow", "POST,OPTIONS")
	c.JSON(http.StatusMethodNotAllowed, nil)
}

func (h *Handlers) postUsers(c *gin.Context) {
	var user models.UserToCreate
	h.handleError(c, c.ShouldBindJSON(&user), http.StatusBadRequest, ErrInvalidJSONFormat)
	h.handleError(c, user.Validate(nil), http.StatusUnprocessableEntity, ErrValidation)

	result, err := h.us.User.RegisterUser(c.Request.Context(), &domain.User{Name: *user.Name})
	h.handleError(c, err, http.StatusUnprocessableEntity, ErrUserCreateFailed)
	c.JSON(http.StatusOK, result)
}

func (h *Handlers) getSensors(c *gin.Context) {
	sensors, err := h.us.Sensor.GetSensors(c.Request.Context())
	h.handleError(c, err, http.StatusInternalServerError, ErrSensorNotFound)
	c.JSON(http.StatusOK, sensors)
}

func (h *Handlers) postSensors(c *gin.Context) {
	var sensor models.SensorToCreate
	h.handleError(c, c.ShouldBindJSON(&sensor), http.StatusBadRequest, ErrInvalidJSONFormat)
	h.handleError(c, sensor.Validate(nil), http.StatusUnprocessableEntity, ErrValidation)

	result, err := h.us.Sensor.RegisterSensor(c.Request.Context(), &domain.Sensor{
		Type:         domain.SensorType(*sensor.Type),
		SerialNumber: *sensor.SerialNumber,
		Description:  *sensor.Description,
		IsActive:     *sensor.IsActive,
	})
	h.handleError(c, err, http.StatusInternalServerError, ErrSensorCreateFailed)
	c.JSON(http.StatusOK, result)
}

func (h *Handlers) getSensorsSID(c *gin.Context) {
	sensorID := h.parseId(c, "sensor_id")
	sensor, err := h.us.Sensor.GetSensorByID(c.Request.Context(), sensorID)
	h.handleError(c, err, http.StatusNotFound, ErrSensorNotFound)
	c.JSON(http.StatusOK, sensor)
}

func (h *Handlers) getUsersUIDSensors(c *gin.Context) {
	userID := h.parseId(c, "user_id")
	sensors, err := h.us.User.GetUserSensors(c.Request.Context(), userID)
	h.handleError(c, err, http.StatusNotFound, ErrUserNotFound)
	c.JSON(http.StatusOK, sensors)
}

func (h *Handlers) postUsersUIDSensors(c *gin.Context) {
	userID := h.parseId(c, "user_id")
	var sensorID models.SensorToUserBinding
	h.handleError(c, c.ShouldBindJSON(&sensorID), http.StatusBadRequest, ErrInvalidJSONFormat)
	h.handleError(c, sensorID.Validate(nil), http.StatusUnprocessableEntity, ErrValidation)
	if err := h.us.User.AttachSensorToUser(c.Request.Context(), userID, *sensorID.SensorID); err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			h.handleError(c, err, http.StatusNotFound, ErrUserNotFound)
		} else {
			h.handleError(c, err, http.StatusUnprocessableEntity, ErrSensorAttach)
		}
		return
	}
	c.JSON(http.StatusCreated, nil)
}

func (h *Handlers) postEvent(c *gin.Context) {
	var event models.SensorEvent
	h.handleError(c, c.ShouldBindJSON(&event), http.StatusBadRequest, ErrInvalidJSONFormat)
	h.handleError(c, event.Validate(nil), http.StatusUnprocessableEntity, ErrValidation)
	tmp := &domain.Event{
		Payload:            *event.Payload,
		Timestamp:          time.Now(),
		SensorSerialNumber: *event.SensorSerialNumber,
	}
	h.handleError(
		c,
		h.us.Event.ReceiveEvent(c.Request.Context(), tmp),
		http.StatusInternalServerError,
		ErrEventProcessingFailed,
	)
	sensor, err := h.us.Sensor.GetSensorBySerialNumber(c.Request.Context(), *event.SensorSerialNumber)
	h.handleError(c, err, http.StatusNotFound, ErrSensorNotFound)
	h.eb.Publish(sensor.ID, tmp)
	c.Status(http.StatusCreated)
}

func (h *Handlers) getSensorsSIDEvents(c *gin.Context) {
	sensorID := h.parseId(c, "sensor_id")
	_, err := h.us.Sensor.GetSensorByID(c.Request.Context(), sensorID)
	if err != nil {
		h.handleError(c, err, http.StatusNotFound, ErrSensorNotFound)
		return
	}
	_ = h.ws.Handle(c, sensorID)
}

func (h *Handlers) getSensorsSIDHistory(c *gin.Context) {
	sensorID := h.parseId(c, "sensor_id")
	startTime := h.parseDate(c, "start_date")
	endTime := h.parseDate(c, "end_date")
	events, err := h.us.Event.GetSensorHistory(c.Request.Context(), sensorID, startTime, endTime)
	h.handleError(c, err, http.StatusNotFound, ErrSensorNotFound)
	c.JSON(http.StatusOK, events)
}
