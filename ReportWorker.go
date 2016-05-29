package main

import (
	"fmt"
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

type ReportWorker struct {
	Interval    time.Duration
	events      EventGroup
	OnTimeEvent func(t time.Time, events EventGroup)
	reportChann chan string
}

func NewReportWorker(interval time.Duration) *ReportWorker {
	return &ReportWorker{interval, EventGroup{}, nil, make(chan string)}
}

func (rw *ReportWorker) Start() {
	ticker := time.NewTicker(rw.Interval)
	for {
		select {
		case t := <-ticker.C:
			fmt.Println(t, rw.OnTimeEvent != nil)
			if rw.OnTimeEvent != nil {
				go rw.OnTimeEvent(t, rw.events.Clone())
			}

			for key, _ := range rw.events {
				rw.events[key] = 0
			}
		case name := <-rw.reportChann:
			val, ok := rw.events[name]
			if !ok {
				val = 0
			}

			rw.events[name] = val + 1
			fmt.Println("Event", name, rw.events[name])
		}
	}
}

func (rw *ReportWorker) ReportEvent(name string) {
	rw.reportChann <- name
}
