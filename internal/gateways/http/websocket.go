package http

import (
	"encoding/json"
	"log"
	"time"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
)

type WebSocketHandler struct {
	useCases UseCases
	eb       *EventBroker
	close    chan struct{}
}

func NewWebSocketHandler(useCases UseCases) *WebSocketHandler {
	return &WebSocketHandler{
		useCases: useCases,
		close:    make(chan struct{}),
		eb:       NewEventBroker(),
	}
}

func (h *WebSocketHandler) Handle(c *gin.Context, id int64) error {
	conn, err := websocket.Accept(c.Writer, c.Request, nil)
	if err != nil {
		return err
	}
	ctx := conn.CloseRead(c)
	eventChan := h.eb.Subscribe(conn, id)

	go func() {
		defer func() {
			h.eb.Unsubscribe(conn)
			err := conn.Close(websocket.StatusNormalClosure, "Closed")
			if err != nil {
				return
			}
		}()
		timer := time.After(200 * time.Millisecond)
		for {
			select {
			case <-timer:
				event, err := h.useCases.Event.GetLastEventBySensorID(c, id)
				if errorHandler("Error getting last event by sensor id", err) {
					continue
				}
				h.eb.Publish(id, event)
			case event := <-eventChan:
				msg, err := json.Marshal(event)
				if errorHandler("Error marshaling event", err) {
					continue
				}
				err = conn.Write(ctx, websocket.MessageText, msg)
				if errorHandler("Error writing message", err) {
					continue
				}
			case <-c.Done():
				return
			case <-h.close:
				return
			}
		}
	}()

	return nil
}

func (h *WebSocketHandler) Shutdown() error {
	close(h.close)
	h.eb.close()
	return nil
}

func errorHandler(message string, err error) bool {
	if err != nil {
		log.Printf("%s: %v", message, err)
		return true
	}
	return false
}
