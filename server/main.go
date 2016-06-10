package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

func echo(ws *websocket.Conn) {
	io.Copy(ws, ws)
}

type test struct {
	test string `json:"test"`
}

func j(ws *websocket.Conn) {

	dec := json.NewDecoder(ws)
	for {
		var t test
		if err := dec.Decode(&t); err == nil || err == io.EOF {
			log.Printf("got: %#v\n", t)
			if err == io.EOF {
				break
			}
		} else if err != nil {
			log.Fatalln("err:", err.Error())
		}
	}
}

func main() {
	http.Handle("/echo", websocket.Handler(echo))
	http.Handle("/j", websocket.Handler(j))
	log.Println("Listening on :12345")
	log.Panicln(http.ListenAndServe(":12345", nil))
}
