package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

type suprême struct {
	fileName    string    `json:"fileName"`
	magnetURI   string    `json:"magnetURI"`
	waveformURI string    `json:"waveformURI"`
	duration    time.Time `json:"duration"`
	createdAt   time.Time `json:"createdAt"`
}

func index(ws *websocket.Conn) {
	for {
		str, _ := json.Marshal(suprême{"only a test", "asdf", "waveform.png", time.Now(), time.Now()})
		websocket.Message.Send(ws, str)
		log.Printf("sent: %#v\n", str)
	}
}

func main() {
	http.Handle("/", websocket.Handler(index))
	log.Println("Listening on :12345")
	log.Panicln(http.ListenAndServe(":12345", nil))
}
