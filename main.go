package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/websocket"
)

func main() {
	//service := &DummyGossipService{}
	wsClients := NewWSConnections()
	service := NewMongoGossipService()
	workerPool := NewGossipWorkerPool()
	workerPool.OnEvent(func(eg *GossipEventGroup) {
		g, err := service.FindGossipByLabel(eg.Gossip)
		if err == nil {
			for label, value := range eg.EventGroup.Events {
				gEvent := &GossipClassifierEvent{Label: label, Value: value, GossipId: g.ID, Timestamp: time.Now()}
				service.SaveEvent(gEvent)
			}
		} else {
			log.Println(err)
		}
		wsClients.BroadcastChann <- EventGroupPayload{eg.Gossip, eg.EventGroup.Time.Unix(), eg.EventGroup.Events}
	})

	gossips, _ := service.FindAllGossip()
	for _, gossip := range gossips {
		classifiers, _ := service.FindClassifiersByGossip(gossip)
		workerPool.BuildWorker(WorkerID(gossip.Label), gossip, classifiers)
	}

	http.Handle("/gossip/", CorsMiddleware(NewGossipResourceHandler(service, workerPool)))
	http.Handle("/events", websocket.Handler(func(ws *websocket.Conn) {
		log.Println("New WS connection")
		defer wsClients.Remove(ws)
		wsClients.Add(ws)
		io.Copy(ws, ws)
	}))

	go StopAllWorkersAtExit(workerPool)
	go workerPool.StartAll()
	go wsClients.ListenBroadcasts()
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "content-type")
		next.ServeHTTP(w, r)
	})
}

func StopAllWorkersAtExit(p *GossipWorkerPool) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		p.StopAll()
		os.Exit(1)
	}()
}
