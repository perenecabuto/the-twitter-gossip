package main

import (
	"fmt"
	"regexp"

	"github.com/dghubble/go-twitter/twitter"
)

func main() {
	tracks := []string{"cartola"}
	stream := NewTwitterStream(tracks)

	//printer := NewTweetPrinter()
	//stream.AddListener(printer)

	anythingClassifier := &MessageClassifier{"qq coisa", []*regexp.Regexp{
		regexp.MustCompile(".*"),
	}, func(tweet *twitter.Tweet) {
		fmt.Println("QQ Coisa porra")
	}}

	happinessClassifier := &MessageClassifier{"bom", []*regexp.Regexp{
		regexp.MustCompile("problema"),
		regexp.MustCompile("login"),
		regexp.MustCompile("odeio"),
		regexp.MustCompile("raiva"),
	}, func(tweet *twitter.Tweet) {
		fmt.Println(tweet)
		fmt.Println("SE FUDEU!!!")
	}}

	problemClassifier := &MessageClassifier{"ruim", []*regexp.Regexp{
		regexp.MustCompile("indo bem"),
		regexp.MustCompile("bom"),
		regexp.MustCompile("gostei"),
		regexp.MustCompile("foda"),
	}, func(tweet *twitter.Tweet) {
		fmt.Println("Uhuuulllll!!!")
	}}

	classifiers := []*MessageClassifier{anythingClassifier, happinessClassifier, problemClassifier}
	messageClassifier := NewMessageClassifierListener(classifiers)
	stream.AddListener(messageClassifier)

	stream.Listen()
}
