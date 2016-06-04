package main

import (
	"log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type GossipService interface {
	FindAllGossip() ([]*Gossip, error)
	FindGossipByLabel(label string) (*Gossip, error)
	FindClassifiersByGossip(g *Gossip) ([]*GossipClassifier, error)
	FindClassifiersEvents(label string) ([]*GossipClassifierEvent, error)
	SaveGossip(g *Gossip, c []*GossipClassifier) error
	SaveEvent(e *GossipClassifierEvent) error
}

type MongoGossipService struct {
	session      *mgo.Session
	dbName       string
	gossipC      *mgo.Collection
	classifiersC *mgo.Collection
	eventsC      *mgo.Collection
}

func NewMongoGossipService() *MongoGossipService {
	s, err := mgo.Dial("mongodb://localhost")
	if err != nil {
		panic(err)
	}

	dbName := "TheTwitterGossip"
	gossipC := s.DB(dbName).C("gossip")
	index := mgo.Index{Key: []string{"label"}, Unique: true, DropDups: true, Background: true, Sparse: true}
	err = gossipC.EnsureIndex(index)
	if err != nil {
		panic(err)
	}

	classifiersC := s.DB(dbName).C("gossip_classifiers")
	eventsC := s.DB(dbName).C("gossip_classifier_events")

	return &MongoGossipService{s, dbName, gossipC, classifiersC, eventsC}
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

func (s *MongoGossipService) SaveGossip(g *Gossip, classifiers []*GossipClassifier) error {
	var err error
	var found *Gossip
	if found, err = s.FindGossipByLabel(g.Label); found != nil && err == nil {
		err = s.gossipC.Update(bson.M{"_id": found.ID}, g)
		g.ID = found.ID
	} else {
		log.Println("create", g)
		err = s.gossipC.Insert(g)
		found, err = s.FindGossipByLabel(g.Label)
	}
	if err != nil {
		panic(err)
	}

	s.classifiersC.RemoveAll(bson.M{"gossipid": found.ID})
	for _, c := range classifiers {
		c.GossipId = found.ID
		err = s.classifiersC.Insert(c)
		if err != nil {
			panic(err)
		}
	}

	return nil
}

func (s *MongoGossipService) SaveEvent(e *GossipClassifierEvent) error {
	return s.eventsC.Insert(e)
}

func (s *MongoGossipService) FindClassifiersEvents(label string) ([]*GossipClassifierEvent, error) {
	g, err := s.FindGossipByLabel(label)
	if err != nil {
		return nil, err
	}

	result := []*GossipClassifierEvent{}
	err = s.eventsC.Find(bson.M{"gossipid": g.ID}).Limit(30).Sort("-timestamp").All(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
