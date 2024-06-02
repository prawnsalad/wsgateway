package main

import (
	"flag"
	"log"
	"net/http"
	"runtime"
	"syscall"
	"time"

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

	noFiles, err := getMaxUlimit()
	if err != nil {
		log.Println("Error reading max ulimit: " + err.Error())
	}

	log.Printf("Starting wsgateway. CONFIG=%s NOFILES=%d GOMAXPROCS=%d NUMCPU=%d", configPath, noFiles, runtime.GOMAXPROCS(0), runtime.NumCPU())

	loadedConfig, err := loadConfigFromFile(configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	config = loadedConfig

	library, stream := initComponents()
	runWorkers(library)
	startHttpServer(library, stream)
}

func initComponents() (*connectionlookup.ConnectionLookup, streams.Stream) {
	library, err := connectionlookup.NewConnectionLookup(config.ConnectionRedisSync.Addr)
	if err != nil {
		log.Fatal("Error starting: ", err.Error())
	}

	var stream streams.Stream
	if config.StreamAmqp.Addr != "" {
		stream, err = streams.NewStreamAmqp(
			config.StreamAmqp.Addr,
			config.StreamAmqp.Exchange,
			config.StreamAmqp.Queue,
			config.StreamAmqp.RoutingKey,
		)
		if err != nil {
			log.Fatal("Error starting AMQP stream: ", err.Error())
		}
	} else if config.StreamRedis.Addr != "" {
		stream, err = streams.NewStreamRedis(config.StreamRedis.Addr, config.StreamRedis.StreamName)
		if err != nil {
			log.Fatal("Error starting redis stream: ", err.Error())
		}
	}

	return library, stream
}

func runWorkers(library *connectionlookup.ConnectionLookup) {
	for _, worker := range workersOnBoot {
		worker(library)
	}
}

func startHttpServer(library *connectionlookup.ConnectionLookup, stream streams.Stream) {
	applyWsHandlers(library, stream)
	applyHttpHandlers(library, stream)

	initMetrics(library)

	listenStr := config.ListenAddr
	log.Printf("Internal whitelist %v", config.InternalEndpointWhitelist)
	log.Printf("Listening on %s", listenStr)
	srv := http.Server{
		Addr:        listenStr,
		ReadTimeout: time.Second * 10,
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatalln("Error starting server: ", err.Error())
	}
}

func getMaxUlimit() (uint64, error) {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return 0, err
	}
	return rLimit.Max, nil
}
