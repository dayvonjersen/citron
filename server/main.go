package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"io"
	"log"
	"time"

	"github.com/leavengood/websocket"
	"github.com/valyala/fasthttp"

	"github.com/cskr/pubsub"
)

//
// JSON sent on both ends of the websocket takes the following form:
//
// {"event": (eventName), "message": (arbitraryData)}
//

// Unmarshaling into this struct lets us read the Event
// and Unmarshal Message into the appropriate struct
type Payload struct {
	Event   string          `json:"event"`
	Message json.RawMessage `json:"message"`
}

// Marshaling into this struct sends appropriately formatted data
type PayloadGeneric struct {
	Event   string      `json:"event"`
	Message interface{} `json:"message"`
}

// A Suprême represents an audio
type Suprême struct {
	FileName    string    `json:"fileName"`
	MagnetURI   string    `json:"magnetURI"`
	WaveformURI []float64 `json:"waveformURI"`
	Duration    uint64    `json:"duration"`
	CreatedAt   time.Time `json:"createdAt"`
}

// A Range represents a request for more audio
type Range struct {
	Start int `json:"start"`
	Limit int `json:"limit"`
}

// The route of the websocket
func index(c *websocket.Conn) {
	defer c.Close()

	// is used for initial load and additional requests for infinite scroll
	sendRange := func(event string, start, limit int) {
		i := 0
		for _, hash := range db.getRange(start, limit) {
			s := db.get(hash)
			c.WriteJSON(&PayloadGeneric{Event: event, Message: s})
			log.Println("sent:", hash)
			i++
		}
		c.WriteJSON(&PayloadGeneric{Event: "loaded", Message: i})
	}

	// initial load
	sendRange("suprême", 0, 5)

	exitSender := make(chan struct{})
	exitIndex := make(chan struct{})
	hashChan := ps.Sub("infohash")
	errorChan := make(chan error)

	// sender
	go func() {
		for {
			select {
			case h := <-hashChan:
				if hash, ok := h.(string); ok {
					s := db.get(hash)
					c.WriteJSON(&PayloadGeneric{Event: "suprême", Message: s})
					log.Println("sent:", hash)
				}
			case e := <-errorChan:
				c.WriteJSON(&PayloadGeneric{Event: "error", Message: e.Error()})
				log.Println("sent:", e)
			case <-exitSender:
				close(exitIndex)
				return
			}
		}
	}()

	// receiver
	go func() {
		for {
			var p Payload
			err := c.ReadJSON(&p)
			switch err {
			case nil:
				switch p.Event {
				case "suprême":
					var s Suprême
					err := json.Unmarshal(p.Message, &s)
					if err != nil {
						errorChan <- err
						close(exitSender)
						return
					}
					s.FileName = html.EscapeString(s.FileName)
					if len(s.FileName) > 255 {
						s.FileName = s.FileName[:255]
					}
					hash, err := GetInfoHash(s.MagnetURI)
					if err != nil {
						errorChan <- err
						close(exitSender)
						return
					} else {
						s.CreatedAt = time.Now()
						log.Println("got:", hash)
						db.set(hash, s)
						ps.Pub(hash, "infohash")
					}
				case "range":
					var r Range
					err := json.Unmarshal(p.Message, &r)
					if err != nil {
						errorChan <- err
						close(exitSender)
						return
					}
					sendRange("infinitescroll", r.Start, r.Limit)
				}
			case io.EOF:
				log.Println("reader shutting down")
				close(exitSender)
				return
			default:
				log.Println("error:", err)
				close(exitSender)
				return
			}
		}
	}()

	<-exitIndex
}

// generic key-value store
var db datastore

// generic publish-subscribe
var ps = pubsub.New(100)

func main() {
	// you do this in order to be able to defer or something.
	Main()
}

func Main() {
	var (
		bindAddr      string
		httpPort      int
		documentRoot  string
		websocketPort int
		certFile      string
		keyFile       string
	)
	flag.StringVar(&bindAddr, "bind-addr", "", "leave blank for 0.0.0.0")
	flag.IntVar(&httpPort, "http-port", 443, "")
	flag.StringVar(&documentRoot, "document-root", "../client", "path to fs")
	flag.IntVar(&websocketPort, "websocket-port", 12345, "")
	flag.StringVar(&certFile, "cert", "cert.pem", "path to cert")
	flag.StringVar(&keyFile, "key", "key.pem", "path to key")
	flag.Parse()

	db.init()
	defer ps.Shutdown()

	// redirect http traffic to https
	go fasthttp.ListenAndServe(bindAddr+":80", func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.SetStatusCode(fasthttp.StatusTemporaryRedirect) // Use StatusMovedPermanently  if you know what you doing
		ctx.Response.Header.Set("Location", "https://"+string(ctx.Request.Host())+string(ctx.Path()))
	})

	bindHttp := fmt.Sprintf("%s:%d", bindAddr, httpPort)
	log.Println("    HTTPS Listening on", bindHttp)
	go func() {
		log.Fatal(fasthttp.ListenAndServeTLS(bindHttp, certFile, keyFile, fasthttp.FSHandler(documentRoot, 0)))
	}()

	bindWebsocket := fmt.Sprintf("%s:%d", bindAddr, websocketPort)
	log.Println("Websocket Listening on", bindWebsocket)
	log.Fatal(fasthttp.ListenAndServeTLS(bindWebsocket, certFile, keyFile, func(ctx *fasthttp.RequestCtx) {
		upgrader := websocket.FastHTTPUpgrader{
			CheckOrigin: func(ctx *fasthttp.RequestCtx) bool { return true },
			Handler:     index,
		}
		upgrader.UpgradeHandler(ctx)
	}))
}
