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
			case <-time.After(time.Hour * 30):
				str, _ := json.Marshal(Suprême{"fake", "asdf", "waveform.png", 3695, time.Now()})
				ws.Write(str)
				log.Printf("sent: %s\n", str)
				return
			}
		}
	}()

	dec := json.NewDecoder(ws)
	for {
		select {
		case <-done:
			return
		default:
			if dec.More() {
				var s Suprême
				err := dec.Decode(&s)
				log.Printf("got: %#v, %v\n", s, err)
				switch err {
				case io.EOF:
					close(done)
				case nil:
					str, _ := json.Marshal(Suprême{"hello from go!", s.MagnetURI, s.WaveformURI, s.Duration, time.Now()})
					ws.Write(str)
					log.Printf("sent: %s\n", str)
				default:
					log.Println("error:", err)
				}
			}
		}
	}
}

func main() {
	http.Handle("/", websocket.Handler(index))
	log.Println("Listening on :12345")
	log.Panicln(http.ListenAndServe(":12345", nil))
}
