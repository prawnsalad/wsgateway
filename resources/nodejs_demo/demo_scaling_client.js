const WebSocket = require("ws");

const w = new WebSocket("ws://localhost:5000/connect");
// w.addEventListener("message", (m) => console.log(m.data));
w.addEventListener("close", (m) => {
  console.log("Closed", m);
});

let cnt = 0;
w.addEventListener("open", async () => {
  tick();

  function tick() {
    cnt++;
    w.send(cnt);

    w.addEventListener("message", listener);

    function listener(e) {
      w.removeEventListener("message", listener);

      const expectedMsg = "Echo: " + cnt;
      if (e.data !== expectedMsg) {
        console.log(
          `Unexpected message. Expected "${expectedMsg} but got "${e.data}"`
        );
      }

      setTimeout(tick, 0);
    }
  }
});
