package conf

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"sync"

	"github.com/bwmarrin/snowflake"
)

var (
	node     *snowflake.Node
	nodelock sync.RWMutex
)

func ID() string {
	nodelock.RLock()
	defer nodelock.RUnlock()
	idSlice := &bytes.Buffer{}
	binary.Write(idSlice, binary.LittleEndian, node.Generate())
	return base64.RawURLEncoding.EncodeToString(idSlice.Bytes())
}
