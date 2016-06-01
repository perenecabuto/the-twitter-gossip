package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type EventGroupPayload struct {
	Gossip    string         `json:"gossip"`
	Timestamp int64          `json:"timestamp"`
	Events    map[string]int `json:"events"`
}

type GossipEventHistoryPayload struct {
	Gossip  string               `json:"gossip"`
	History []*EventGroupPayload `json:"history"`
}

type GossipPayload struct {
	Gossip      string              `json:"gossip"`
	Subjects    []string            `json:"subjects"`
	Classifiers map[string][]string `json:"classifiers"`
	WorkerState string              `json:"state"`
}

type GossipListPayload struct {
	Gossips []*GossipPayload `json:"gossips"`
}

func (p *GossipEventHistoryPayload) FromModelList(gossipLabel string, list []*GossipClassifierEvent) {
	history := []*EventGroupPayload{}
	var eventGroup *EventGroupPayload
	for _, e := range list {
		if eventGroup == nil {
			eventGroup = &EventGroupPayload{gossipLabel, e.Timestamp.Unix(), map[string]int{}}
		}
		if e.Timestamp.Unix() != eventGroup.Timestamp {
			history = append(history, eventGroup)
			eventGroup = nil
			continue
		}
		eventGroup.Events[e.Label] = e.Value
	}
	p.History = history
}

func (p *GossipPayload) ToModel() (*Gossip, []*GossipClassifier) {
	gossip := &Gossip{Label: p.Gossip, Subjects: p.Subjects}
	classifiers := []*GossipClassifier{}
	for label, patterns := range p.Classifiers {
		classifiers = append(classifiers, &GossipClassifier{Label: label, Patterns: patterns})
	}
	return gossip, classifiers
}

type GossipResourceHandler struct {
	service GossipService
	pool    *GossipWorkerPool
}

func NewGossipResourceHandler(s GossipService, p *GossipWorkerPool) *GossipResourceHandler {
	return &GossipResourceHandler{s, p}
}

func (h *GossipResourceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	splittedPath := strings.Split(r.URL.Path, "/")
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
			h.ClassifierHistory(gossip, w, r)
		} else {
			h.Get(gossip, w, r)
		}
	} else if r.Method == "POST" {
		h.Create(w, r)
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
	log.Println("GET", gossipLabel)
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

	log.Println("POST", payload)
	gossip, classifiers := payload.ToModel()
	if err := h.service.SaveGossip(gossip, classifiers); err != nil {
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

func (h *GossipResourceHandler) StartWorker(gossipLabel string, w http.ResponseWriter, r *http.Request) {
	go h.pool.StartWorker(WorkerID(gossipLabel))
	h.Get(gossipLabel, w, r)
}

func (h *GossipResourceHandler) StopWorker(gossipLabel string, w http.ResponseWriter, r *http.Request) {
	go h.pool.StopWorker(WorkerID(gossipLabel))
	h.Get(gossipLabel, w, r)
}

func (h *GossipResourceHandler) ClassifierHistory(gossipLabel string, w http.ResponseWriter, r *http.Request) {
	events, err := h.service.FindClassifiersEvents(gossipLabel)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	payload := &GossipEventHistoryPayload{gossipLabel, nil}
	payload.FromModelList(gossipLabel, events)

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

	cPayload := map[string][]string{}
	for _, c := range classifiers {
		cPayload[c.Label] = c.Patterns
	}

	state := string(h.pool.WorkerState(WorkerID(g.Label)))
	return &GossipPayload{g.Label, g.Subjects, cPayload, state}
}
