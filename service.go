package main

import (
	"log"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type GossipService interface {
	FindAllGossip() ([]*Gossip, error)
	FindGossipByLabel(label string) (*Gossip, error)
	FindClassifiersByGossip(g *Gossip) ([]*GossipClassifier, error)
	FindClassifiersEvents(label string, from time.Time, to time.Time, limit int) ([]*GossipClassifierEvent, error)
	CreateGossip(g *Gossip, c []*GossipClassifier) error
	UpdateGossip(gossipLabel string, g *Gossip, c []*GossipClassifier) error
	RemoveGossip(gossipLabel string) error
	SaveClassifiers(g *Gossip, c []*GossipClassifier) error
	SaveEvent(e *GossipClassifierEvent) error
}

type MongoGossipService struct {
	session      *mgo.Session
	gossipC      *mgo.Collection
	classifiersC *mgo.Collection
	eventsC      *mgo.Collection
}

func NewMongoGossipService(mongoURL string, dbName string) *MongoGossipService {
	log.Println("Connecting to Mongo on DB " + dbName)
	s, err := mgo.Dial(mongoURL)
	if err != nil {
		panic(err)
	}

	gossipC := s.DB(dbName).C("gossip")
	index := mgo.Index{Key: []string{"label"}, Unique: true, DropDups: true, Background: true, Sparse: true}
	err = gossipC.EnsureIndex(index)
	if err != nil {
		panic(err)
	}

	classifiersC := s.DB(dbName).C("gossip_classifiers")
	eventsC := s.DB(dbName).C("gossip_classifier_events")

	return &MongoGossipService{s, gossipC, classifiersC, eventsC}
}

func (s *MongoGossipService) FindAllGossip() ([]*Gossip, error) {
	results := []*Gossip{}
	err := s.gossipC.Find(nil).All(&results)
	return results, err
}

func (s *MongoGossipService) FindGossipByLabel(label string) (*Gossip, error) {
	result := &Gossip{}
	err := s.gossipC.Find(bson.M{"label": label}).One(&result)
	if err != nil {
		return nil, err
	}
	return result, err
}

func (s *MongoGossipService) FindClassifiersByGossip(g *Gossip) ([]*GossipClassifier, error) {
	results := []*GossipClassifier{}
	err := s.classifiersC.Find(bson.M{"gossipid": g.ID}).All(&results)
	return results, err
}

func (s *MongoGossipService) CreateGossip(g *Gossip, classifiers []*GossipClassifier) error {
	log.Println("Create", g.Label)
	err := s.gossipC.Insert(g)
	if err != nil {
		return err
	}

	g, err = s.FindGossipByLabel(g.Label)
	if err = s.SaveClassifiers(g, classifiers); err != nil {
		return err
	}
	return nil
}

func (s *MongoGossipService) UpdateGossip(gossipLabel string, g *Gossip, classifiers []*GossipClassifier) error {
	log.Println("Update", gossipLabel)
	found, err := s.FindGossipByLabel(gossipLabel)
	if err != nil {
		return err
	}
	if err = s.gossipC.Update(bson.M{"_id": found.ID}, g); err != nil {
		return err
	}
	if err = s.SaveClassifiers(found, classifiers); err != nil {
		return err
	}
	return nil
}

func (s *MongoGossipService) RemoveGossip(gossipLabel string) error {
	g, err := s.FindGossipByLabel(gossipLabel)
	if err != nil {
		return err
	}

	s.classifiersC.RemoveAll(bson.M{"gossipid": g.ID})
	s.eventsC.RemoveAll(bson.M{"gossipid": g.ID})
	err = s.gossipC.Remove(g)
	if err != nil {
		return err
	}
	return nil
}

func (s *MongoGossipService) SaveClassifiers(g *Gossip, classifiers []*GossipClassifier) error {
	s.classifiersC.RemoveAll(bson.M{"gossipid": g.ID})
	for _, c := range classifiers {
		c.GossipId = g.ID
		err := s.classifiersC.Insert(c)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *MongoGossipService) SaveEvent(e *GossipClassifierEvent) error {
	return s.eventsC.Insert(e)
}

func (s *MongoGossipService) FindClassifiersEvents(
	label string, from time.Time, to time.Time, limit int) ([]*GossipClassifierEvent, error) {
	log.Println("Searching events with label:", label, "from:", from, "to:", to, "limited by:", limit)
	g, err := s.FindGossipByLabel(label)
	if err != nil {
		return nil, err
	}

	result := []*GossipClassifierEvent{}
	query := bson.M{"gossipid": g.ID, "timestamp": bson.M{"$gte": from, "$lte": to.Add(time.Second)}}
	err = s.eventsC.Find(query).Limit(limit).Sort("-timestamp").All(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
