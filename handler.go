package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type GossipResourceHandler struct {
	service GossipService
	pool    *GossipWorkerPool
}

func NewGossipResourceHandler(s GossipService, p *GossipWorkerPool) *GossipResourceHandler {
	return &GossipResourceHandler{s, p}
}

func (h *GossipResourceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	splittedPath := strings.SplitN(r.URL.Path, "/", 4)
	gossip := splittedPath[2]
	action := ""
	if len(splittedPath) > 3 {
		action = splittedPath[3]
	}

	log.Println("HANDLE", r.URL.Path, r.Method, splittedPath, gossip, action)
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "GET" {
		if len(gossip) == 0 {
			h.List(w, r)
		} else if action == "start" {
			h.StartWorker(gossip, w, r)
		} else if action == "stop" {
			h.StopWorker(gossip, w, r)
		} else if action == "history" {
			h.EventsHistory(gossip, w, r)
		} else {
			h.Get(gossip, w, r)
		}
	} else if r.Method == "POST" {
		h.Create(w, r)
	} else if r.Method == "PUT" {
		h.Update(gossip, w, r)
	} else if r.Method == "DELETE" {
		h.Delete(gossip, w, r)
	}
}

func (h *GossipResourceHandler) List(w http.ResponseWriter, r *http.Request) {
	payload := &GossipListPayload{[]*GossipPayload{}}
	gossips, err := h.service.FindAllGossip()
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}
	for _, g := range gossips {
		payload.Gossips = append(payload.Gossips, h.buildGossipPayload(g))
	}
	response, _ := json.Marshal(payload)
	w.Write([]byte(response))
}

func (h *GossipResourceHandler) Get(gossipLabel string, w http.ResponseWriter, r *http.Request) {
	gossip, err := h.service.FindGossipByLabel(gossipLabel)
	if gossip == nil {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}
	payload := h.buildGossipPayload(gossip)
	response, _ := json.Marshal(payload)
	w.Write([]byte(response))
}

func (h *GossipResourceHandler) Create(w http.ResponseWriter, r *http.Request) {
	payload := &GossipPayload{}
	if err := json.NewDecoder(r.Body).Decode(payload); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}
	gossip, classifiers := payload.ToModel()
	if err := h.service.CreateGossip(gossip, classifiers); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}

	h.pool.BuildWorker(WorkerID(gossip.Label), gossip, classifiers)
	h.StartWorker(gossip.Label, w, r)
}

func (h *GossipResourceHandler) Update(gossipLabel string, w http.ResponseWriter, r *http.Request) {
	payload := &GossipPayload{}
	if err := json.NewDecoder(r.Body).Decode(payload); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}
	gossip, classifiers := payload.ToModel()
	if err := h.service.UpdateGossip(gossipLabel, gossip, classifiers); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}

	wasRunning := h.pool.WorkerState(WorkerID(gossip.Label)) == STARTED
	h.pool.BuildWorker(WorkerID(gossip.Label), gossip, classifiers)

	if wasRunning {
		h.StartWorker(gossip.Label, w, r)
	} else {
		h.Get(gossip.Label, w, r)
	}
}

func (h *GossipResourceHandler) Delete(gossipLabel string, w http.ResponseWriter, r *http.Request) {
	h.pool.StopWorker(WorkerID(gossipLabel))
	if err := h.service.RemoveGossip(gossipLabel); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte("{\"ok\": true}"))
}

func (h *GossipResourceHandler) StartWorker(gossipLabel string, w http.ResponseWriter, r *http.Request) {
	go h.pool.StartWorker(WorkerID(gossipLabel))
	h.Get(gossipLabel, w, r)
}

func (h *GossipResourceHandler) StopWorker(gossipLabel string, w http.ResponseWriter, r *http.Request) {
	go h.pool.StopWorker(WorkerID(gossipLabel))
	h.Get(gossipLabel, w, r)
}

func (h *GossipResourceHandler) EventsHistory(gossipLabel string, w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	startUnixTime, err := strconv.ParseInt(params.Get("from"), 10, 64)
	if err != nil {
		startUnixTime = time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	}
	endUnixTime, err := strconv.ParseInt(params.Get("to"), 10, 64)
	if err != nil {
		endUnixTime = time.Now().Unix()
	}

	from := time.Unix(startUnixTime, 0)
	to := time.Unix(endUnixTime, 0)
	limit, err := strconv.Atoi(params.Get("limit"))
	if err != nil {
		limit = 30
	}
	events, err := h.service.FindClassifiersEvents(gossipLabel, from, to, limit)
	if err != nil {
		log.Println(err.Error())
		http.NotFound(w, r)
		return
	}

	payload := NewGossipEventHistoryPayloadFromModel(gossipLabel, events)
	response, err := json.Marshal(payload)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte(response))
}

func (h *GossipResourceHandler) buildGossipPayload(g *Gossip) *GossipPayload {
	classifiers, err := h.service.FindClassifiersByGossip(g)
	if err != nil {
		log.Println(err)
		return nil
	}

	state := string(h.pool.WorkerState(WorkerID(g.Label)))
	return NewGossipPayloadFromModel(g, classifiers, state)
}
