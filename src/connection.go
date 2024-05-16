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
	idleTimeout = 60 * time.Second
)

type ConnectionHandlers struct {
	Libray            *connectionlookup.ConnectionLookup
	Stream            streams.Stream
	SetTags           map[string]string
	StreamIncludeTags []string
	JsonExtractVars   map[string]string
}

func (c *ConnectionHandlers) OnOpen(socket *gws.Conn) {
	counterConnections.Inc()

	_ = socket.SetDeadline(time.Now().Add(idleTimeout))

	id := uuid.NewString()
	con := connectionlookup.NewConnection(id, socket)
	con.StreamIncludeTags = c.StreamIncludeTags
	con.JsonExtractVars = &c.JsonExtractVars
	socket.Session().Store("con", con)

	c.Libray.AddConnection(con, c.SetTags)
	c.Stream.PublishConnection(con, streams.EventOpen)
}

func (c *ConnectionHandlers) OnClose(socket *gws.Conn, err error) {
	counterDisconnections.Inc()

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
	counterClientRecievedMsgs.Inc()

	_ = socket.SetDeadline(time.Now().Add(idleTimeout))

	storeCon, isOk := socket.Session().Load("con")
	if !isOk {
		log.Println("Error: Socket missing connection instance")
		return
	}
	con := storeCon.(*connectionlookup.Connection)

	log.Printf("Received message", len(con.KeyVals), message)

	mType := streams.MessageText
	if message.Opcode == gws.OpcodeBinary {
		mType = streams.MessageBinary
	}

	c.Stream.PublishMessage(con, mType, message.Bytes())
	message.Close()
}

func sendMessageAndClose(socket *gws.Conn, payloadType gws.Opcode, payload []byte) {
	counterDisconnections.Inc()
	log.Println("Sending message and closing")
	socket.WriteMessage(payloadType, payload)

	socket.WriteClose(1000, []byte("Closing socket"))
}

func sendMessageToConnections(conns []*connectionlookup.Connection, payloadType gws.Opcode, payload []byte) {
	broadcaster := gws.NewBroadcaster(payloadType, payload)
	defer broadcaster.Close()

	for _, con := range conns {
		broadcaster.Broadcast(con.Socket)
	}

	counterClientSentMsgs.Add(float64(len(conns)))
}
