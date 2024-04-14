package main

import (
	"net/http"
	"time"

	"com.wsgateway/connectionlookup"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var counterClientSentMsgs = promauto.NewCounter(prometheus.CounterOpts{
	Name: "messages_sent_counter",
	Help: "The number of websocket messages sent",
})

var counterClientRecievedMsgs = promauto.NewCounter(prometheus.CounterOpts{
	Name: "messages_recieved_counter",
	Help: "The number of websocket messages",
})

var counterConnections = promauto.NewCounter(prometheus.CounterOpts{
	Name: "connections_counter",
	Help: "The number of websocket connections made",
})

var counterDisconnections = promauto.NewCounter(prometheus.CounterOpts{
	Name: "disconnections_counter",
	Help: "The number of websocket disconnections made",
})

var gaugeConnections = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "active_connections_gauge",
	Help: "The number of active websocket connections",
})

func initMetrics(library *connectionlookup.ConnectionLookup) {
	if !config.Prometheus.Enabled {
		return
	}

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		for {
			numCons := library.NumConnections()
			gaugeConnections.Set(float64(numCons))

			time.Sleep(1 * time.Second)
		}
	}()
}
