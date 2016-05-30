package main

type GossipService interface {
	FindAllGossip() []*Gossip
	FindGossipByLabel(label string) *Gossip
	FindClassifiersByGossip(g *Gossip) []*GossipClassifier
}

type DummyGossipService struct{}

func (s *DummyGossipService) FindAllGossip() []*Gossip {
	list := []*Gossip{}
	list = append(list, s.FindGossipByLabel("cu"))
	list = append(list, s.FindGossipByLabel("cartola"))
	list = append(list, s.FindGossipByLabel("problema"))
	list = append(list, s.FindGossipByLabel("fofoca"))
	return list
}

func (s *DummyGossipService) FindGossipByLabel(label string) *Gossip {
	gossip := &Gossip{
		Label:    label,
		Subjects: []string{label},
	}

	return gossip
}

func (s *DummyGossipService) FindClassifiersByGossip(g *Gossip) []*GossipClassifier {
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
	}
}
