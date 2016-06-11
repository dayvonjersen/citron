package main

import (
	"io"
	"log"
	"net/http"
	"strings"
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

func getURI(magnet string) string {
	// what could go wrong?
	return strings.Split(strings.TrimLeft(magnet, "magnet:?xt=urn:btih:"), "&")[0]
}

var pubsubhubbub chan string

var upgrader = websocket.Upgrader{
	// if you encounter this error:
	//
	// """upgrade: websocket: origin not allowed"""
	//
	// this gets around it:
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

	tmp := db.getRange(0, 5)
	db.init()
	for _, uri := range tmp {
		s := db.get(uri)
		c.WriteJSON(s)
		log.Println("sent:", uri)
	}

	esc := make(chan struct{})
	go func() {
		for {
			if uri, ok := <-pubsubhubbub; ok {
				s := db.get(uri)
				c.WriteJSON(s)
				log.Println("sent:", uri)
			} else {
				log.Println("writer shutting down")
				close(esc)
			}
		}
	}()

	go func() {
	here:
		for {
			var s Suprême
			err := c.ReadJSON(&s)
			switch err {
			case io.EOF:
				log.Println("reader shutting down")
				break here
			case nil:

				// validate structure ...

				uri := getURI(s.MagnetURI)
				s.CreatedAt = time.Now()

				log.Println("got:", uri)

				db.set(uri, s)
				pubsubhubbub <- uri

			default:
				log.Println("error:", err)
				break here
			}
		}
	}()

	<-esc
}

var db datastore

func main() {
	//db.init()
	pubsubhubbub = make(chan string)
	http.HandleFunc("/", index)
	log.Println("Listening on :12345")
	log.Panicln(http.ListenAndServe(":12345", nil))
}
