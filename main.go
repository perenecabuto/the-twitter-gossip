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
	clientMessage := make(chan string)

	var service GossipService = &DummyGossipService{}
	for _, gossip := range service.FindAllGossip() {
		gossipClassifiers := service.FindClassifiersByGossip(gossip)
		worker := BuildWorkerForGossip(gossip, gossipClassifiers)
		worker.SetOnEvent(func(label string) func(t time.Time, events EventGroup) {
			return func(t time.Time, events EventGroup) {
				fmt.Println(t, "(", label, ") Sending events: ", events)
				response := &map[string]interface{}{
					"gossip": label,
					"events": events,
				}
				msg, err := json.Marshal(response)
				if err != nil {
					log.Println(err)
					return
				}

				clientMessage <- string(msg)
			}
		}(gossip.Label))
		fmt.Println("(", gossip.Label, ")", "worker started")
	}

	http.Handle("/events", websocket.Handler(func(ws *websocket.Conn) {
		for msg := range clientMessage {
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

func BuildWorkerForGossip(gossip *Gossip, gossipClassifiers []*GossipClassifier) *TimeEventWorker {
	fmt.Println("Listenning Gossip: ", gossip.Label)
	stream := NewTwitterStream(gossip.Subjects)

	classifiers := ConvertToMessageClassifiers(gossipClassifiers)
	classifierListener := NewMessageClassifierListener(classifiers)
	stream.AddListener(classifierListener)

	//printer := NewTweetPrinter()
	//stream.AddListener(printer)

	workerInterval := 10 * time.Second
	worker := NewTimeEventWorker(workerInterval)

	classifierListener.SetOnMatch(func(label string, t *twitter.Tweet) {
		//fmt.Println("(", gossip.Label, "), Matched: ", label)
		go worker.ReportEvent(label)
	})

	go stream.Listen()
	go worker.Start()

	return worker
}
