package main

import (
	"log"
	"runtime"
	"time"

	"com.wsgateway/connectionlookup"
)

func runDevHelpers(library *connectionlookup.ConnectionLookup) {
	go func() {
		for {
			mem := &runtime.MemStats{}
			runtime.ReadMemStats(mem)
			log.Printf("Connections: %d Goroutines: %d memory: %v", library.NumConnections(), runtime.NumGoroutine(), mem.HeapAlloc / 1000)
			time.Sleep(time.Second*5)
		}
	}()

	go func() {		
		for {
			time.Sleep(time.Second*5)

			cons := library.GetConnections()
			newTags := map[string]string{
				"seen": "1",
			}

			start := time.Now()
			for _, con := range cons {
				library.SetKeys(con, newTags)
			}

			log.Printf("Marked %d connections as seen took %v\n", len(cons), time.Since(start))
		}
	}()

	// Multiple workers to split the load up between threads
	for i:=0; i<1; i++ {
		go workerSend1000Sec(i, library)
		time.Sleep(time.Millisecond * 130)
	}
}

func workerSend1000Sec(workerId int, library *connectionlookup.ConnectionLookup) {
	cnt := 0

	for {
		//time.Sleep(time.Second / 10000)
		time.Sleep(time.Nanosecond * (1000000 / 10))

		cons := library.GetConnectionsWithKeys(map[string]string{"foo": "bar"})
		start := time.Now()
		for _, conn := range cons {
			conn.Socket.WriteString("routine message")
		}

		cnt++
		if cnt == 1000 {
			cnt = 0
			log.Printf("[%d] Sending %d messages took %v\n", workerId, len(cons), time.Since(start))
		}
	}
}