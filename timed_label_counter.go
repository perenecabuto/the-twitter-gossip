package main

import (
	"log"
	"time"
)

type EventGroup map[string]int

func (eg EventGroup) Clone() EventGroup {
	evcopy := EventGroup{}
	for k, v := range eg {
		evcopy[k] = v
	}
	return evcopy
}

type TimedLabelCounter struct {
	interval    time.Duration
	events      EventGroup
	reportChann chan string
	stopChann   chan bool
	OnTimeChann chan EventGroup
}

func NewTimedLabelCounter(interval time.Duration) *TimedLabelCounter {
	return &TimedLabelCounter{interval: interval, events: EventGroup{},
		reportChann: make(chan string), stopChann: make(chan bool),
		OnTimeChann: make(chan EventGroup)}
}

func (tlc *TimedLabelCounter) ReportEvent(name string) {
	tlc.reportChann <- name
}

func (tlc *TimedLabelCounter) Start() {
	ticker := time.NewTicker(tlc.interval)
	for {
		select {
		case <-ticker.C:
			tlc.OnTimeChann <- tlc.events.Clone()
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
