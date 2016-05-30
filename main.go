package main

import (
	"encoding/json"
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

func main() {
	clientMessage := make(chan *GossipPayload)
	service := &DummyGossipService{}

	for _, gossip := range service.FindAllGossip() {
		gossipClassifiers := service.FindClassifiersByGossip(gossip)
		worker := NewGossipWorker(gossip, gossipClassifiers, clientMessage)
		go worker.Start()
		log.Println("(", gossip.Label, ")", "worker started")
	}

	http.Handle("/events", websocket.Handler(func(ws *websocket.Conn) {
		for payload := range clientMessage {
			msg, _ := json.Marshal(payload)
			log.Println("Sending ", string(msg))
			if _, err := ws.Write([]byte(msg)); err != nil {
				log.Println(err)
				ws.Close()
				return
			}
		}
	}))

	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
