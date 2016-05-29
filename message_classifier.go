package main

import (
	"regexp"

	"github.com/dghubble/go-twitter/twitter"
)

type MessageClassifier struct {
	Label    string
	patterns []*regexp.Regexp
	callback func(t *twitter.Tweet)
}

func NewMessageClassifier(label string, patterns []string) *MessageClassifier {
	re_list := []*regexp.Regexp{}
	for _, p := range patterns {
		re_list = append(re_list, regexp.MustCompile(p))
	}

	return &MessageClassifier{label, re_list, nil}
}

func (mc *MessageClassifier) Matches(message string) bool {
	for _, pattern := range mc.patterns {
		if pattern.MatchString(message) {
			return true
		}
	}
	return false
}

func (mc *MessageClassifier) SetOnMatch(callback func(t *twitter.Tweet)) {
	mc.callback = callback
}

func (mc *MessageClassifier) OnMatch(tweet *twitter.Tweet) {
	if mc.callback != nil {
		mc.callback(tweet)
	}
}

type ClassifierMatchCallback func(label string, t *twitter.Tweet)

type MessageClassifierListener struct {
	inputChann  chan *twitter.Tweet
	classifiers []*MessageClassifier
	callback    ClassifierMatchCallback
}

func NewMessageClassifierListener(classifiers []*MessageClassifier) *MessageClassifierListener {
	listener := &MessageClassifierListener{make(chan *twitter.Tweet), []*MessageClassifier{}, nil}
	for _, classifier := range classifiers {
		listener.AddMessageClassifier(classifier)
	}
	return listener
}

func (mcl *MessageClassifierListener) SetOnMatch(callback ClassifierMatchCallback) {
	mcl.callback = callback
}

func (mcl *MessageClassifierListener) AddMessageClassifier(mc *MessageClassifier) {
	mcl.classifiers = append(mcl.classifiers, mc)
	mc.SetOnMatch(func(t *twitter.Tweet) {
		if mcl.callback != nil {
			mcl.callback(mc.Label, t)
		}
	})
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
