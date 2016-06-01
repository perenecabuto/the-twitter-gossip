package main

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type Gossip struct {
	ID       bson.ObjectId `bson:"_id,omitempty"`
	Label    string
	Subjects []string
}

type GossipClassifier struct {
	ID       bson.ObjectId `bson:"_id,omitempty",json:"-"`
	GossipId bson.ObjectId
	Label    string
	Patterns []string
}

type GossipClassifierEvent struct {
	ID         bson.ObjectId `bson:"_id,omitempty",json:"-"`
	GossipId   bson.ObjectId
	Label      string
	Value      int
	Annotation string
	Timestamp  time.Time
}
