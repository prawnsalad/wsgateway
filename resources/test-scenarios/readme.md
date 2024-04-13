Various websocket scenarios to be run against wsgateway.

Each open a websocket connection and vary in use cases from message floods, delays, instant disconnections, etc. Each scenario is run and executed concurrently with other scenarios to simulate a non-expected runtime on the websocket server.

~~~~shell
$ node run.js 
Finished running flood-10-messages.js in 0.053s
Finished running instant-disconnect.js in 0.053s
Finished running message-disconnect.js in 0.052s
Finished running instant-disconnect.js in 0.046s
[...]
~~~