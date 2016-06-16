package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/url"
	"regexp"
	"time"

	"github.com/leavengood/websocket"
	"github.com/valyala/fasthttp"

	"github.com/cskr/pubsub"
)

type Payload struct {
	Event   string          `json:"event"`
	Message json.RawMessage `json:"message"`
}
type PayloadGeneric struct {
	Event   string      `json:"event"`
	Message interface{} `json:"message"`
}

type Suprême struct {
	FileName    string    `json:"fileName"`
	MagnetURI   string    `json:"magnetURI"`
	WaveformURI []float64 `json:"waveformURI"`
	Duration    uint64    `json:"duration"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Range struct {
	Start int `json:"start"`
	Limit int `json:"limit"`
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

func index(c *websocket.Conn) {
	log.Println("----------------------------------------------")
	defer c.Close()

	esc := make(chan struct{})
	infohash := ps.Sub("infohash")
	errors := make(chan string)
	go func() {
		for {
			select {
			case <-esc:
				return
			case h := <-infohash:
				if hash, ok := h.(string); ok {
					s := db.get(hash)
					c.WriteJSON(&PayloadGeneric{Event: "suprême", Message: s})
					log.Println("sent:", hash)
				}
			case e := <-errors:
				c.WriteJSON(&PayloadGeneric{Event: "error", Message: e})
				log.Println("sent:", e)
			}
		}
	}()

	sendRange := func(event string, start, limit int) {
		i := 0
		for _, hash := range db.getRange(start, limit) {
			s := db.get(hash)
			c.WriteJSON(&PayloadGeneric{Event: event, Message: s})
			i++
		}
		c.WriteJSON(&PayloadGeneric{Event: "loaded", Message: i})
	}
	// initial load
	sendRange("suprême", 0, 5)

	go func() {
	here:
		for {
			var p Payload
			err := c.ReadJSON(&p)
			log.Printf("%#v\n", p)
			switch err {
			case io.EOF:
				log.Println("reader shutting down")
				break here
			case nil:
				switch p.Event {
				case "suprême":
					var s Suprême
					json.Unmarshal(p.Message, &s)
					if err != nil {
						errors <- err.Error()
						close(esc)
					}
					s.FileName = html.EscapeString(s.FileName)
					if len(s.FileName) > 255 {
						s.FileName = s.FileName[:255]
					}
					hash, err := getFileHash(s.MagnetURI)
					if err != nil {
						errors <- err.Error()
						close(esc)
					} else {
						s.CreatedAt = time.Now()
						log.Println("got:", hash)
						db.set(hash, s)
						ps.Pub(hash, "infohash")
					}
				case "range":
					var r Range
					json.Unmarshal(p.Message, &r)
					if err != nil {
						errors <- err.Error()
						close(esc)
					}
					sendRange("infinitescroll", r.Start, r.Limit)
				}
			default:
				log.Println("error:", err)
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

	log.Println("Listening on :443")
	go fasthttp.ListenAndServeTLS(":443", "cert.pem", "key.pem", fasthttp.FSHandler("../client", 0))

	log.Println("Listening on :12345")
	fasthttp.ListenAndServeTLS(":12345", "cert.pem", "key.pem", func(ctx *fasthttp.RequestCtx) {
		upgrader := websocket.FastHTTPUpgrader{
			CheckOrigin: func(ctx *fasthttp.RequestCtx) bool { return true },
			Handler:     index,
		}
		upgrader.UpgradeHandler(ctx)
	})
}
