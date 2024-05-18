package streams

import (
	"net/url"
	"regexp"
	"slices"
	"strings"

	"com.wsgateway/connectionlookup"
	"github.com/tidwall/gjson"
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
func makeTagString(con *connectionlookup.Connection) string {
	tags := url.Values{}

	con.KeyValsLock.RLock()
	defer con.KeyValsLock.RUnlock()

	for _, tag := range con.KeyVals {
		if slices.Contains(*con.StreamIncludeTags, tag.Key) {
			tags.Add(tag.Key, tag.KeyVal)
		}
	}

	return tags.Encode()
}

var r, _ = regexp.Compile(`{[a-zA-Z0-9_\-:]+}`)
// Starting with `start`, replace any {variables} in `json` with the values from `vars`
// eg: "action-{command:default}" = "action-default", or if vars contains "command":"join" then "action-join"
func replaceConnectionVars(start string, json string, vars map[string]string) string {
	if !strings.Contains(start, "{") {
		return start
	}

	matches := r.FindAll([]byte(start), -1)
	if matches == nil {
		return start
	}

	for _, match := range matches {
		varName := string(match)[1:len(match)-1]
		defaultVal := "_"

		if strings.Contains(varName, ":") {
			parts := strings.SplitN(varName, ":", 2)
			varName = parts[0]
			defaultVal = parts[1]
		}

		jsonPath := vars[varName]
		value := gjson.Get(json, jsonPath)
		val := value.String()

		// If we didn't find this variable, fallback to an underscore
		if val == "" {
			val = defaultVal
		}
		start = strings.ReplaceAll(start, string(match), val)
	}

	return start
}