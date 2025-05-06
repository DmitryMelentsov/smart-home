package http

import (
	"homework/internal/domain"
	"sync"

	"github.com/coder/websocket"
)

const buffer int = 10

type EventBroker struct {
	subscriptions map[*websocket.Conn]chan *domain.Event
	ids           map[int64][]*websocket.Conn
	mu            sync.RWMutex
}

func NewEventBroker() *EventBroker {
	return &EventBroker{
		ids:           make(map[int64][]*websocket.Conn),
		subscriptions: make(map[*websocket.Conn]chan *domain.Event),
	}
}

func (b *EventBroker) Subscribe(conn *websocket.Conn, sensorID int64) chan *domain.Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	if ch, ok := b.subscriptions[conn]; ok {
		return ch
	}

	ch := make(chan *domain.Event, buffer)
	b.ids[sensorID] = append(b.ids[sensorID], conn)
	b.subscriptions[conn] = ch

	return ch
}

func (b *EventBroker) Unsubscribe(conn *websocket.Conn) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if ch, ok := b.subscriptions[conn]; ok {
		close(ch)
		delete(b.subscriptions, conn)
	}
}

func (b *EventBroker) Publish(sensorID int64, event *domain.Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, conn := range b.ids[sensorID] {
		if ch, ok := b.subscriptions[conn]; ok {
			select {
			case ch <- event:
			default:
			}
		}
	}
}

func (b *EventBroker) close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	for sensorID, ch := range b.subscriptions {
		close(ch)
		delete(b.subscriptions, sensorID)
	}
	for conn := range b.ids {
		delete(b.ids, conn)
	}
}
