package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type GossipPayload struct {
	Gossip      string              `json:"gossip"`
	Subjects    []string            `json:"subjects"`
	Classifiers map[string][]string `json:"classifiers"`
	WorkerState string              `json:"state"`
}

func (p *GossipPayload) ToModel() (*Gossip, []*GossipClassifier) {
	gossip := &Gossip{Label: p.Gossip, Subjects: p.Subjects}
	classifiers := []*GossipClassifier{}
	for label, patterns := range p.Classifiers {
		classifiers = append(classifiers, &GossipClassifier{Label: label, Patterns: patterns})
	}
	return gossip, classifiers
}

type GossipListPayload struct {
	Gossips []*GossipPayload `json:"gossip"`
}

type GossipResourceHandler struct {
	service GossipService
}

func (gah *GossipResourceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	splittedPath := strings.Split(r.URL.Path, "/")
	gossip := strings.TrimSpace(splittedPath[len(splittedPath)-1])
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "GET" {
		if len(gossip) == 0 {
			gah.List(w, r)
		} else {
			gah.Get(gossip, w, r)
		}
	} else if r.Method == "POST" {
		gah.Create(w, r)
	}
}

func (gah *GossipResourceHandler) List(w http.ResponseWriter, r *http.Request) {
	log.Println("list")
	payload := &GossipListPayload{[]*GossipPayload{}}
	gossips, err := gah.service.FindAllGossip()
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}
	for _, g := range gossips {
		payload.Gossips = append(payload.Gossips, gah.buildGossipPayload(g))
	}
	response, _ := json.Marshal(payload)
	w.Write([]byte(response))
}

func (gah *GossipResourceHandler) Get(label string, w http.ResponseWriter, r *http.Request) {
	log.Println("GET", label)
	gossip, err := gah.service.FindGossipByLabel(label)
	if gossip == nil {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}
	payload := gah.buildGossipPayload(gossip)
	response, _ := json.Marshal(payload)
	w.Write([]byte(response))
}

func (gah *GossipResourceHandler) Create(w http.ResponseWriter, r *http.Request) {
	payload := &GossipPayload{}
	if err := json.NewDecoder(r.Body).Decode(payload); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}

	log.Println("POST", payload)
	gossip, classifiers := payload.ToModel()
	if err := gah.service.Save(gossip, classifiers); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}
}

func (gah *GossipResourceHandler) buildGossipPayload(g *Gossip) *GossipPayload {
	classifiers, err := gah.service.FindClassifiersByGossip(g)
	if err != nil {
		log.Println(err)
		return nil
	}

	cPayload := map[string][]string{}
	for _, c := range classifiers {
		cPayload[c.Label] = c.Patterns
	}
	return &GossipPayload{g.Label, g.Subjects, cPayload}
}
