package main

import (
	"regexp"

	"github.com/dghubble/go-twitter/twitter"
)

type MessageClassifier struct {
	Label         string
	patterns      []*regexp.Regexp
	MatchCallback func(t *twitter.Tweet)
}

func (mc *MessageClassifier) Matches(message string) bool {
	for _, pattern := range mc.patterns {
		if pattern.MatchString(message) {
			return true
		}
	}
	return false
}

func (mc *MessageClassifier) OnMatch(tweet *twitter.Tweet) {
	mc.MatchCallback(tweet)
}

type MessageClassifierListener struct {
	inputChann        chan *twitter.Tweet
	classifiers       []*MessageClassifier
	OnClassifierMatch func(label string, t *twitter.Tweet)
}

func NewMessageClassifierListener(classifiers []*MessageClassifier) *MessageClassifierListener {
	listener := &MessageClassifierListener{make(chan *twitter.Tweet), []*MessageClassifier{}, nil}
	for _, classifier := range classifiers {
		listener.AddMessageClassifier(classifier)
	}
	return listener
}

func (mcl *MessageClassifierListener) AddMessageClassifier(mc *MessageClassifier) {
	mcl.classifiers = append(mcl.classifiers, mc)
	mc.MatchCallback = func(t *twitter.Tweet) {
		mcl.OnClassifierMatch(mc.Label, t)
	}
}

func (tp *MessageClassifierListener) InputChann() chan *twitter.Tweet {
	return tp.inputChann
}

func (tp *MessageClassifierListener) OnTweet(tweet *twitter.Tweet) {
	for _, classifier := range tp.classifiers {
		if classifier.Matches(tweet.Text) {
			classifier.OnMatch(tweet)
		}
	}
}
