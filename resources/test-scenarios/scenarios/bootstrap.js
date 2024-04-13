const WebSocket = require("ws");

function rand(min, max) {
  // inclusive
  return Math.floor(Math.random() * (max - min + 1) + min);
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

module.exports = function () {
  const ws = new WebSocket("ws://127.0.0.1:5000/connect");

  ws.on("error", console.error);
  ws.on("close", () => {
    process.exit(0);
  });

  function onOpen(cb) {
    ws.on("open", cb);
  }

  function onMessage(cb) {
    ws.on("message", cb);
  }

  return { rand, sleep, ws, onOpen, onMessage };
};
