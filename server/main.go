package main

import (
	"io"
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

func echo(ws *websocket.Conn) {
	io.Copy(ws, ws)
}

type test struct {
	value int `json:"test"`
}

func j(ws *websocket.Conn) {
	var t test
	for {
		websocket.JSON.Receive(ws, &t)
		log.Printf("got: %#v\n", t.value)
	}
}

func main() {
	http.Handle("/echo", websocket.Handler(echo))
	http.Handle("/j", websocket.Handler(j))
	log.Println("Listening on :12345")
	log.Panicln(http.ListenAndServe(":12345", nil))
}
