package main

import (
	"log"

	"golang.org/x/net/websocket"
)

type WSConnections struct {
	connections map[*websocket.Conn]bool
}

func NewWSConnections() *WSConnections {
	return &WSConnections{make(map[*websocket.Conn]bool)}
}

func (wsc *WSConnections) Add(ws *websocket.Conn) {
	wsc.connections[ws] = true
}

func (wsc *WSConnections) Remove(ws *websocket.Conn) {
	delete(wsc.connections, ws)
	ws.Close()
}

func (wsc *WSConnections) BroadcastJSON(payload *GossipEventPayload) {
	log.Println("BroadcastJSON to", len(wsc.connections), "connections")
	for ws := range wsc.connections {
		if err := websocket.JSON.Send(ws, payload); err != nil {
			log.Println(err)
			wsc.Remove(ws)
		}
	}
}

func (wsc *WSConnections) BroadcastJSONFromChann(chann chan *GossipEventPayload) {
	for payload := range chann {
		wsc.BroadcastJSON(payload)
	}
}
