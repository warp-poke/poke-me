package core

import (
	"strings"
	"sync"
	"time"

	"github.com/samuel/go-zookeeper/zk"
	"github.com/spf13/viper"
)

// ZK handle zk state
type ZK struct {
	conn *zk.Conn
}

// ZNode handle a znode state
type ZNode struct {
	conn    *zk.Conn
	path    string
	Values  chan []byte
	version int32
	err     error
	lock    sync.RWMutex
}

// NewZK return a wk connection
func NewZK(servers []string, timeout time.Duration) (*ZK, error) {
	conn, _, err := zk.Connect(viper.GetStringSlice("zk.servers"), timeout)
	if err != nil {
		return nil, err
	}

	zk := ZK{
		conn: conn,
	}

	return &zk, nil
}

// ZNode return a znode handler
func (_zk *ZK) ZNode(path string) (*ZNode, error) {
	dirs := strings.Split(path, "/")
	for idx := range dirs {
		// Skip first index - aka not ensure /
		if idx == 0 {
			continue
		}

		node := strings.Join(dirs[0:idx+1], "/")

		exist, _, err := _zk.conn.Exists(node)
		if err != nil {
			return nil, err
		}

		if !exist {
			flags := int32(0)
			acl := zk.WorldACL(zk.PermAll)
			_, err := _zk.conn.Create(node, []byte(""), flags, acl)
			if err != nil {
				return nil, err
			}
		}
	}

	znode := ZNode{
		conn:   _zk.conn,
		path:   path,
		Values: make(chan []byte),
	}

	_, stat, err := _zk.conn.Get(znode.path)
	if err != nil {
		return nil, err
	}
	znode.version = stat.Version

	go func() {
		for {
			znode.lock.RLock()
			v, stat, event, err := _zk.conn.GetW(znode.path)

			if err != nil {
				znode.err = err
				znode.lock.RUnlock()
				return
			}

			znode.version = stat.Version
			znode.lock.RUnlock()
			znode.Values <- v

			<-event
		}
	}()
	return &znode, nil
}

// Update set a new znode value
func (_znode *ZNode) Update(value []byte) error {
	_znode.lock.Lock()
	defer _znode.lock.Unlock()

	_, err := _znode.conn.Set(_znode.path, value, _znode.version)
	if err != nil {
		return err
	}

	return nil
}
