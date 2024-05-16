package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"com.wsgateway/connectionlookup"
	"com.wsgateway/streams"
	"github.com/lxzan/gws"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func applyWsHandlers(library *connectionlookup.ConnectionLookup, stream streams.Stream) {
	for _, c := range config.Endpoints {

		// individual endpoint MaxMessageSizeKb takes precedence over global
		maxPayloadSize := 0
		if c.MaxMessageSizeKb > 0 {
			maxPayloadSize = c.MaxMessageSizeKb * 1024
		} else {
			maxPayloadSize = config.MaxMessageSizeKb * 1024
		}

		applyWsEndpointHandlers(&EndpointConfig{
			Path:               c.Path,
			SetTags:            c.SetTags,
			StreamIncludeTags:  c.StreamIncludeTags,
			ReadMaxPayloadSize: maxPayloadSize,
			JsonExtractVars:    c.JsonExtractVars,
		}, library, stream)
	}
}

type EndpointConfig struct {
	Path               string
	SetTags            map[string]string
	StreamIncludeTags  []string
	ReadMaxPayloadSize int
	JsonExtractVars    map[string]string
}

func applyWsEndpointHandlers(conf *EndpointConfig, library *connectionlookup.ConnectionLookup, stream streams.Stream) {
	log.Printf("Creating websocket endpoint at path %s", conf.Path)

	wsHandlers := &ConnectionHandlers{
		Libray:          library,
		Stream:          stream,
		SetTags:         conf.SetTags,
		JsonExtractVars: conf.JsonExtractVars,
	}
	upgrader := gws.NewUpgrader(wsHandlers, &gws.ServerOption{
		ReadAsyncEnabled:   true,         // Parallel message processing
		CompressEnabled:    true,         // Enable compression
		Recovery:           gws.Recovery, // Exception recovery
		ReadMaxPayloadSize: conf.ReadMaxPayloadSize,
	})

	http.HandleFunc(conf.Path, func(writer http.ResponseWriter, request *http.Request) {
		socket, err := upgrader.Upgrade(writer, request)
		if err != nil {
			return
		}
		go func() {
			socket.ReadLoop() // Blocking prevents the context from being GC.
		}()
	})
}

func applyHttpHandlers(library *connectionlookup.ConnectionLookup, stream streams.Stream) {
	internal := http.NewServeMux()
	http.Handle("/", middlewareAuthInternal(internal))

	internal.Handle("/metrics", promhttp.Handler())

	internal.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Request: /status")

		w.Header().Add("Content-Type", "text/plain")

		log.Println("Calling dump connection...")
		dump := library.DumpConnections()
		log.Println("...", len(dump))
		for _, con := range dump {
			// bring the id tag to the tasrt of the line just for ease of readability
			w.Write([]byte("id=" + con["id"] + " "))

			for key, val := range con {
				if key != "id" {
					w.Write([]byte(key + "=" + val + " "))
				}
			}
			w.Write([]byte("\n"))
		}
	})

	internal.HandleFunc("/tree", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Request: /tree")

		w.Header().Add("Content-Type", "text/plain")

		log.Println("Calling GetAllKeysAndValue...")
		dump := library.GetAllKeysAndValue()
		for key, vals := range dump {
			// bring the id tag to the tasrt of the line just for ease of readability
			w.Write([]byte(key + "\n"))

			for _, val := range vals {
				w.Write([]byte(" - " + val + "\n"))
			}
			w.Write([]byte("\n"))
		}
	})

	internal.HandleFunc("/settags", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			return
		}

		cons := getConsFromQueryStringVals(library, r)
		if len(cons) == 0 {
			w.Write([]byte("0"))
			return
		}

		postBody, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("Error reading post body in settags request: " + err.Error())
			w.WriteHeader(503)
			return
		}

		postedNewTags, err := url.ParseQuery(string(postBody))
		if err != nil {
			log.Println("Error parsing posted tags in settags request: " + err.Error())
			w.WriteHeader(503)
			return
		}

		newTags := map[string]string{}
		for k, v := range postedNewTags {
			newTags[k] = v[0]
		}
		for _, con := range cons {
			library.SetKeys(con, newTags)
		}

		w.Write([]byte(strconv.Itoa(len(cons))))
	})

	internal.HandleFunc("/sendAndClose", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			return
		}

		cons := getConsFromQueryStringVals(library, r)
		log.Println("/sendAndClose Number of connections matched: " + strconv.Itoa(len(cons)))
		if len(cons) != 1 {
			log.Println("Attempt to send and close a socket without specifying a single socket")
			w.WriteHeader(400)
			return
		}

		con := cons[0]

		data, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("Error reading post body in send request: " + err.Error())
			w.WriteHeader(503)
			return
		}

		wsOpcode := gws.OpcodeText
		if r.Header.Get("Content-Type") == "application/octet-stream" {
			wsOpcode = gws.OpcodeBinary
		}

		sendMessageAndClose(con.Socket, wsOpcode, data)
	})

	internal.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			return
		}

		cons := getConsFromQueryStringVals(library, r)
		log.Println("Number of connections matched: " + strconv.Itoa(len(cons)))
		if len(cons) == 0 {
			w.Write([]byte("0"))
			return
		}

		// TODO: Find a way to stream this to websockets if it's large so we
		//       don't use up all memory
		data, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("Error reading post body in send request: " + err.Error())
			w.WriteHeader(503)
			return
		}

		wsOpcode := gws.OpcodeText
		if r.Header.Get("Content-Type") == "application/octet-stream" {
			wsOpcode = gws.OpcodeBinary
		}

		sendMessageToConnections(cons, wsOpcode, data)
		w.Write([]byte(strconv.Itoa(len(cons))))
	})
}

func middlewareAuthInternal(handler http.Handler) http.Handler {
	// If unauthorized, purposelly don't return anything, just ignore the request
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			log.Printf("Unauthorized access attempt from an unknown host: %s", host)
			return
		}

		remoteAddr := net.ParseIP(host)
		for _, internalRange := range config.InternalEndpointWhitelistInet {
			if internalRange.Contains(remoteAddr) {
				handler.ServeHTTP(w, r)
				return
			}
		}
	})
}

func getConsFromQueryStringVals(library *connectionlookup.ConnectionLookup, r *http.Request) []*connectionlookup.Connection {
	var cons []*connectionlookup.Connection

	if r.URL.Query().Get("id") != "" {
		ids := strings.Split(r.URL.Query().Get("id"), ",")
		for _, id := range ids {
			con, exists := library.GetConnectionById(id)
			if exists {
				cons = append(cons, con)
			}
		}
	} else {
		searchTags := map[string]string{}
		for k, v := range r.URL.Query() {
			searchTags[k] = v[0]
		}
		cons = library.GetConnectionsWithKeys(searchTags)
	}

	return cons
}
