package main

import "log"

type GossipWorkerPool struct {
	pool       map[string]*GossipWorker
	EventChann chan *GossipEventPayload
}

func NewGossipWorkerPool() *GossipWorkerPool {
	return &GossipWorkerPool{make(map[string]*GossipWorker), make(chan *GossipEventPayload)}
}

func (gwp GossipWorkerPool) BuildWorker(id string, g *Gossip, c []*GossipClassifier) {
	gwp.AddWorker(id, NewGossipWorker(g, c))
}

func (gwp GossipWorkerPool) AddWorker(id string, w *GossipWorker) {
	if _, ok := gwp.pool[id]; ok {
		gwp.RemoveWorker(id)
	}
	gwp.pool[id] = w
	go gwp.listenTo(w)
}

func (gwp GossipWorkerPool) StartWorker(id string) {
	if w, ok := gwp.pool[id]; ok {
		go w.Start()
	}
}

func (gwp GossipWorkerPool) listenTo(w *GossipWorker) {
	for event := range w.EventChann {
		gwp.EventChann <- event
	}
}

func (gwp GossipWorkerPool) StopWorker(id string) {
	log.Println("Stop worker", id)
	if w, ok := gwp.pool[id]; ok {
		w.Stop()
	}
}

func (gwp *GossipWorkerPool) RemoveWorker(id string) {
	if w, ok := gwp.pool[id]; ok {
		go w.Stop()
		close(w.EventChann)
		delete(gwp.pool, id)
	}
}

func (gwp GossipWorkerPool) StartAll() {
	for _, w := range gwp.pool {
		go w.Start()
	}
}

func (gwp GossipWorkerPool) StopAll() {
	log.Println("Stop All (", len(gwp.pool), ")")
	for id, _ := range gwp.pool {
		gwp.StopWorker(id)
	}
}
