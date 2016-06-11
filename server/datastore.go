// replace actual implementation with whatever
package main

import (
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/codykrainock/nanobase" // " A dead-simple flat-file database written in Go. "
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

// temp?

type lmod struct {
	Name    string
	ModTime time.Time
}

type lmodSlice []lmod

func (l lmodSlice) Len() int {
	return len(l)
}
func (l lmodSlice) Less(i, j int) bool {
	return l[i].ModTime.After(l[j].ModTime)
}
func (l lmodSlice) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (db *datastore) getRange(start, limit int) []string {
	dir, err := os.OpenFile(".db", os.O_RDONLY, os.ModeDir)
	checkErr(err)
	files, err := dir.Readdir(-1)
	checkErr(err)
	ls := lmodSlice{}
	for _, file := range files {
		log.Println(file.Name())
		ls = append(ls, lmod{strings.TrimRight(file.Name(), ".json"), file.ModTime()})
	}
	sort.Sort(ls)
	ret := []string{}
	for i, l := range ls {
		if i < start {
			continue
		}
		log.Println(l.Name)
		ret = append(ret, l.Name)
		if i > limit {
			break
		}
	}
	return ret
}
