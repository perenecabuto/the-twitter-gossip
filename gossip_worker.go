package main

import (
	"log"
	"time"
)

const DEFAULT_INTERVAL = 1 * time.Second

type GossipEventPayload struct {
	Gossip string     `json:"gossip"`
	Events EventGroup `json:"events"`
}

type GossipWorkerState string

const (
	STARTED  GossipWorkerState = "STARTED"
	STARTING                   = "STARTING"
	STOPPING                   = "STOPPING"
	STOPPED                    = "STOPPED"
)

type GossipWorker struct {
	gossip      *Gossip
	stream      *TwitterStream
	timeCounter *TimedLabelCounter
	listener    *MessageClassifierListener
	State       GossipWorkerState
	stopChann   chan bool
	startChann  chan bool
	EventChann  chan *GossipEventPayload
}

func NewGossipWorker(gossip *Gossip, gossipClassifiers []*GossipClassifier) *GossipWorker {
	log.Println("Listenning Gossip: ", gossip.Label)
	stream := NewTwitterStream(gossip.Subjects)
	classifiers := ConvertMessageClassifiers(gossipClassifiers)
	timeCounter := NewTimedLabelCounter(DEFAULT_INTERVAL)
	listener := NewMessageClassifierListener(classifiers)
	stream.AddListener(listener)

	return &GossipWorker{gossip, stream, timeCounter, listener, STOPPED, make(chan bool), make(chan bool), make(chan *GossipEventPayload)}
}

func (gw *GossipWorker) Start() {
	if gw.State == STOPPED {
		gw.State = STARTING
		go gw.run()
		gw.startChann <- true
		<-gw.startChann
		gw.State = STARTED
	}
}

func (gw *GossipWorker) Stop() {
	if gw.State == STARTED {
		gw.State = STOPPING
		gw.stopChann <- true
		<-gw.stopChann
		gw.State = STOPPED
	}
}

func (gw *GossipWorker) run() {
	for {
		select {
		case <-gw.startChann:
			log.Println("! Starting GossipWorker:", gw.gossip.Label)
			go gw.stream.Listen()
			go gw.timeCounter.Start()
			gw.startChann <- true

		case events := <-gw.timeCounter.OnTimeChann:
			log.Println("Gossip:", gw.gossip.Label, "Events:", events)
			if len(events) > 0 {
				gw.EventChann <- &GossipEventPayload{gw.gossip.Label, events}
			}

		case label := <-gw.listener.OnMatchChann:
			go gw.timeCounter.ReportEvent(label)

		case <-gw.stopChann:
			log.Println("! Stopping GossipWorker:", gw.gossip.Label)
			gw.timeCounter.Stop()
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
