# wsgateway

## Building and testing

~~~shell
# Build wsgateway for your local system
$ make build

# Build wsgateway for various systems including 32/64 bit darwin/linux/windows
$ make build-all

# Run tests
$ make test
go test -v ./...
=== RUN   TestAddConnection
--- PASS: TestAddConnection (0.00s)
[...]
PASS
ok      com.wsgateway/connectionlookup  (cached)

# Run from code
$ make run
go run .
2024/04/08 22:21:15 Starting wsgateway. CONFIG=config.yml NOFILES=1048576 GOMAXPROCS=10 NUMCPU=10
2024/04/08 22:21:15 Connecting to redis for connection data at localhost:6379
2024/04/08 22:21:15 Connecting to redis for streaming at localhost:6379
2024/04/08 22:21:15 Creating websocket endpoint at path /connect
2024/04/08 22:21:15 Creating websocket endpoint at path /connect/v2
2024/04/08 22:21:15 Listening on 0.0.0.0:5000
~~~

On startup important bits of information is reported that will impact your system:
- CONFIG - The config file being loaded
- NOFILES - The max number of open connections wsgateway can open. Change this via your OS ulimits
- GOMAXPROCS - Max number of goroutines able to be used. Usually the same number of CPU cores
- NUMCPU - Number of CPU cores available. More cores = more concurrency


## Usage
`curl -v -d"Broadcast this message" http://localhost:6666/send?admin=1`

Send a message to all connections that have all the tags in the query string. Setting the `Content-Type` header to `application/octet-stream` will send the posted body as a binary websocket message. All other types will be sent as a text message. Returns the number of connections the message was sent to.

`curl -v -d"admin=1" http://localhost:6666/settags?foo=bar`

Set `admin=1` tag on all connections that have all the tags set in the query string. Setting an empty value deletes the tag from the connection. Returns the number of connections the tag was set on.

`curl http://localhost:6666/status`

Get a list of all connections and their tags.

#### Tag selectors
All values in the query string will be used to search for connections with those exact values. Analogous to an "AND" query. Including an `id` value to the query string searches for that specific connection and ignores all other query string values. `id` query string may contain multiple IDs comma delimited (eg. `?id=1,2,3`).



## Development

#### Worker files

Build in your own .go worker file that extends functionality. Eg:
~~~go
// filename: myworker.go

func init() {
	workersOnBoot = append(workersOnBoot, runDevHelpers)
}

func runDevHelpers(library *connectionlookup.ConnectionLookup) {
	// ...
}
~~~

The included `worker-devhelpers.go-ignore` is an example that when renamed appropriately will run simulated connections and messages, removing the need for real websocket connections to test wsgateway.
