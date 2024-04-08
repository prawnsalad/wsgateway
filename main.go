package main

import (
	"log"
	"net/http"
	"runtime"

	"com.wsgateway/connectionlookup"
	"com.wsgateway/streams"
)

var config *Config

// adhoc .go workers can append themselves here during init() to extend functionality
var workersOnBoot = []func(*connectionlookup.ConnectionLookup){}

func main() {
	log.Printf("Starting wsgateway. GOMAXPROCS=%d NumCPU=%d", runtime.GOMAXPROCS(0), runtime.NumCPU())

	loadedConfig, err := loadConfig(`
listen_addr: 0.0.0.0:5000
connection_redis_sync:
  addr: redis://localhost:6379/0?client_name=wsgateway
stream_redis:
  addr: redis://localhost:6379/0?client_name=wsgatewaystream&pool_size=1000
endpoints:
  - path: /connect
    set_tags:
      foo: bar
      other: tag
    stream_include_tags:
      - foo
      - group

  - path: /connect/v2
    set_tags:
      version: 2
    stream_include_tags:
      - version
      - group
`)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	config = loadedConfig

	startHttpServer()
}

func startHttpServer() {
	library, err := connectionlookup.NewConnectionLookup(config.ConnectionRedisSync.Addr)
	if err != nil {
		log.Fatal("Error starting: ", err.Error())
	}

	stream, err := streams.NewStreamRedis(config.StreamRedis.Addr)
	if err != nil {
		log.Fatal("Error starting: ", err.Error())
	}

	// Allow workers to boot up with the library instance
	for _, worker := range workersOnBoot {
		worker(library)
	}

	applyWsHandlers(library, stream)
	applyHttpHandlers(library, stream)

	listenStr := config.ListenAddr
	log.Printf("Listening on %s", listenStr)
	http.ListenAndServe(listenStr, nil)
}

