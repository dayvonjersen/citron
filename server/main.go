package main

import (
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/gorilla/websocket"
)

type Suprême struct {
	FileName    string    `json:"fileName"`
	MagnetURI   string    `json:"magnetURI"`
	WaveformURI []float64 `json:"waveformURI"`
	Duration    uint64    `json:"duration"`
	CreatedAt   time.Time `json:"createdAt"`
}

func getFileHash(magnet string) (string, error) {
	u, err := url.Parse(magnet)
	if err != nil {
		return "", err
	}
	if u.Scheme != "magnet" {
		return "", fmt.Errorf("invalid magnet: url of is of scheme %s", u.Scheme)
	}

	q := u.Query()

	xt, ok := q["xt"]
	if !ok {
		return "", fmt.Errorf("invalid magnet: missing \"xt\" parameter")
	}
	if len(xt) != 1 {
		return "", fmt.Errorf("invalid magnet: invalid \"xt\" parameter")
	}

	urn := xt[0]
	if urn[0:9] != "urn:btih:" {
		return "", fmt.Errorf("invalid magnet: invalid urn")
	}

	hash := urn[9:]

	if m, _ := regexp.Match(`^[0-9A-Fa-f]{40}$`, []byte(hash)); !m {
		return "", fmt.Errorf("invalid magnet: invalid hash")
	}

	return hash, nil
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
	for _, hash := range tmp {
		s := db.get(hash)
		c.WriteJSON(s)
		log.Println("sent:", hash)
	}

	esc := make(chan struct{})
	go func() {
		for {
			if hash, ok := <-pubsubhubbub; ok {
				s := db.get(hash)
				c.WriteJSON(s)
				log.Println("sent:", hash)
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
				s.FileName = html.EscapeString(s.FileName)
				if len(s.FileName) > 255 {
					s.FileName = s.FileName[:255]
				}

				hash, err := getFileHash(s.MagnetURI)
				if err != nil {
					// send error message
					log.Println("error:", err)
				} else {

					s.CreatedAt = time.Now()

					log.Println("got:", hash)

					db.set(hash, s)
					pubsubhubbub <- hash
				}

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
	db.init()
	pubsubhubbub = make(chan string)
	http.HandleFunc("/", index)
	log.Println("Listening on :12345")
	log.Panicln(http.ListenAndServe(":12345", nil))
}
