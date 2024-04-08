// https://k6.io/docs/using-k6/protocols/websockets/
// Test wsgateway with real websocket connections using the k6 load testing tool
// k6 run --vus 10 --duration 30s k6-runner.js

// import exec from "k6/execution";
import ws from "k6/ws";
import { check } from "k6";

export const options = {
  vus: 1,
  iterations: 1,
  duration: "60s",
};

export default function () {
  const url = "ws://127.0.0.1:5000/connect";
  const params = { tags: { my_tag: "hello" } };

  const res = ws.connect(url, params, function (socket) {
    socket.on("open", function open() {
      // console.log('open', exec.vu.idInTest);

      socket.setInterval(function timeout() {
        socket.ping();
        socket.send("hello");
      }, rand(1, 3) * 1000);

      socket.on("message", function message(data) {
        check(data, {
          // 'routine message' is sent by the included devhelpers.go worker
          "data is correct": (r) =>
            (r && r === "hello") || r === "routine message",
        });
      });
      socket.setInterval(function timeout() {
        socket.close();
      }, rand(5, 40) * 1000);
    });
  });

  check(res, { "status is 101": (r) => r && r.status === 101 });
}

function rand(min, max) {
  return Math.floor(Math.random() * (max - min + 1) + min);
}
