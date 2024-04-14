const { ws, onMessage, onOpen, rand } = require("./bootstrap")();

onOpen(() => {
  ws.close();
});
