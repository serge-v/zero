package control44

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"sort"
)

var dbname = os.Getenv("HOME") + "/" + "c44.db"

type db struct {
	m map[incident]int
}

func (db *db) dump() {
	var list []incident
	for k := range db.m {
		list = append(list, k)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Dispatched.Before(list[j].Dispatched)
	})
	for _, e := range list {
		fmt.Printf("%+v\n", e)
	}
}

func (db *db) load() {
	db.m = make(map[incident]int)
	f, err := os.Open(dbname)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	dec := gob.NewDecoder(f)
	if err := dec.Decode(&db.m); err != nil {
		log.Fatal(err)
	}
}

func (db *db) get() []incident {
	var list []incident
	for k := range db.m {
		list = append(list, k)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Dispatched.After(list[j].Dispatched)
	})
	return list
}

func (db *db) add(list []incident) {
	idx := len(db.m) + 1
	for _, e := range list {
		_, ok := db.m[e]
		if ok {
			continue
		}
		db.m[e] = idx
		idx++
	}
}

func (db *db) save() {
	tmpname := dbname + ".tmp"
	f, err := os.Create(tmpname)
	if err != nil {
		log.Fatal(err)
	}
	enc := gob.NewEncoder(f)
	err = enc.Encode(db.m)
	if err != nil {
		log.Fatal(err)
	}
	err = f.Close()
	if err != nil {
		log.Fatal(err)
	}
	if err := os.Rename(tmpname, dbname); err != nil {
		log.Fatal(err)
	}
}
