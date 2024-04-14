const { ws, onMessage, onOpen, sleep } = require("./bootstrap")();

onOpen(async () => {
  let running = true;

  setTimeout(() => {
    running = false;
  }, 10_000);

  while (running) {
    await sleep(0);
    ws.send("single message");
  }

  ws.close();
});
