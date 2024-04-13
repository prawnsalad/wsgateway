const { ws, onMessage, onOpen, sleep } = require("./bootstrap")();

onOpen(async () => {
  ws.send("message 1");
  await sleep(1000);
  ws.send("message 2");
  await sleep(1000);
  ws.close();
});
