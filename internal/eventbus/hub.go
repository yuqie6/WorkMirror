package eventbus

import (
	"context"
	"sync"
	"time"
)

type Event struct {
	Type      string         `json:"type"`
	Timestamp int64          `json:"timestamp"`
	Data      map[string]any `json:"data,omitempty"`
}

type Hub struct {
	mu   sync.RWMutex
	subs map[chan Event]struct{}
}

func NewHub() *Hub {
	return &Hub{subs: make(map[chan Event]struct{})}
}

func (h *Hub) Publish(evt Event) {
	if h == nil {
		return
	}
	if evt.Timestamp == 0 {
		evt.Timestamp = time.Now().UnixMilli()
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for ch := range h.subs {
		select {
		case ch <- evt:
		default:
			// 慢消费者直接丢弃，避免阻塞采集链路
		}
	}
}

func (h *Hub) Subscribe(ctx context.Context, buffer int) <-chan Event {
	if buffer <= 0 {
		buffer = 16
	}
	ch := make(chan Event, buffer)

	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()

	go func() {
		<-ctx.Done()
		h.mu.Lock()
		delete(h.subs, ch)
		h.mu.Unlock()
		close(ch)
	}()

	return ch
}
