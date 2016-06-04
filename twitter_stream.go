package main

import (
	"log"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

var (
	config = oauth1.NewConfig("PtBsLxlPzwD39F1dfPsxxKfhL", "TMduA8ViNysuVaOztHk8yFGdj0VK6NWGDCrxe081bYd0QZyqYd")
	token  = oauth1.NewToken("1018696039-nYjPFKCIB69nRoKMvUNrV5CzhJzKn0gbXEhshRe", "luYhEOT2Id0uhXcMz5a4cDcbAfhYefw4kXUvUsIQS8lZ2")
)

type TwitterStreamListener interface {
	OnTweet(*twitter.Tweet)
}

type TwitterStream struct {
	tracks    []string
	listeners []TwitterStreamListener
	stopChann chan bool
}

func NewTwitterStream(tracks []string) *TwitterStream {
	return &TwitterStream{tracks, []TwitterStreamListener{}, make(chan bool)}
}

func (ts *TwitterStream) AddListener(listener TwitterStreamListener) {
	ts.listeners = append(ts.listeners, listener)
}

func (ts *TwitterStream) Listen() {
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)
	params := &twitter.StreamFilterParams{Track: ts.tracks, StallWarnings: twitter.Bool(true)}
	stream, err := client.Streams.Filter(params)
	if err != nil {
		log.Panic("Error!!!", err)
	}

	log.Println("! Listen to messages about ", ts.tracks)
	for {
		select {
		case message := <-stream.Messages:
			tweet, ok := message.(*twitter.Tweet)
			if ok {
				for _, l := range ts.listeners {
					go l.OnTweet(tweet)
				}
			}
		case <-ts.stopChann:
			log.Println("! Stopping TwitterStream: ", ts.tracks)
			stream.Stop()
			ts.stopChann <- true
			return
		}
	}
}

func (ts *TwitterStream) Stop() {
	ts.stopChann <- true
	<-ts.stopChann
}
