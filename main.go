package main

import (
	"regexp"

	"github.com/dghubble/go-twitter/twitter"
)

func main() {
	tracks := []string{"cartola"}
	stream := NewTwitterStream(tracks)

	//printer := NewTweetPrinter()
	//stream.AddListener(printer)

	report := NewReportWorker()

	anythingClassifier := &MessageClassifier{"qq coisa", []*regexp.Regexp{
		regexp.MustCompile(".*"),
	}, func(tweet *twitter.Tweet) {
		report.ReportEvent("qq coisa")
	}}

	happinessClassifier := &MessageClassifier{"ruim", []*regexp.Regexp{
		regexp.MustCompile("problema"),
		regexp.MustCompile("login"),
		regexp.MustCompile("odeio"),
		regexp.MustCompile("raiva"),
	}, func(tweet *twitter.Tweet) {
		report.ReportEvent("ruim")
	}}

	problemClassifier := &MessageClassifier{"bom", []*regexp.Regexp{
		regexp.MustCompile("indo bem"),
		regexp.MustCompile("bom"),
		regexp.MustCompile("gostei"),
		regexp.MustCompile("foda"),
	}, func(tweet *twitter.Tweet) {
		report.ReportEvent("bom")
	}}

	classifiers := []*MessageClassifier{anythingClassifier, happinessClassifier, problemClassifier}
	messageClassifier := NewMessageClassifierListener(classifiers)
	stream.AddListener(messageClassifier)

	go report.Start()
	stream.Listen()
}
