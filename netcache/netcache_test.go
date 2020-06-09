package netcache

import (
	"fmt"
	"gotest.tools/assert"
	"testing"
	"time"
)

const ConnectionTimeout = time.Second * 5

func isready(nc *NetCache, peers ...PeerAddress) bool {
	for _,p := range peers {
		if c,ok := nc.cache[p]; ok {
			select {
				case _, ok = <- c.ready:
				default:
					return false
			}
		} else {
			return false
		}
	}
	return true
}

func isnotready(nc *NetCache, peers ...PeerAddress) bool {
	for _,p := range peers {
		if c,ok := nc.cache[p]; ok {
			select {
			case _, ok = <- c.ready:
				return false
			default:
				break
			}
		}
	}
	return true
}

type TestConnectionFactory struct {}
type TestConnection struct {peer PeerAddress}

func (f *TestConnectionFactory) New(peer PeerAddress) Connection {
	return &TestConnection{peer}
}

func (c *TestConnection) Open() {
	fmt.Println("open ", c.peer)
	time.Sleep(ConnectionTimeout)
	fmt.Println("opened ", c.peer)
}

func (c *TestConnection) Close() {
	fmt.Println("colse ", c.peer)
}

func Test_OneConnection(t *testing.T) {
	nc := New(&TestConnectionFactory{})
	nc.GetConnection(1)
	nc.Shutdown()
}

func Test_ConcurrentConnections(t *testing.T) {
	nc := New(&TestConnectionFactory{})
	go func() {nc.GetConnection(1)}()
	go func() {nc.GetConnection(2)}()
	go func() {nc.GetConnection(2)}()
	time.Sleep(ConnectionTimeout+time.Second)
	assert.Assert(t,isready(nc,1,2))
	nc.Shutdown()
}

func Test_Callback(t *testing.T) {
	nc := New(&TestConnectionFactory{})
	go func() {nc.GetConnection(1)} ()
	go func() {nc.GetConnection(2)} ()
	go func() {nc.GetConnection(3)} ()
	time.Sleep(time.Millisecond*100)
	nc.OnNewRemoteConnection(2,&TestConnection{2})
	assert.Assert(t,isready(nc,2))
	nc.OnNewRemoteConnection(1,&TestConnection{1})
	assert.Assert(t,isready(nc,1))
	assert.Assert(t,isnotready(nc,3))
	fmt.Println("shutdown")
	nc.Shutdown()
}
