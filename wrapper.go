package gorillawswrapper

import (
	"sync"
	"time"

	"github.com/shovon/go-stoppable"

	"github.com/gorilla/websocket"
)

const (
	writeWait = 60 * time.Second

	pongWait = 60 * time.Second

	pingPeriod = (pongWait * 9) / 10
)

type Message struct {
	MessageType int
	Message     []byte
}

// A wrapper for Gorilla's WebSocket
type Wrapper struct {
	stoppable.Stoppable
	writeMut *sync.Mutex
	readMut  *sync.Mutex
	c        *websocket.Conn
	messages chan Message
}

func NewWrapper(c *websocket.Conn) Wrapper {
	wrapper := Wrapper{stoppable.NewStoppable(), &sync.Mutex{}, &sync.Mutex{}, c, make(chan Message)}

	go wrapper.pingLoop()
	go wrapper.readLoop()

	return wrapper
}

func (w *Wrapper) pingLoop() {
	w.c.SetReadDeadline(time.Now().Add(pongWait))
	w.c.SetPongHandler(func(string) error {
		w.c.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

loop:
	for {
		select {
		case <-time.After(pingPeriod):
			w.c.SetWriteDeadline(time.Now().Add(writeWait))
			w.WriteMessage(websocket.PingMessage, nil)
		case <-w.OnStopped():
			break loop
		}
	}
}

func (w *Wrapper) readLoop() {
	defer close(w.messages)
	for {
		w.readMut.Lock()
		t, message, err := w.c.ReadMessage()
		w.c.SetReadDeadline(time.Now().Add(pongWait))
		w.readMut.Unlock()
		if err != nil {
			w.Stop()
			return
		}
		w.messages <- Message{t, message}
	}
}

func (w *Wrapper) setWriteDeadline() error {
	return w.c.SetWriteDeadline(time.Now().Add(writeWait))
}

func (w *Wrapper) WriteMessage(messageType int, data []byte) error {
	w.writeMut.Lock()
	defer w.writeMut.Unlock()
	err := w.setWriteDeadline()
	if err != nil {
		w.Stop()
		return err
	}
	err = w.c.WriteMessage(messageType, data)
	if err != nil {
		w.Stop()
	}
	return err
}

func (w *Wrapper) WriteTextMessage(message string) error {
	return w.WriteMessage(websocket.TextMessage, []byte(message))
}

func (w *Wrapper) WriteBinaryMessage(data []byte) error {
	return w.WriteMessage(websocket.BinaryMessage, data)
}

func (w *Wrapper) WriteJSON(v interface{}) error {
	w.writeMut.Lock()
	defer w.writeMut.Unlock()
	err := w.setWriteDeadline()
	if err != nil {
		w.Stop()
		return err
	}
	err = w.c.WriteJSON(v)
	if err != nil {
		w.Stop()
	}
	return err
}

func (w *Wrapper) MessagesChannel() <-chan Message {
	return w.messages
}
