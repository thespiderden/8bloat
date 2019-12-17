package kv

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	errInvalidKey = errors.New("invalid key")
	errNoSuchKey  = errors.New("no such key")
)

type Database struct {
	data    map[string][]byte
	basedir string
	m       sync.RWMutex
}

func NewDatabse(basedir string) (db *Database, err error) {
	err = os.Mkdir(basedir, 0755)
	if err != nil && !os.IsExist(err) {
		return
	}

	return &Database{
		data:    make(map[string][]byte),
		basedir: basedir,
	}, nil
}

func (db *Database) Set(key string, val []byte) (err error) {
	if len(key) < 1 {
		return errInvalidKey
	}

	db.m.Lock()
	defer func() {
		if err != nil {
			delete(db.data, key)
		}
		db.m.Unlock()
	}()

	db.data[key] = val

	err = ioutil.WriteFile(filepath.Join(db.basedir, key), val, 0644)

	return
}

func (db *Database) Get(key string) (val []byte, err error) {
	if len(key) < 1 {
		return nil, errInvalidKey
	}

	db.m.RLock()
	defer db.m.RUnlock()

	data, ok := db.data[key]
	if !ok {
		data, err = ioutil.ReadFile(filepath.Join(db.basedir, key))
		if err != nil {
			err = errNoSuchKey
			return nil, err
		}

		db.data[key] = data
	}

	val = make([]byte, len(data))
	copy(val, data)

	return
}

func (db *Database) Remove(key string) {
	if len(key) < 1 || strings.ContainsRune(key, os.PathSeparator) {
		return
	}

	db.m.Lock()
	defer db.m.Unlock()

	delete(db.data, key)
	os.Remove(filepath.Join(db.basedir, key))

	return
}
