package main

import "log"

type DummyGossipService struct{}

func (s *DummyGossipService) FindAllGossip() ([]*Gossip, error) {
	result := []*Gossip{}
	labels := []string{"cartola", "example", "problema", "fofoca"}
	for _, l := range labels {
		g, _ := s.FindGossipByLabel(l)
		result = append(result, g)
	}
	return result, nil
}

func (s *DummyGossipService) FindGossipByLabel(label string) (*Gossip, error) {
	gossip := &Gossip{
		Label:    label,
		Subjects: []string{label},
	}

	return gossip, nil
}

func (s *DummyGossipService) FindClassifiersByGossip(g *Gossip) ([]*GossipClassifier, error) {
	return []*GossipClassifier{
		&GossipClassifier{Label: "Anything", Patterns: []string{
			".*",
		}},
		&GossipClassifier{Label: "Bad", Patterns: []string{
			"x",
			"corrup",
			"defeito",
			"porra",
			"problema",
			"login",
			"odeio",
			"raiva",
		}},
		&GossipClassifier{Label: "Good", Patterns: []string{
			"legal",
			"bem",
			"indo bem",
			"bom",
			"gostei",
			"foda",
		}},
	}, nil
}

func (s *DummyGossipService) SaveGossip(g *Gossip, c []*GossipClassifier) error {
	log.Printf("SaveGossip %v %+v\n", g, c)
	return nil
}

func (s *DummyGossipService) SaveEvent(e *GossipClassifierEvent) error {
	log.Printf("SaveEvent %v %+v\n", e)
	return nil
}
