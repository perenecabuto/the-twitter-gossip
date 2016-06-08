package main

import (
	"log"
	"os"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

var (
	consumerKey    = os.Getenv("TWITTER_CONSUMER_KEY")
	consumerSecret = os.Getenv("TWITTER_CONSUMER_SECRET")
	tokenValue     = os.Getenv("TWITTER_TOKEN")
	tokenSecret    = os.Getenv("TWITTER_TOKEN_SECRET")
	config         = oauth1.NewConfig(consumerKey, consumerSecret)
	token          = oauth1.NewToken(tokenValue, tokenSecret)
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
