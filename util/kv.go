package util

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
	cache   map[string][]byte
	basedir string
	m       sync.RWMutex
}

func NewDatabse(basedir string) (db *Database, err error) {
	err = os.Mkdir(basedir, 0755)
	if err != nil && !os.IsExist(err) {
		return
	}

	return &Database{
		cache:   make(map[string][]byte),
		basedir: basedir,
	}, nil
}

func (db *Database) Set(key string, val []byte) (err error) {
	if len(key) < 1 || strings.ContainsRune(key, os.PathSeparator) {
		return errInvalidKey
	}

	err = ioutil.WriteFile(filepath.Join(db.basedir, key), val, 0644)
	if err != nil {
		return
	}

	db.m.Lock()
	db.cache[key] = val
	db.m.Unlock()

	return
}

func (db *Database) Get(key string) (val []byte, err error) {
	if len(key) < 1 || strings.ContainsRune(key, os.PathSeparator) {
		return nil, errInvalidKey
	}

	db.m.RLock()
	data, ok := db.cache[key]
	db.m.RUnlock()

	if !ok {
		data, err = ioutil.ReadFile(filepath.Join(db.basedir, key))
		if err != nil {
			err = errNoSuchKey
			return nil, err
		}

		db.m.Lock()
		db.cache[key] = data
		db.m.Unlock()
	}

	val = make([]byte, len(data))
	copy(val, data)

	return
}

func (db *Database) Remove(key string) {
	if len(key) < 1 || strings.ContainsRune(key, os.PathSeparator) {
		return
	}

	os.Remove(filepath.Join(db.basedir, key))

	db.m.Lock()
	delete(db.cache, key)
	db.m.Unlock()

	return
}
