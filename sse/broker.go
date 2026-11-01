package sse

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"

	"github.com/coderhaqi/rhui"
)

// EventOptions specifies metadata for an SSE message.
type EventOptions struct {
	EventName string
	ID        string
}

// Client represents a connected SSE client stream.
type Client struct {
	id     string
	sendCh chan []byte
}

// Broker manages connected SSE clients and broadcasting.
type Broker struct {
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	mu         sync.RWMutex
}

// NewBroker creates an active SSE Broker.
func NewBroker() *Broker {
	b := &Broker{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
	}
	go b.run()
	return b
}

func (b *Broker) run() {
	for {
		select {
		case client := <-b.register:
			b.mu.Lock()
			b.clients[client.id] = client
			b.mu.Unlock()

		case client := <-b.unregister:
			b.mu.Lock()
			if _, ok := b.clients[client.id]; ok {
				delete(b.clients, client.id)
				close(client.sendCh)
			}
			b.mu.Unlock()

		case msg := <-b.broadcast:
			b.mu.RLock()
			for _, client := range b.clients {
				select {
				case client.sendCh <- msg:
				default:
					// Drop if client buffer is blocked
				}
			}
			b.mu.RUnlock()
		}
	}
}

// Stream wraps a client session for an HTTP connection.
type Stream struct {
	client *Client
	broker *Broker
	w      http.ResponseWriter
	flusher http.Flusher
}

// Connect initiates an SSE connection on the given http.ResponseWriter.
func (b *Broker) Connect(w http.ResponseWriter, r *http.Request) (*Stream, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("streaming unsupported by response writer")
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	client := &Client{
		id:     fmt.Sprintf("%p", r),
		sendCh: make(chan []byte, 16),
	}

	b.register <- client

	return &Stream{
		client:  client,
		broker:  b,
		w:       w,
		flusher: flusher,
	}, nil
}

// Wait blocks and flushes events to the client until disconnected.
func (s *Stream) Wait() {
	for msg := range s.client.sendCh {
		_, _ = s.w.Write(msg)
		s.flusher.Flush()
	}
}

// Close disconnects the client stream.
func (s *Stream) Close() {
	s.broker.unregister <- s.client
}

// BroadcastComponent renders a RHUI component and broadcasts it as an SSE event to all clients.
func (b *Broker) BroadcastComponent(engine *rhui.Engine, compName string, data any, opts EventOptions) error {
	var buf bytes.Buffer
	if err := engine.Render(&buf, compName, data); err != nil {
		return err
	}

	msg := formatSSE(opts.EventName, opts.ID, buf.String())
	b.broadcast <- []byte(msg)
	return nil
}

func formatSSE(event, id, data string) string {
	var sb bytes.Buffer
	if event != "" {
		sb.WriteString("event: " + event + "\n")
	}
	if id != "" {
		sb.WriteString("id: " + id + "\n")
	}
	// Multi-line data handling
	lines := bytes.Split([]byte(data), []byte("\n"))
	for _, line := range lines {
		sb.WriteString("data: ")
		sb.Write(line)
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
	return sb.String()
}
