package main

type DummyGossipService struct {
}

func (dr *DummyGossipService) FindGossipByLabel(label string) *Gossip {
	gossip := &Gossip{
		Label:    label,
		Subjects: []string{"pt"},
	}

	return gossip
}

func (dr *DummyGossipService) FindClassifiersByGossip(g *Gossip) []*GossipClassifier {
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
