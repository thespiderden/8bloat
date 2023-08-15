package service

import (
	"io/fs"
	"strings"

	"spiderden.org/8b/conf"
)

type staticfs struct {
	underlying fs.FS
}

func (s *staticfs) Open(path string) (fs.File, error) {
	f, err := s.underlying.Open(s.stripPath(path))
	if err != nil {
		return nil, err
	}

	if stat, _ := f.Stat(); stat.IsDir() {
		f.Close()
		return nil, fs.ErrNotExist
	}

	return f, nil
}

func (s staticfs) stripPath(path string) string {
	cut, _ := strings.CutSuffix(path, conf.Get().AssetStamp)
	return cut
}
