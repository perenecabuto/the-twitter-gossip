package main

import (
	"regexp"

	"github.com/dghubble/go-twitter/twitter"
)

type MessageClassifier struct {
	Label    string
	patterns []*regexp.Regexp
}

func NewMessageClassifier(label string, patterns []string) *MessageClassifier {
	reList := []*regexp.Regexp{}
	for _, p := range patterns {
		reList = append(reList, regexp.MustCompile(p))
	}

	return &MessageClassifier{label, reList}
}

func (mc *MessageClassifier) Matches(message string) bool {
	for _, pattern := range mc.patterns {
		if pattern.MatchString(message) {
			return true
		}
	}
	return false
}

type ClassifierMatchCallback func(label string, t *twitter.Tweet)

type MessageClassifierListener struct {
	classifiers  []*MessageClassifier
	inputChann   chan *twitter.Tweet
	OnMatchChann chan string
}

func NewMessageClassifierListener(classifiers []*MessageClassifier) *MessageClassifierListener {
	return &MessageClassifierListener{classifiers, make(chan *twitter.Tweet), make(chan string)}
}

func (mcl *MessageClassifierListener) AddMessageClassifier(mc *MessageClassifier) {
	mcl.classifiers = append(mcl.classifiers, mc)
}

func (mcl *MessageClassifierListener) OnTweet(tweet *twitter.Tweet) {
	for _, classifier := range mcl.classifiers {
		if classifier.Matches(tweet.Text) {
			mcl.OnMatchChann <- classifier.Label
		}
	}
}
