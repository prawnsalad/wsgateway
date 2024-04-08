package streams

import (
	"net/url"

	"com.wsgateway/connectionlookup"
)

type Stream interface {
	PublishConnection(con *connectionlookup.Connection, event StreamEvent)
	PublishMessage(con *connectionlookup.Connection, messageType MessageType, message []byte)
}

type StreamEvent string
type MessageType string

func (mt MessageType) String() string {
	if mt == MessageBinary {
		return "binary"
	} else if mt == MessageText {
		return "text"
	}

	return ""
}

func (e StreamEvent) String() string {
	if e == EventOpen {
		return "open"
	} else if e == EventClose {
		return "close"
	} else if e == EventMessage {
		return "message"
	}

	return ""
}

const (
	EventOpen StreamEvent = "open"
	EventClose StreamEvent = "close"
	EventMessage StreamEvent = "message"
	MessageText MessageType = "text"
	MessageBinary MessageType = "binary"
)

// These connection tags will be included in any stream messages
var includeTags = map[string]bool{"foo":true, "group":true}
func makeTagString(con *connectionlookup.Connection) string {
	tags := url.Values{}
	for _, tag := range con.KeyVals {
		_, exists := includeTags[tag.Key]
		if exists {
			tags.Add(tag.Key, tag.KeyVal)
		}
	}

	return tags.Encode()
}