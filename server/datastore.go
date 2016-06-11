// replace actual implementation with whatever
package main

import (
	"github.com/codykrainock/nanobase" // " A dead-simple flat-file database written in Go. "
	"log"
)

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

type datastore struct {
	conn *nanobase.Nanobase
}

func (db *datastore) init() {
	c, err := nanobase.Connect(".db")
	checkErr(err)
	db.conn = c
}

func (db *datastore) get(key string) Suprême {
	var val Suprême
	checkErr(db.conn.Get(key, &val))
	return val
}

func (db *datastore) set(key string, val Suprême) {
	checkErr(db.conn.Put(key, val))
}
