package main

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Suprême struct {
	FileName    string    `json:"fileName"`
	MagnetURI   string    `json:"magnetURI"`
	WaveformURI string    `json:"waveformURI"`
	Duration    uint64    `json:"duration"`
	CreatedAt   time.Time `json:"createdAt"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	}}

func index(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	defer c.Close()

	for {
		var s Suprême
		err := c.ReadJSON(&s)
		log.Printf("got: %v, %v\n", s, err)
		switch err {
		case io.EOF:
			break
		case nil:
			c.WriteJSON(Suprême{"hello from go!", s.MagnetURI, s.WaveformURI, s.Duration, time.Now()})
			log.Printf("sent: [stuff]")
		default:
			log.Println("error:", err)
		}
	}
}

func main() {
	http.HandleFunc("/", index)
	log.Println("Listening on :12345")
	log.Panicln(http.ListenAndServe(":12345", nil))
}
