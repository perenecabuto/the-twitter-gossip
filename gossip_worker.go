package main

import (
	"log"
	"time"

	"github.com/dghubble/go-twitter/twitter"
)

type GossipPayload struct {
	Gossip string     `json:"gossip"`
	Events EventGroup `json:"events"`
}

type GossipWorker struct {
	stream *TwitterStream
	worker *TimeEventWorker
}

func NewGossipWorker(gossip *Gossip, gossipClassifiers []*GossipClassifier, eventChann chan *GossipPayload) *GossipWorker {
	log.Println("Listenning Gossip: ", gossip.Label)
	stream := NewTwitterStream(gossip.Subjects)

	classifiers := ConvertToMessageClassifiers(gossipClassifiers)
	classifierListener := NewMessageClassifierListener(classifiers)
	stream.AddListener(classifierListener)

	workerInterval := 10 * time.Second
	worker := NewTimeEventWorker(workerInterval)
	worker.SetOnEvent(func(t time.Time, events EventGroup) {
		log.Println("(", gossip.Label, ") time trigger: ", events)
		if len(events) > 0 {
			eventChann <- &GossipPayload{gossip.Label, events}
		}
	})

	classifierListener.SetOnMatch(func(label string, t *twitter.Tweet) {

		go worker.ReportEvent(label)
	})

	return &GossipWorker{stream, worker}
}

func (gw *GossipWorker) Start() {
	go gw.stream.Listen()
	gw.worker.Start()
}
