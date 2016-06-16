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

	"github.com/cskr/pubsub"
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

var upgrader = websocket.Upgrader{
	// this gets around this error:
	// "upgrade: websocket: origin not allowed"
	CheckOrigin: func(r *http.Request) bool {
		return true
	}}

func index(w http.ResponseWriter, r *http.Request) {
	log.Println("----------------------------------------------")
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	defer c.Close()

	esc := make(chan struct{})
	infohash := ps.Sub("infohash")
	go func() {
		for {
			select {
			case <-esc:
				return
			case h := <-infohash:
				if hash, ok := h.(string); ok {
					s := db.get(hash)
					c.WriteJSON(s)
					log.Println("sent:", hash)
				}
			}
		}
	}()

	tmp := db.getRange(0, 5)
	db.init()
	for _, hash := range tmp {
		infohash <- hash
	}

	/*
		go func() {
			rand.Seed(time.Now().Unix())
			for {
				select {
				case <-time.After(time.Second * 2):
					ps.Pub(tmp[rand.Intn(len(tmp))], "infohash")
				case <-esc:
					return
				}
			}
		}()
	*/

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
					log.Printf("error: %v\n", err)
					close(esc)
				} else {
					s.CreatedAt = time.Now()
					log.Println("got:", hash)
					db.set(hash, s)
					ps.Pub(hash, "infohash")
				}
			default:
				log.Printf("error: %v\n", err)
				close(esc)
				break here
			}
		}
	}()

	<-esc
}

var db datastore
var ps = pubsub.New(0)

func main() {
	Main()
}
func Main() {
	db.init()
	defer ps.Shutdown()
	http.HandleFunc("/", index)
	log.Println("Listening on :12345")
	http.ListenAndServe(":12345", nil)
}
