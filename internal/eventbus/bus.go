package eventbus

import (
	"sync"
	"time"
)

// Event is the unit that flows through the internal runtime bus.
type Event struct {
	Name       string    `json:"name"`
	OccurredAt time.Time `json:"occurredAt"`
	Payload    any       `json:"payload,omitempty"`
}

// Subscriber receives published events.
type Subscriber chan Event

// Bus is a small in-memory event bus used by the Phase 1 prototype.
type Bus struct {
	mu          sync.RWMutex
	nextID      int
	subscribers map[int]Subscriber
}

// New creates a new event bus.
func New() *Bus {
	return &Bus{
		subscribers: make(map[int]Subscriber),
	}
}

// Publish fan-outs an event to all subscribers without blocking the caller.
func (b *Bus) Publish(event Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, subscriber := range b.subscribers {
		select {
		case subscriber <- event:
		default:
		}
	}
}

// Subscribe registers a new subscriber.
func (b *Bus) Subscribe(buffer int) (int, Subscriber) {
	b.mu.Lock()
	defer b.mu.Unlock()

	id := b.nextID
	b.nextID++

	subscriber := make(Subscriber, buffer)
	b.subscribers[id] = subscriber

	return id, subscriber
}

// Unsubscribe removes a subscriber and closes its channel.
func (b *Bus) Unsubscribe(id int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	subscriber, ok := b.subscribers[id]
	if !ok {
		return
	}

	delete(b.subscribers, id)
	close(subscriber)
}
