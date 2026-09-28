package api

import (
	"fmt"
	"net/http"
)

type Broker struct {
	Notifier       chan string
	newClients     chan chan string
	closingClients chan chan string
	clients        map[chan string]bool
}

func NewBroker() *Broker {
	return &Broker{
		Notifier:       make(chan string, 1),
		newClients:     make(chan chan string),
		closingClients: make(chan chan string),
		clients:        make(map[chan string]bool),
	}
}

func (broker *Broker) Start() {
	for {
		select {
		case s := <-broker.newClients:
			broker.clients[s] = true
		case s := <-broker.closingClients:
			delete(broker.clients, s)
		case event := <-broker.Notifier:
			for clientMessageChan := range broker.clients {
				select {
				case clientMessageChan <- event:
				default:
					// Cannot send, drop event for this client
				}
			}
		}
	}
}

func (broker *Broker) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	flusher, ok := rw.(http.Flusher)
	if !ok {
		http.Error(rw, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Access-Control-Allow-Origin", "*")

	messageChan := make(chan string, 10)
	broker.newClients <- messageChan

	defer func() {
		broker.closingClients <- messageChan
	}()

	notify := req.Context().Done()

	go func() {
		<-notify
		broker.closingClients <- messageChan
	}()

	for {
		msg, ok := <-messageChan
		if !ok {
			break
		}
		fmt.Fprintf(rw, "data: %s\n\n", msg)
		flusher.Flush()
	}
}

func (broker *Broker) Broadcast(msg string) {
	broker.Notifier <- msg
}
