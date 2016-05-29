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
	events := make(chan string, 1024)
	service := &DummyGossipService{}
	gossip := service.FindGossipByLabel("test 1")
	gossipClassifiers := service.FindClassifiersByGossip(gossip)
	go NewGossipRadarWorker(gossip, gossipClassifiers, events)

	gossip2 := service.FindGossipByLabel("test 2")
	gossipClassifiers2 := service.FindClassifiersByGossip(gossip2)
	go NewGossipRadarWorker(gossip2, gossipClassifiers2, events)

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
		msg, err := json.Marshal(events)
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
