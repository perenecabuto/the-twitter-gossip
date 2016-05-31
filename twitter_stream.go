package main

import (
	"fmt"
	"log"
	"reflect"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

var (
	config     = oauth1.NewConfig("PtBsLxlPzwD39F1dfPsxxKfhL", "TMduA8ViNysuVaOztHk8yFGdj0VK6NWGDCrxe081bYd0QZyqYd")
	token      = oauth1.NewToken("1018696039-nYjPFKCIB69nRoKMvUNrV5CzhJzKn0gbXEhshRe", "luYhEOT2Id0uhXcMz5a4cDcbAfhYefw4kXUvUsIQS8lZ2")
	httpClient = config.Client(oauth1.NoContext, token)
	client     = twitter.NewClient(httpClient)
)

type TwitterStreamListener interface {
	OnTweet(*twitter.Tweet)
}

type TwitterStream struct {
	tracks    []string
	listeners []TwitterStreamListener
	stream    *twitter.Stream
}

func NewTwitterStream(tracks []string) *TwitterStream {
	params := &twitter.StreamFilterParams{
		Track:         tracks,
		StallWarnings: twitter.Bool(true),
	}
	stream, err := client.Streams.Filter(params)
	if err != nil {
		log.Panic("Error!!!", err)
	}

	return &TwitterStream{tracks: tracks, stream: stream, listeners: []TwitterStreamListener{}}
}

func (ts *TwitterStream) Listen() {
	fmt.Println("Waiting for messages about ", ts.tracks)
	for message := range ts.stream.Messages {
		tweet, ok := message.(*twitter.Tweet)
		if ok {
			for _, listener := range ts.listeners {
				go func(l TwitterStreamListener) {
					go l.OnTweet(tweet)
				}(listener)
			}
		}
	}
}

func (ts *TwitterStream) AddListener(listener TwitterStreamListener) {
	fmt.Println("Add ", reflect.TypeOf(listener))
	ts.listeners = append(ts.listeners, listener)
}
