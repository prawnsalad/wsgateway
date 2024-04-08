### wsgateway

`curl -v -d"Broadcast this message" http://localhost:6666/send?admin=1`

Send a message to all connections that have all the tags in the query string. Setting the `Content-Type` header to `application/octet-stream` will send the posted body as a binary websocket message. All other types will be sent as a text message. Returns the number of connections the message was sent to.

`curl -v -d"admin=1" http://localhost:6666/settags?foo=bar`

Set `admin=1` tag on all connections that have all the tags set in the query string. Setting an empty value deletes the tag from the connection. Returns the number of connections the tag was set on.

`curl http://localhost:6666/status`

Get a list of all connections and their tags.

##### Tag selectors
All values in the query string will be used to search for connections with those exact values. Analogous to an "AND" query. Including an `id` value to the query string searches for that specific connection and ignores all other query string values. `id` query string may contain multiple IDs comma delimited (eg. `?id=1,2,3`).


##### Worker files

Include your own .go worker file that extends functionality. Eg:
~~~go
// filename: myworker.go

func init() {
	workersOnBoot = append(workersOnBoot, runDevHelpers)
}

func runDevHelpers(library *connectionlookup.ConnectionLookup) {
	// ...
}
~~~

The included `devhelpers.go-ignore` is an example that when renamed appropriately will run simulated connections and messages, removing the need for real websocket connections to test wsgateway.
