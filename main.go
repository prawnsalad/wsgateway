package main

import (
	"log"
	"net/http"
	"runtime"

	"com.wsgateway/connectionlookup"
	"com.wsgateway/streams"
)

func main() {
	log.Printf("Starting wsgateway. GOMAXPROCS=%d NumCPU=%d", runtime.GOMAXPROCS(0), runtime.NumCPU())
	startHttpServer()
}

func startHttpServer() {
	// "redis://user:password@localhost:6379/0?protocol=3&client_name=wsgateway"
	library, err := connectionlookup.NewConnectionLookup("redis://host.orb.internal:6379/0?client_name=wsgateway")
	if err != nil {
		log.Fatal("Error starting: ", err.Error())
	}

	stream, err := streams.NewStreamRedis("redis://host.orb.internal:6379/0?client_name=wsgatewaystream&pool_size=1000")
	if err != nil {
		log.Fatal("Error starting: ", err.Error())
	}

	// TODO: remove these dev helpers
	runDevHelpers(library)

	applyHttpHandlers(library, stream)

	listenStr := "0.0.0.0:6666"
	log.Printf("Listening on %s", listenStr)
	http.ListenAndServe(listenStr, nil)
}

