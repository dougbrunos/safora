package api

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Broker struct {
	mu      sync.Mutex
	clients map[chan string]struct{}
}

func NewBroker() *Broker {
	return &Broker{clients: make(map[chan string]struct{})}
}

func (b *Broker) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	flusher, ok := rw.(http.Flusher)
	if !ok {
		http.Error(rw, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")

	ch := make(chan string, 10)
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()
	defer func() {
		b.mu.Lock()
		delete(b.clients, ch)
		b.mu.Unlock()
	}()

	// A comment line every 15s keeps idle connections alive (long copies can be silent).
	keepAlive := time.NewTicker(15 * time.Second)
	defer keepAlive.Stop()

	for {
		select {
		case <-keepAlive.C:
			fmt.Fprint(rw, ": ping\n\n")
			flusher.Flush()
		case msg := <-ch:
			fmt.Fprintf(rw, "data: %s\n\n", msg)
			flusher.Flush()
		case <-req.Context().Done():
			return
		}
	}
}

// Broadcast drops the event for clients whose buffer is full.
func (b *Broker) Broadcast(msg string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.clients {
		select {
		case ch <- msg:
		default:
		}
	}
}
