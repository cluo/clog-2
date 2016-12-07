package main

import (
	"encoding/binary"
	"github.com/tecbot/gorocksdb"
)

var (
	dbs    map[string]*gorocksdb.DB
	dbOpts *gorocksdb.Options
	ro     *gorocksdb.ReadOptions
	wo     *gorocksdb.WriteOptions
)

func init() {
	dbs = make(map[string]*gorocksdb.DB)
	bbto := gorocksdb.NewDefaultBlockBasedTableOptions()
	bbto.SetBlockCache(gorocksdb.NewLRUCache(50 * 1024 * 1024)) // 50 MB
	dbOpts = gorocksdb.NewDefaultOptions()
	dbOpts.SetBlockBasedTableFactory(bbto)
	dbOpts.SetCreateIfMissing(true)
	ro = gorocksdb.NewDefaultReadOptions()
	wo = gorocksdb.NewDefaultWriteOptions()
}

func getDB(channel string) *gorocksdb.DB {
	db, ok := dbs[channel]
	if !ok {
		var err error
		db, err = gorocksdb.OpenDb(dbOpts, "dbs/"+channel)
		if err != nil {
			panic(err)
		}
		dbs[channel] = db
	}
	return db
}

func writeWord(channel string, date int64, lineNumber int, word string) {
	db := getDB(channel)
	buf := make([]byte, 12)
	binary.LittleEndian.PutUint64(buf, uint64(date))
	binary.LittleEndian.PutUint32(buf[8:], uint32(lineNumber))
	db.Put(wo, []byte(word), buf)
}
