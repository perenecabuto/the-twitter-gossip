package main

import (
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

func main() {
	clientMessage := make(chan *GossipEventPayload)
	//service := &DummyGossipService{}
	service := NewMongoGossipService()

	gossips, _ := service.FindAllGossip()
	for _, gossip := range gossips {
		gossipClassifiers, _ := service.FindClassifiersByGossip(gossip)
		worker := NewGossipWorker(gossip, gossipClassifiers, clientMessage)
		go worker.Start()
		log.Println("(", gossip.Label, ")", "worker started")
	}

	http.Handle("/events", websocket.Handler(func(ws *websocket.Conn) {
		for payload := range clientMessage {
			log.Println("Sending ", payload)
			if err := websocket.JSON.Send(ws, payload); err != nil {
				log.Println(err)
				ws.Close()
				return
			}
		}
	}))

	http.Handle("/gossip/", &GossipResourceHandler{service})

	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
