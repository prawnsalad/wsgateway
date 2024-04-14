const { ws, onMessage, onOpen, sleep } = require("./bootstrap")();

onOpen(async () => {
  ws.send("single message");
  ws.close();
});
