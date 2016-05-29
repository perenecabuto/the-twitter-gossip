package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"golang.org/x/net/websocket"
)

func main() {
	events := make(chan string)

	var service GossipService = &DummyGossipService{}
	for _, gossip := range service.FindAllGossip() {
		gossipClassifiers := service.FindClassifiersByGossip(gossip)
		go NewGossipRadarWorker(gossip, gossipClassifiers, events)
	}

	http.Handle("/events", websocket.Handler(func(ws *websocket.Conn) {
		for msg := range events {
			if _, err := ws.Write([]byte(msg)); err != nil {
				log.Println(err)
				return
			}
		}
	}))

	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

func NewGossipRadarWorker(gossip *Gossip, gossipClassifiers []*GossipClassifier, eventsChann chan<- string) {
	fmt.Println("Listenning Gossip: ", gossip.Label)
	stream := NewTwitterStream(gossip.Subjects)

	classifiers := ConvertToMessageClassifiers(gossipClassifiers)
	classifierListener := NewMessageClassifierListener(classifiers)
	stream.AddListener(classifierListener)

	//printer := NewTweetPrinter()
	//stream.AddListener(printer)

	// REPORTS
	reportInterval := 10 * time.Second
	report := NewTimeEventWorker(reportInterval)
	report.SetOnEvent(func(t time.Time, events EventGroup) {
		response := &map[string]interface{}{
			"gossip": gossip.Label,
			"events": events,
		}
		msg, err := json.Marshal(response)
		if err != nil {
			log.Println(err)
			return
		}

		eventsChann <- string(msg)
		fmt.Printf("Sent: %s\n", msg)
	})

	classifierListener.SetOnMatch(func(label string, t *twitter.Tweet) {
		go report.ReportEvent(label)
	})

	go report.Start()
	go stream.Listen()
}
