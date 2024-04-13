const { ws, onMessage, onOpen, rand } = require("./bootstrap")();

onOpen(() => {
  setTimeout(() => {
    ws.close();
  }, rand(0, 3000));
});
