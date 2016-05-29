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
	tracks := []string{"cartola"}
	stream := NewTwitterStream(tracks)

	//printer := NewTweetPrinter()
	//stream.AddListener(printer)

	reportInterval := 10 * time.Second
	report := NewReportWorker(reportInterval)

	anythingClassifier := &MessageClassifier{"qq coisa", []*regexp.Regexp{
		regexp.MustCompile(".*"),
	}, nil}

	happinessClassifier := &MessageClassifier{"ruim", []*regexp.Regexp{
		regexp.MustCompile("porra"),
		regexp.MustCompile("problema"),
		regexp.MustCompile("login"),
		regexp.MustCompile("odeio"),
		regexp.MustCompile("raiva"),
	}, nil}

	problemClassifier := &MessageClassifier{"bom", []*regexp.Regexp{
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
		report.OnTimeEvent = func(t time.Time, events EventGroup) {
			msg, err := json.Marshal(events)
			if err != nil {
				log.Println(err)
				return
			}

			m, err := ws.Write([]byte(msg))
			if err != nil {
				log.Println(err)
				return
			}
			fmt.Printf("Sent: %s\n", msg[:m])
		}

		for {
			var msg *interface{}
			err := websocket.JSON.Receive(ws, &msg)
			if err != nil {
				return
			}

			fmt.Printf("Receive: %s\n", msg)
		}
	}))

	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
