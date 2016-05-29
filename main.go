package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"golang.org/x/net/websocket"
)

func main() {
	tracks := []string{"pt"}
	stream := NewTwitterStream(tracks)

	//printer := NewTweetPrinter()
	//stream.AddListener(printer)

	eventsChann := make(chan string)
	reportInterval := 10 * time.Second
	report := NewReportWorker(reportInterval, func(t time.Time, events EventGroup) {
		msg, err := json.Marshal(events)
		if err != nil {
			log.Println(err)
			return
		}

		eventsChann <- string(msg)
		fmt.Printf("Sent: %s\n", msg)
	})

	anythingClassifier := &MessageClassifier{"qq coisa", []*regexp.Regexp{
		regexp.MustCompile(".*"),
	}, nil}

	happinessClassifier := &MessageClassifier{"ruim", []*regexp.Regexp{
		regexp.MustCompile("corrup"),
		regexp.MustCompile("defeito"),
		regexp.MustCompile("porra"),
		regexp.MustCompile("problema"),
		regexp.MustCompile("login"),
		regexp.MustCompile("odeio"),
		regexp.MustCompile("raiva"),
	}, nil}

	problemClassifier := &MessageClassifier{"bom", []*regexp.Regexp{
		regexp.MustCompile("legal"),
		regexp.MustCompile("bem"),
		regexp.MustCompile("indo bem"),
		regexp.MustCompile("bom"),
		regexp.MustCompile("gostei"),
		regexp.MustCompile("foda"),
	}, nil}

	classifiers := []*MessageClassifier{anythingClassifier, happinessClassifier, problemClassifier}
	messageClassifier := NewMessageClassifierListener(classifiers)
	messageClassifier.OnClassifierMatch = func(label string, t *twitter.Tweet) {
		go report.ReportEvent(label)
	}

	stream.AddListener(messageClassifier)

	go report.Start()
	go stream.Listen()

	http.Handle("/events", websocket.Handler(func(ws *websocket.Conn) {
		for msg := range eventsChann {
			if _, err := ws.Write([]byte(msg)); err != nil {
				log.Println(err)
				return
			}
		}
	}))

	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
