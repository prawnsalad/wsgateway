package main

import (
	"log"
	"time"

	"com.wsgateway/connectionlookup"
	"com.wsgateway/streams"
	"github.com/google/uuid"
	"github.com/lxzan/gws"
)

const (
	PingInterval = 5 * time.Second
	//PingWait     = 10 * time.Second
	PingWait     = 10 * time.Minute
)

type ConnectionHandlers struct{
	Libray *connectionlookup.ConnectionLookup
	Stream *streams.StreamRedis
}

func (c *ConnectionHandlers) OnOpen(socket *gws.Conn) {
	_ = socket.SetDeadline(time.Now().Add(PingInterval + PingWait))

	id := uuid.NewString()
	con := &connectionlookup.Connection{
		Id: id,
		Socket: socket,
		KeyVals: make(map[string]*connectionlookup.ConnectionLockList),
	}
	socket.Session().Store("con", con)

	c.Libray.AddConnection(con, map[string]string{
		"foo": "bar",
	})

	c.Stream.PublishConnection(con, streams.EventOpen)
}

func (c *ConnectionHandlers) OnClose(socket *gws.Conn, err error) {
	storeCon, isOk := socket.Session().Load("con")
	if !isOk {
		log.Println("Error: Socket missing connection instance")
		return
	}
	con := storeCon.(*connectionlookup.Connection)

	c.Libray.RemoveConnection(con)
	c.Stream.PublishConnection(con, streams.EventClose)
}

func (c *ConnectionHandlers) OnPing(socket *gws.Conn, payload []byte) {
	_ = socket.SetDeadline(time.Now().Add(PingInterval + PingWait))
	_ = socket.WritePong(nil)
}

func (c *ConnectionHandlers) OnPong(socket *gws.Conn, payload []byte) {}

func (c *ConnectionHandlers) OnMessage(socket *gws.Conn, message *gws.Message) {
	storeCon, isOk := socket.Session().Load("con")
	if !isOk {
		log.Println("Error: Socket missing connection instance")
		return
	}
	con := storeCon.(*connectionlookup.Connection)

	mType := streams.MessageText
	if message.Opcode == gws.OpcodeBinary {
		mType = streams.MessageBinary
	}

	c.Stream.PublishMessage(con, mType, message.Bytes())
	message.Close()
}