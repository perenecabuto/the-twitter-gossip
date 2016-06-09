package main

import "time"

type EventGroupPayload struct {
	Gossip    string         `json:"gossip"`
	Timestamp int64          `json:"timestamp"`
	Events    map[string]int `json:"events"`
}

func EventGroupPayloadFromModel(gossipLabel string, e *GossipClassifierEvent) *EventGroupPayload {
	return &EventGroupPayload{gossipLabel, e.Timestamp.Unix(), e.Events}
}

type GossipEventHistoryPayload struct {
	Gossip  string               `json:"gossip"`
	History []*EventGroupPayload `json:"history"`
}

func NewGossipEventHistoryPayloadFromModel(gossipLabel string, list []*GossipClassifierEvent) *GossipEventHistoryPayload {
	history := []*EventGroupPayload{}
	for _, e := range list {
		history = append(history, EventGroupPayloadFromModel(gossipLabel, e))
	}
	return &GossipEventHistoryPayload{gossipLabel, history}
}

type GossipPayload struct {
	Gossip         string              `json:"gossip"`
	Subjects       []string            `json:"subjects"`
	Classifiers    map[string][]string `json:"classifiers"`
	WorkerState    string              `json:"state"`
	WorkerInterval time.Duration       `json:"interval"`
}

func NewGossipPayloadFromModel(g *Gossip, classifiers []*GossipClassifier, state string) *GossipPayload {
	cPayload := map[string][]string{}
	for _, c := range classifiers {
		cPayload[c.Label] = c.Patterns
	}

	return &GossipPayload{g.Label, g.Subjects, cPayload, state, g.WorkerInterval / time.Second}
}

func (p *GossipPayload) ToModel() (*Gossip, []*GossipClassifier) {
	interval := p.WorkerInterval * time.Second
	gossip := &Gossip{Label: p.Gossip, Subjects: p.Subjects, WorkerInterval: interval}
	classifiers := []*GossipClassifier{}
	for label, patterns := range p.Classifiers {
		classifiers = append(classifiers, &GossipClassifier{Label: label, Patterns: patterns})
	}
	return gossip, classifiers
}

type GossipListPayload struct {
	Gossips []*GossipPayload `json:"gossips"`
}
