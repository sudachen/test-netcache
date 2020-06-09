package netcache

import (
	"fmt"
	"testing"
	"time"
)

type TestConnectionFactory struct {}
type TestConnection struct {peer PeerAddress}

func (f *TestConnectionFactory) New(peer PeerAddress) Connection {
	return &TestConnection{peer}
}

func (c *TestConnection) Open() {
	fmt.Println("open ", c.peer)
	time.Sleep(5*time.Second)
	fmt.Println("opened ", c.peer)
}

func (c *TestConnection) Close() {
	fmt.Println("colse ", c.peer)
}

func Test_OneConnection(t *testing.T) {
	nc := New(&TestConnectionFactory{})
	go func() {nc.GetConnection(1)}()
	nc.GetConnection(2)
	go func() {nc.GetConnection(3)}()
	time.Sleep(300*time.Millisecond)
	nc.OnNewRemoteConnection(2,&TestConnection{2})
	nc.Shutdown()
}
