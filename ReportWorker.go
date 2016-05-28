package main

import (
	"fmt"
	"time"
)

type ReportWorker struct {
	events map[string]int
}

func NewReportWorker() *ReportWorker {
	return &ReportWorker{map[string]int{}}
}

func (rw *ReportWorker) Start() {
	ticker := time.NewTicker(5 * time.Second)
	for t := range ticker.C {
		fmt.Println(t, rw.events)
	}
}

func (rw *ReportWorker) ReportEvent(name string) {
	val, ok := rw.events[name]
	if !ok {
		val = 0
	}

	rw.events[name] = val + 1
}
