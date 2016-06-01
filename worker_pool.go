package main

import "log"

type WorkerID string

type GossipWorkerPool struct {
	pool      map[WorkerID]*GossipWorker
	stopChann chan WorkerID
	callback  func(p *GossipEventPayload)
}

func NewGossipWorkerPool() *GossipWorkerPool {
	return &GossipWorkerPool{make(map[WorkerID]*GossipWorker), make(chan WorkerID), nil}
}

func (gwp *GossipWorkerPool) OnEvent(callback func(p *GossipEventPayload)) {
	gwp.callback = callback
}

func (gwp GossipWorkerPool) BuildWorker(id WorkerID, g *Gossip, c []*GossipClassifier) {
	gwp.AddWorker(id, NewGossipWorker(g, c))
}

func (gwp *GossipWorkerPool) AddWorker(id WorkerID, w *GossipWorker) {
	gwp.RemoveWorker(id)
	gwp.pool[id] = w
	go gwp.listenTo(id, w)
}

func (gwp GossipWorkerPool) StartWorker(id WorkerID) {
	if w, ok := gwp.pool[id]; ok {
		go w.Start()
	}
}

func (gwp GossipWorkerPool) listenTo(id WorkerID, w *GossipWorker) {
	for {
		select {
		case p, ok := <-w.EventChann:
			if ok && gwp.callback != nil {
				go gwp.callback(p)
			}
		case closeID := <-gwp.stopChann:
			if closeID == id {
				log.Println("Stoping POOL listener for:", id)
				return
			}
		}

	}

	log.Println("Closing listenTo", w)
}

func (gwp GossipWorkerPool) StopWorker(id WorkerID) {
	if w, ok := gwp.pool[id]; ok {
		gwp.stopChann <- id
		w.Stop()
	}
}

func (gwp *GossipWorkerPool) RemoveWorker(id WorkerID) {
	if w, ok := gwp.pool[id]; ok {
		gwp.StopWorker(id)
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
	log.Println("Stop All (", len(gwp.pool), ") workers")
	for id, _ := range gwp.pool {
		log.Println("")
		gwp.StopWorker(id)
	}
}
