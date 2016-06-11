package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

type Suprême struct {
	FileName    string    `json:"fileName"`
	MagnetURI   string    `json:"magnetURI"`
	WaveformURI string    `json:"waveformURI"`
	Duration    uint64    `json:"duration"`
	CreatedAt   time.Time `json:"createdAt"`
}

func index(ws *websocket.Conn) {

	done := make(chan struct{})

	go func() {

		for {
			select {
			case <-done:
				return
			case <-time.After(time.Second * 2):
				str, _ := json.Marshal(Suprême{"hello from go!", "asdf", "waveform.png", 3695, time.Now()})
				ws.Write(str)
				log.Printf("sent: %s\n", str)
				return
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		default:
			var s Suprême
			err := websocket.JSON.Receive(ws, &s)
			log.Printf("got: %#v, %v\n", s, err)
			switch err {
			case io.EOF:
				close(done)
			case nil:
				log.Println("(do work with message here...)")
			default:
				log.Fatalln(err)
			}
		}
	}
}

func main() {
	http.Handle("/", websocket.Handler(index))
	log.Println("Listening on :12345")
	log.Panicln(http.ListenAndServe(":12345", nil))
}
