package netcache

import (
	"sync"
)

type Connection interface {
	Open()
	Close()
}

type ConnectionFactory interface {
	New(address PeerAddress) Connection
}

type PeerAddress int32

type ConnectionHolder struct {
	mu sync.Mutex
	peer PeerAddress
	connection Connection
	ready chan struct{}
}

func (h *ConnectionHolder) Open(factory ConnectionFactory) {
	c := factory.New(h.peer)
	c.Open()
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.connection != nil {
		c.Close()
	} else {
		h.connection = c
		close(h.ready)
	}
}

type NetCache struct {
	mu sync.Mutex
	cache map[PeerAddress]*ConnectionHolder
	factory ConnectionFactory
	shutdown bool
}

func New(factory ConnectionFactory) *NetCache {
	return &NetCache{
		cache: map[PeerAddress]*ConnectionHolder{},
		factory: factory,
	}
}

/*
  GetConnection get connection or create one and store it,
  this caller should be blocked until the connection is returned. Should be thread safe.
*/
func (nc *NetCache) GetConnection(peer PeerAddress) Connection {
	nc.mu.Lock()
	if nc.shutdown {
		nc.mu.Unlock()
		panic("network cache is not available")
	}
	c, ok := nc.cache[peer]
	if !ok {
		c = &ConnectionHolder{ready: make(chan struct{}),peer:peer}
		nc.cache[peer] = c
		go c.Open(nc.factory)
	}
	nc.mu.Unlock()
	<-c.ready
	return c.connection
}

/*
  A callback function that is called whenever a remote peer establishes a connection
  with the local node. Should be threadsafe.
*/
func (nc *NetCache) OnNewRemoteConnection(peer PeerAddress, connection Connection) {
	nc.mu.Lock()
	if nc.shutdown {
		nc.mu.Unlock()
		panic("network cache is not available")
	}
	c, ok := nc.cache[peer]
	if !ok {
		c = &ConnectionHolder{ready: make(chan struct{}),peer:peer}
		nc.cache[peer] = c
	}
	nc.mu.Unlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.connection != nil {
		connection.Close()
	} else {
		c.connection = connection
		close(c.ready)
	}
}

/*
  Shutdown does graceful shutdown, all background workers should be stopped before this method returns
*/
func (nc *NetCache) Shutdown() {
	nc.mu.Lock()
	nc.shutdown = true
	wg := sync.WaitGroup{}
	wg.Add(len(nc.cache))
	for _,c := range nc.cache {
		go func(ready chan struct{}) {
			<-ready
			wg.Done()
		}(c.ready)
	}
	nc.mu.Unlock()
	wg.Wait()
	for _,c := range nc.cache {
		if c.connection != nil {
			c.connection.Close()
		}
	}
}



