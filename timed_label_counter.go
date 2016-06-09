package main

import (
	"log"
	"time"
)

type EventCount map[string]int

type EventGroup struct {
	Time   time.Time
	Events EventCount
}

func NewEventGroup(t time.Time, events EventCount) *EventGroup {
	eg := &EventGroup{t, make(EventCount)}
	for k, v := range events {
		eg.Events[k] = v
	}
	return eg
}

type TimedLabelCounter struct {
	events      EventCount
	reportChann chan string
	stopChann   chan bool
	OnTimeChann chan *EventGroup
}

func NewTimedLabelCounter(labels []string) *TimedLabelCounter {
	events := EventCount{}
	for _, l := range labels {
		events[l] = 0
	}
	return &TimedLabelCounter{events: events,
		reportChann: make(chan string), stopChann: make(chan bool),
		OnTimeChann: make(chan *EventGroup)}
}

func (tlc *TimedLabelCounter) ReportEvent(name string) {
	tlc.reportChann <- name
}

func (tlc *TimedLabelCounter) Start(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case t := <-ticker.C:
			tlc.OnTimeChann <- NewEventGroup(t, tlc.events)
			for key, _ := range tlc.events {
				tlc.events[key] = 0
			}
		case name := <-tlc.reportChann:
			val, ok := tlc.events[name]
			if !ok {
				val = 0
			}

			tlc.events[name] = val + 1
		case <-tlc.stopChann:
			log.Println("! Stopping TimedLabelCounter")
			tlc.stopChann <- true
			return
		}
	}
}

func (tlc *TimedLabelCounter) Stop() {
	tlc.stopChann <- true
	<-tlc.stopChann
}
