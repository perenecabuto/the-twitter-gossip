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
	service := &DummyGossipService{}
	gossip := service.FindGossipByLabel("test")

	stream := NewTwitterStream(gossip.Subjects)

	//printer := NewTweetPrinter()
	//stream.AddListener(printer)

	gossipClassifiers := service.FindClassifiersByGossip(gossip)
	classifiers := ConvertToMessageClassifiers(gossipClassifiers)
	classifierListener := NewMessageClassifierListener(classifiers)
	stream.AddListener(classifierListener)

	// REPORTS
	eventsChann := make(chan string)
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

	http.Handle("/events", websocket.Handler(func(ws *websocket.Conn) {
		for msg := range eventsChann {
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
