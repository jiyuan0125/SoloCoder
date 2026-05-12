package ws

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID   string
	Conn *websocket.Conn
	Send chan []byte
}

type Manager struct {
	clients    sync.Map
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
}

func NewManager() *Manager {
	return &Manager{
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (m *Manager) Run() {
	for {
		select {
		case client := <-m.Register:
			m.clients.Store(client.ID, client)
		case client := <-m.Unregister:
			if _, ok := m.clients.LoadAndDelete(client.ID); ok {
				close(client.Send)
			}
		case message := <-m.Broadcast:
			m.clients.Range(func(key, value interface{}) bool {
				client := value.(*Client)
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					m.clients.Delete(client.ID)
				}
				return true
			})
		}
	}
}

func (c *Client) Read(manager *Manager) {
	defer func() {
		manager.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("ws read error: %v", err)
			break
		}
	}
}

func (c *Client) Write() {
	defer c.Conn.Close()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.Conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func BroadcastCallNumber(manager *Manager, serialNum string, doctorName string) error {
	msg := map[string]interface{}{
		"type":        "call_number",
		"serial_num":  serialNum,
		"doctor_name": doctorName,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	manager.Broadcast <- data
	return nil
}
