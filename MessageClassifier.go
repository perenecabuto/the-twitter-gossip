package main

import (
	"regexp"

	"github.com/dghubble/go-twitter/twitter"
)

type MessageClassifier struct {
	label         string
	patterns      []*regexp.Regexp
	matchCallback func(tweet *twitter.Tweet)
}

func NewMessageClassifier(label string, patterns []*regexp.Regexp, callback func(tweet *twitter.Tweet)) *MessageClassifier {
	return &MessageClassifier{label, patterns, callback}
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
	mc.matchCallback(tweet)
}

type MessageClassifierListener struct {
	inputChann  chan *twitter.Tweet
	classifiers []*MessageClassifier
}

func NewMessageClassifierListener(classifiers []*MessageClassifier) *MessageClassifierListener {
	return &MessageClassifierListener{make(chan *twitter.Tweet), classifiers}
}

func (mcl *MessageClassifierListener) AddMessageClassifier(mc *MessageClassifier) {
	mcl.classifiers = append(mcl.classifiers, mc)
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
