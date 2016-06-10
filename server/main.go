package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

type Suprême struct {
	FileName    string    `json:"fileName"`
	MagnetURI   string    `json:"magnetURI"`
	WaveformURI string    `json:"waveformURI"`
	Duration    time.Time `json:"duration"`
	CreatedAt   time.Time `json:"createdAt"`
}

func index(ws *websocket.Conn) {
	go func() {
		for {
			var s Suprême
			err := websocket.JSON.Receive(ws, &s)
			log.Printf("got: %#v, %v\n", s, err)
		}
	}()

	for {
		<-time.After(time.Second * 2)
		str, _ := json.Marshal(Suprême{"only a test", "asdf", "waveform.png", time.Now(), time.Now()})
		ws.Write(str)
		log.Printf("sent: %s\n", str)
	}
}

func main() {
	http.Handle("/", websocket.Handler(index))
	log.Println("Listening on :12345")
	log.Panicln(http.ListenAndServe(":12345", nil))
}
