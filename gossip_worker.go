package main

import (
	"log"
	"time"
)

type GossipEventPayload struct {
	Gossip string     `json:"gossip"`
	Events EventGroup `json:"events"`
}

type GossipWorker struct {
	gossip     *Gossip
	stream     *TwitterStream
	worker     *TimedLabelCounter
	listener   *MessageClassifierListener
	eventChann chan *GossipEventPayload
}

func NewGossipWorker(gossip *Gossip, gossipClassifiers []*GossipClassifier, eventChann chan *GossipEventPayload) *GossipWorker {
	log.Println("Listenning Gossip: ", gossip.Label)
	stream := NewTwitterStream(gossip.Subjects)
	classifiers := ConvertMessageClassifiers(gossipClassifiers)
	listener := NewMessageClassifierListener(classifiers)
	stream.AddListener(listener)

	workerInterval := 10 * time.Second
	worker := NewTimedLabelCounter(workerInterval)

	return &GossipWorker{gossip, stream, worker, listener, eventChann}
}

func (gw *GossipWorker) Start() {
	go gw.stream.Listen()
	go gw.worker.Start()
	gw.run()
}

func (gw *GossipWorker) run() {
	for {
		select {
		case events := <-gw.worker.OnTimeChann:
			log.Println("Gossip:", gw.gossip.Label, "Events:", events)
			if len(events) > 0 {
				gw.eventChann <- &GossipEventPayload{gw.gossip.Label, events}
			}
		case label := <-gw.listener.OnMatchChann:
			go gw.worker.ReportEvent(label)
		}
	}
}

func ConvertMessageClassifiers(gclassifiers []*GossipClassifier) []*MessageClassifier {
	classifiers := []*MessageClassifier{}
	for _, gclassifier := range gclassifiers {
		newClassifier := NewMessageClassifier(gclassifier.Label, gclassifier.Patterns)
		classifiers = append(classifiers, newClassifier)
	}
	return classifiers
}
