const { ws, onMessage, onOpen, rand } = require("./bootstrap")();

onOpen(() => {
  ws.send("one message");
  ws.close();
});
