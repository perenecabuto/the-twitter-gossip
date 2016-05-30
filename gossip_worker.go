package main

import (
	"log"
	"time"

	"github.com/dghubble/go-twitter/twitter"
)

type GossipEventPayload struct {
	Gossip string     `json:"gossip"`
	Events EventGroup `json:"events"`
}

type GossipWorker struct {
	stream *TwitterStream
	worker *TimeEventWorker
}

func NewGossipWorker(gossip *Gossip, gossipClassifiers []*GossipClassifier, eventChann chan *GossipEventPayload) *GossipWorker {
	log.Println("Listenning Gossip: ", gossip.Label)
	stream := NewTwitterStream(gossip.Subjects)
	classifiers := ConvertMessageClassifiers(gossipClassifiers)
	classifierListener := NewMessageClassifierListener(classifiers)
	stream.AddListener(classifierListener)

	workerInterval := 10 * time.Second
	worker := NewTimeEventWorker(workerInterval)
	worker.SetOnEvent(func(t time.Time, events EventGroup) {
		log.Println("(", gossip.Label, ") time trigger: ", events)
		if len(events) > 0 {
			eventChann <- &GossipEventPayload{gossip.Label, events}
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

func ConvertMessageClassifiers(gclassifiers []*GossipClassifier) []*MessageClassifier {
	classifiers := []*MessageClassifier{}
	for _, gclassifier := range gclassifiers {
		newClassifier := NewMessageClassifier(gclassifier.Label, gclassifier.Patterns)
		classifiers = append(classifiers, newClassifier)
	}
	return classifiers
}
