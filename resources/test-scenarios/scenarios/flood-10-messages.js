const { ws, onMessage, onOpen, rand } = require("./bootstrap")();

onOpen(() => {
  for (let i = 0; i < 10; i++) {
    ws.send("message number " + i);
  }
  ws.close();
});
