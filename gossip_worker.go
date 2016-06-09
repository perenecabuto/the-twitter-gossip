package main

import (
	"log"
	"time"
)

const DEFAULT_INTERVAL = 10 * time.Second

type GossipEventGroup struct {
	Gossip     string      `json:"gossip"`
	EventGroup *EventGroup `json:"events"`
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
	EventChann  chan *GossipEventGroup
}

func NewGossipWorker(gossip *Gossip, gossipClassifiers ClassifierList) *GossipWorker {
	log.Println("Listenning Gossip: ", gossip.Label)
	stream := NewTwitterStream(gossip.Subjects)
	classifiers := ConvertMessageClassifiers(gossipClassifiers)
	timeCounter := NewTimedLabelCounter(gossipClassifiers.Labels())
	listener := NewMessageClassifierListener(classifiers)
	stream.AddListener(listener)

	return &GossipWorker{gossip, stream, timeCounter, listener, STOPPED, make(chan bool), make(chan bool), make(chan *GossipEventGroup)}
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
			interval := gw.gossip.WorkerInterval
			if interval < 1 {
				interval = DEFAULT_INTERVAL
			}
			log.Println("! Starting GossipWorker:", gw.gossip.Label, " interval:", interval)
			go gw.stream.Listen()
			go gw.timeCounter.Start(interval)
			gw.startChann <- true

		case group := <-gw.timeCounter.OnTimeChann:
			if len(group.Events) > 0 {
				log.Println("Gossip:", gw.gossip.Label, "Events:", group)
				gw.EventChann <- &GossipEventGroup{gw.gossip.Label, group}
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
