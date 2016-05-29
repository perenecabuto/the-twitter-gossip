package main

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type Gossip struct {
	ID       bson.ObjectId `bson:"_id,omitempty`
	Label    string
	Subjects []string
}

type GossipClassifier struct {
	ID       bson.ObjectId `bson:"_id,omitempty`
	Label    string
	Patterns []string
}

type GossipEvent struct {
	ID         bson.ObjectId `bson:"_id,omitempty`
	Name       string
	Value      string
	Annotation string
	Timestamp  time.Time
}
