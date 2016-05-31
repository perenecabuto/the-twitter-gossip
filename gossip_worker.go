package main

import (
	"log"
	"time"
)

const DEFAULT_INTERVAL = 10 * time.Second

type GossipEventPayload struct {
	Gossip string     `json:"gossip"`
	Events EventGroup `json:"events"`
}

type GossipWorker struct {
	gossip     *Gossip
	stream     *TwitterStream
	worker     *TimedLabelCounter
	listener   *MessageClassifierListener
	stopChann  chan bool
	EventChann chan *GossipEventPayload
}

func NewGossipWorker(gossip *Gossip, gossipClassifiers []*GossipClassifier) *GossipWorker {
	log.Println("Listenning Gossip: ", gossip.Label)
	stream := NewTwitterStream(gossip.Subjects)
	classifiers := ConvertMessageClassifiers(gossipClassifiers)
	worker := NewTimedLabelCounter(DEFAULT_INTERVAL)
	listener := NewMessageClassifierListener(classifiers)
	stream.AddListener(listener)

	return &GossipWorker{gossip, stream, worker, listener, make(chan bool), make(chan *GossipEventPayload)}
}

func (gw *GossipWorker) Start() {
	go gw.stream.Listen()
	go gw.worker.Start()
	gw.run()
}

func (gw *GossipWorker) Stop() {
	gw.stopChann <- true
	<-gw.stopChann
}

func (gw *GossipWorker) run() {
	log.Println(gw.gossip.Label+":", "worker started")
	for {
		select {
		case events := <-gw.worker.OnTimeChann:
			log.Println("Gossip:", gw.gossip.Label, "Events:", events)
			if len(events) > 0 {
				gw.EventChann <- &GossipEventPayload{gw.gossip.Label, events}
			}
		case label := <-gw.listener.OnMatchChann:
			go gw.worker.ReportEvent(label)
		case <-gw.stopChann:
			log.Println("! Stopping GossipWorker:", gw.gossip.Label)
			gw.worker.Stop()
			gw.stream.Stop()
			gw.stopChann <- true
			return
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
