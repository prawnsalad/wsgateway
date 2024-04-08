## Performance test

This is a current snapshot of performance running on a local development Macbook. This is not scientific and should not be considered an accurate benchmark, but more an indication of the ballpark wsgateway currently plays in.

#### Environment
Hardware: Macbook pro m1. 10 CPU cores

`worker-devhelpers.go-ignore` renamed to `worker-devhelpers.go`. This set tags on various connections, looks connections up via tags, sends messages to various connections batching them up using the broadcast feature, all within a constant loop.

wsgateway is run via `make run`.

The k6 load testing script `k6-runner.js` is run on the same machine, connecting via localhost. Running with 100 VUS (virtual users) for 30s each. This connects, every 1-3 seconds sends a websocket "ping" message and a "hello" text message. It reads messages sent from the server and checks it is as expected.

#### Highlights
~~~shell
Client side load testing overview:
ws_connecting.........: avg=3.76ms min=245.29µs med=5.12ms max=8.12ms p(90)=6.94ms p(95)=7.8ms
ws_msgs_received......: 9852818 164212.544057/s
ws_msgs_sent..........: 2555    42.583051/s

Server side wsgateway snippet:
2024/04/09 00:14:31 [0] Sending 50 messages took 100.625µs
2024/04/09 00:14:31 [0] Sending 50 messages took 96.709µs
2024/04/09 00:14:31 Connections: 50 Goroutines: 54 memory: 35947
~~~

Here, wsgateway took roughly 100microseconds to send a message to 50 connections. It was using 35947kb of allocated memory.

### Full client and server output
~~~shell
$ k6 run --vus 100 --duration 30s k6-runner.js

          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     script: k6-runner.js
     output: -

  scenarios: (100.00%) 1 scenario, 100 max VUs, 1m0s max duration (incl. graceful stop):
           * default: 100 looping VUs for 30s (gracefulStop: 30s)


     ✓ data is correct
     ✓ status is 101

     checks................: 100.00% ✓ 9853011       ✗ 0
     data_received.........: 168 MB  2.8 MB/s
     data_sent.............: 88 kB   1.5 kB/s
     iteration_duration....: avg=20.96s min=5s       med=20s    max=40s    p(90)=35s    p(95)=38s
     iterations............: 193     3.216645/s
     vus...................: 5       min=5           max=100
     vus_max...............: 100     min=100         max=100
     ws_connecting.........: avg=3.76ms min=245.29µs med=5.12ms max=8.12ms p(90)=6.94ms p(95)=7.8ms
     ws_msgs_received......: 9852818 164212.544057/s
     ws_msgs_sent..........: 2555    42.583051/s
     ws_session_duration...: avg=20.96s min=5s       med=20s    max=40s    p(90)=35s    p(95)=38s
     ws_sessions...........: 198     3.299978/s


running (1m00.0s), 000/100 VUs, 193 complete and 5 interrupted iterations
default ✓ [======================================] 100 VUs  30s
~~~


~~~shell
wsgateway log snippet:

024/04/09 00:14:26 Connections: 50 Goroutines: 54 memory: 64784
2024/04/09 00:14:26 Marked 50 connections as seen took 18.071666ms
2024/04/09 00:14:27 [0] Sending 50 messages took 228.958µs
2024/04/09 00:14:27 [0] Sending 50 messages took 115.959µs
2024/04/09 00:14:27 [0] Sending 50 messages took 198.417µs
2024/04/09 00:14:27 [0] Sending 50 messages took 220.083µs
2024/04/09 00:14:28 [0] Sending 50 messages took 96.542µs
2024/04/09 00:14:28 [0] Sending 50 messages took 93.458µs
2024/04/09 00:14:28 [0] Sending 50 messages took 264.334µs
2024/04/09 00:14:29 [0] Sending 50 messages took 122.667µs
2024/04/09 00:14:29 [0] Sending 50 messages took 158.625µs
2024/04/09 00:14:29 [0] Sending 50 messages took 101.916µs
2024/04/09 00:14:30 [0] Sending 50 messages took 120.333µs
2024/04/09 00:14:30 [0] Sending 50 messages took 153.708µs
2024/04/09 00:14:30 [0] Sending 50 messages took 110.75µs
2024/04/09 00:14:30 [0] Sending 50 messages took 102µs
2024/04/09 00:14:31 [0] Sending 50 messages took 192.75µs
2024/04/09 00:14:31 [0] Sending 50 messages took 100.625µs
2024/04/09 00:14:31 [0] Sending 50 messages took 96.709µs
2024/04/09 00:14:31 Connections: 50 Goroutines: 54 memory: 35947
~~~