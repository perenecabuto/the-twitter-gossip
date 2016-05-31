package main

import (
	"io"
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

func main() {
	connections := NewWSConnections()
	//service := &DummyGossipService{}
	service := NewMongoGossipService()

	gossips, _ := service.FindAllGossip()
	for _, gossip := range gossips {
		gossipClassifiers, _ := service.FindClassifiersByGossip(gossip)
		worker := NewGossipWorker(gossip, gossipClassifiers, connections.BroadcastChann)
		go worker.Start()
		log.Println("(", gossip.Label, ")", "worker started")
	}

	http.Handle("/events", websocket.Handler(func(ws *websocket.Conn) {
		log.Println("New WS connection")
		defer connections.Remove(ws)
		connections.Add(ws)
		io.Copy(ws, ws)
	}))

	http.Handle("/gossip/", &GossipResourceHandler{service})

	go connections.ListenBroadcasts()
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
