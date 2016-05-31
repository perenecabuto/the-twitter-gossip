package main

import "time"

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
	OnTimeChann chan EventGroup
}

func NewTimedLabelCounter(interval time.Duration) *TimedLabelCounter {
	return &TimedLabelCounter{interval: interval, events: EventGroup{},
		reportChann: make(chan string), OnTimeChann: make(chan EventGroup)}
}

func (rw *TimedLabelCounter) ReportEvent(name string) {
	rw.reportChann <- name
}

func (rw *TimedLabelCounter) Start() {
	ticker := time.NewTicker(rw.interval)
	for {
		select {
		case <-ticker.C:
			rw.OnTimeChann <- rw.events.Clone()
			for key, _ := range rw.events {
				rw.events[key] = 0
			}
		case name := <-rw.reportChann:
			val, ok := rw.events[name]
			if !ok {
				val = 0
			}

			rw.events[name] = val + 1
		}
	}
}
