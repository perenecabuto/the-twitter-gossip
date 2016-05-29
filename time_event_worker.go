package main

import "time"

type EventGroup map[string]int
type TimeEventCallback func(t time.Time, events EventGroup)

func (eg EventGroup) Clone() EventGroup {
	evcopy := EventGroup{}
	for k, v := range eg {
		evcopy[k] = v
	}
	return evcopy
}

type TimeEventWorker struct {
	interval    time.Duration
	events      EventGroup
	reportChann chan string
	onTimeEvent TimeEventCallback
}

func NewTimeEventWorker(interval time.Duration) *TimeEventWorker {
	return &TimeEventWorker{interval: interval, events: EventGroup{}, reportChann: make(chan string)}
}

func (rw *TimeEventWorker) SetOnEvent(callback TimeEventCallback) {
	rw.onTimeEvent = callback
}

func (rw *TimeEventWorker) ReportEvent(name string) {
	rw.reportChann <- name
}

func (rw *TimeEventWorker) Start() {
	ticker := time.NewTicker(rw.interval)
	for {
		select {
		case t := <-ticker.C:
			if rw.onTimeEvent != nil {
				go func() {
					rw.onTimeEvent(t, rw.events.Clone())

					for key, _ := range rw.events {
						rw.events[key] = 0
					}
				}()
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
