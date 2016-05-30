package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
)

type GossipPayload struct {
	Gossip      string   `json:"gossip"`
	Subjects    []string `json:"subjects"`
	Classifiers string   `json:"classifiers"`
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
		http.Error(w, err.Error(), 500)
		return
	}
	classfiers, err := gah.deserializeClassifiers(payload.Classifiers)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	gossip := &Gossip{Label: payload.Gossip, Subjects: payload.Subjects}
	err = gah.service.Save(gossip, classfiers)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func (gah *GossipResourceHandler) deserializeClassifiers(cString string) ([]*GossipClassifier, error) {
	currentLabel := ""
	patterns := []string{}
	result := []*GossipClassifier{}
	appendClassifier := func(label string, patterns []string) {
		if len(patterns) > 0 {
			result = append(result, &GossipClassifier{Label: label, Patterns: patterns})
		}
	}

	for _, line := range strings.Split(cString, "\n") {
		if line[0] == ':' {
			if len(currentLabel) > 0 {
				appendClassifier(currentLabel, patterns)
				patterns = []string{}
			}
			label := line[1:]
			if label != currentLabel {
				currentLabel = label
			}
		} else if len(currentLabel) > 0 {
			patterns = append(patterns, line)
		} else {
			return nil, errors.New("Error deserializeClassifiers: Classifier without label")
		}
	}

	appendClassifier(currentLabel, patterns)
	return result, nil
}

func (gah *GossipResourceHandler) buildGossipPayload(g *Gossip) *GossipPayload {
	classifiers, err := gah.service.FindClassifiersByGossip(g)
	if err != nil {
		log.Println(err)
		return nil
	}
	classifiersString := ""
	for _, c := range classifiers {
		classifiersString += ":" + c.Label + "\n"
		classifiersString += strings.Join(c.Patterns, "\n")
	}

	return &GossipPayload{g.Label, g.Subjects, classifiersString}
}
