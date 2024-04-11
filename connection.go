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
	idleTimeout     = 60 * time.Second
)

type ConnectionHandlers struct{
	Libray *connectionlookup.ConnectionLookup
	Stream streams.Stream
	SetTags map[string]string
}

func (c *ConnectionHandlers) OnOpen(socket *gws.Conn) {
	_ = socket.SetDeadline(time.Now().Add(idleTimeout))

	id := uuid.NewString()
	con := connectionlookup.NewConnection(id, socket)
	socket.Session().Store("con", con)

	c.Libray.AddConnection(con, c.SetTags)
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
	_ = socket.SetDeadline(time.Now().Add(idleTimeout))
	_ = socket.WritePong(payload)
}

func (c *ConnectionHandlers) OnPong(socket *gws.Conn, payload []byte) {
	_ = socket.SetDeadline(time.Now().Add(idleTimeout))
}

func (c *ConnectionHandlers) OnMessage(socket *gws.Conn, message *gws.Message) {
	_ = socket.SetDeadline(time.Now().Add(idleTimeout))

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