package main

type GossipService interface {
	FindAllGossip() []*Gossip
	FindGossipByLabel(label string) *Gossip
	FindClassifiersByGossip(g *Gossip) []*GossipClassifier
}

type DummyGossipService struct{}

func (s *DummyGossipService) FindAllGossip() []*Gossip {
	list := []*Gossip{}
	list = append(list, s.FindGossipByLabel("morte"))
	list = append(list, s.FindGossipByLabel("fofoca"))
	list = append(list, s.FindGossipByLabel("problema"))
	list = append(list, s.FindGossipByLabel("buceta"))
	list = append(list, s.FindGossipByLabel("cu"))
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

func ConvertToMessageClassifiers(gclassifiers []*GossipClassifier) []*MessageClassifier {
	classifiers := []*MessageClassifier{}
	for _, gclassifier := range gclassifiers {
		newClassifier := NewMessageClassifier(gclassifier.Label, gclassifier.Patterns)
		classifiers = append(classifiers, newClassifier)
	}
	return classifiers
}
