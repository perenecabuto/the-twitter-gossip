package main

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type Gossip struct {
	ID             bson.ObjectId `bson:"_id,omitempty"`
	Label          string
	WorkerInterval time.Duration
	Subjects       []string
}

type GossipClassifier struct {
	ID       bson.ObjectId `bson:"_id,omitempty",json:"-"`
	GossipId bson.ObjectId
	Label    string
	Patterns []string
}

type GossipClassifierEvent struct {
	ID        bson.ObjectId `bson:"_id,omitempty",json:"-"`
	GossipId  bson.ObjectId
	Events    map[string]int
	Timestamp time.Time
}
