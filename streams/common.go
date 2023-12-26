package streams

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