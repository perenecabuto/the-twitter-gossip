package main

import (
	"fmt"
	"log"

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
	stream    *twitter.Stream
	listeners []TwitterStreamListener
	stopChann chan bool
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

	return &TwitterStream{tracks, stream, []TwitterStreamListener{}, make(chan bool)}
}

func (ts *TwitterStream) AddListener(listener TwitterStreamListener) {
	ts.listeners = append(ts.listeners, listener)
}

func (ts *TwitterStream) Listen() {
	fmt.Println("Waiting for messages about ", ts.tracks)
	for {
		select {
		case message := <-ts.stream.Messages:
			tweet, ok := message.(*twitter.Tweet)
			if ok {
				for _, listener := range ts.listeners {
					go func(l TwitterStreamListener) {
						go l.OnTweet(tweet)
					}(listener)
				}
			}
		case <-ts.stopChann:
			log.Println("! Stopping TwitterStream: ", ts.tracks)
			ts.stream.Stop()
			ts.stopChann <- true
			return
		}
	}
}

func (ts *TwitterStream) Stop() {
	ts.stopChann <- true
	<-ts.stopChann
}
