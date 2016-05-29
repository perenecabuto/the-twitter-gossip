package main

import (
	"fmt"
	"strings"

	"github.com/dghubble/go-twitter/twitter"
)

var (
	format = strings.Replace(`
	-----------------
	User: %s
	Message: %s
	Source: %s
	Coordinates: %v
	`, "\t", "", 1000)
)

type TweetPrinter struct {
	listener chan *twitter.Tweet
}

func NewTweetPrinter() *TweetPrinter {
	input := make(chan *twitter.Tweet)
	printer := &TweetPrinter{input}
	return printer
}

func (tp *TweetPrinter) InputChann() chan *twitter.Tweet {
	return tp.listener
}

func (tp *TweetPrinter) OnTweet(tweet *twitter.Tweet) {
	var coord *twitter.Coordinates
	if tweet.Coordinates != nil {
		coord = tweet.Coordinates
	}
	fmt.Printf(format,
		tweet.User.ScreenName,
		tweet.Text,
		tweet.Source,
		coord,
	)
}
