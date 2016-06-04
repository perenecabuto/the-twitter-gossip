package main

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

func GossipEventHistoryPayloadFromModel(gossipLabel string, list []*GossipClassifierEvent) *GossipEventHistoryPayload {
	history := []*EventGroupPayload{}
	for _, e := range list {
		history = append(history, EventGroupPayloadFromModel(gossipLabel, e))
	}
	return &GossipEventHistoryPayload{gossipLabel, history}
}

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
	Gossips []*GossipPayload `json:"gossips"`
}
