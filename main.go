package main

import (
	"flag"
	"log"
	"net/http"
	"runtime"

	"com.wsgateway/connectionlookup"
	"com.wsgateway/streams"
)

var configPath string
var config *Config

// adhoc .go workers can append themselves here during init() to extend functionality
var workersOnBoot = []func(*connectionlookup.ConnectionLookup){}

func main() {
	flag.StringVar(&configPath, "config", "./config.yml", "Configuration file path")
	flag.Parse()
	
	log.Printf("Starting wsgateway. config=%s GOMAXPROCS=%d NumCPU=%d", configPath, runtime.GOMAXPROCS(0), runtime.NumCPU())

	loadedConfig, err := loadConfigFromFile(configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	config = loadedConfig

	library, stream := initComponents()
	runWorkers(library)
	startHttpServer(library, stream)
}

func initComponents() (*connectionlookup.ConnectionLookup, *streams.StreamRedis){
	library, err := connectionlookup.NewConnectionLookup(config.ConnectionRedisSync.Addr)
	if err != nil {
		log.Fatal("Error starting: ", err.Error())
	}

	stream, err := streams.NewStreamRedis(config.StreamRedis.Addr)
	if err != nil {
		log.Fatal("Error starting: ", err.Error())
	}

	return library, stream
}

func runWorkers(library *connectionlookup.ConnectionLookup) {
	for _, worker := range workersOnBoot {
		worker(library)
	}
}

func startHttpServer(library *connectionlookup.ConnectionLookup, stream *streams.StreamRedis) {
	applyWsHandlers(library, stream)
	applyHttpHandlers(library, stream)

	listenStr := config.ListenAddr
	log.Printf("Listening on %s", listenStr)
	http.ListenAndServe(listenStr, nil)
}
