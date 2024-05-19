const { strict: assert } = require("node:assert");
const WebSocket = require("ws");

const w = new WebSocket("ws://localhost:5000/connect");
// w.addEventListener("message", (m) => console.log(m.data));
w.addEventListener("close", (m) => {
  console.log("Websocket closed");
});

w.addEventListener("open", async () => {
  w.send("echo this pls");
  assert((await getNextMessage()) === "this pls", "echo message failed");

  w.send("settag foo newval");
  assert((await getNextMessage()) === "settag ok", "settag failed");

  w.send("gettag foo");
  assert(
    (await getNextMessage()) === "newval",
    "unexpected tag value returned"
  );

  // Change our tag value and then broadcast a message to all connections with the new tag value.
  // We should get it back.
  w.send("settag foo 1234");
  assert(
    (await getNextMessage()) === "settag ok",
    "settag before sending message failed"
  );
  w.send("send 1234 a message!");
  assert(
    (await getNextMessage()) === "a message!",
    "sending message to tag foo=newval failed"
  );

  console.log("All tests passed");
  w.close();
});

function getNextMessage() {
  return new Promise((resolve) => {
    w.addEventListener("message", listener);
    const tmr = setTimeout(() => {
      w.removeEventListener("message", listener);
      resolve(null);
    }, 1000);

    function listener(e) {
      w.removeEventListener("message", listener);
      clearTimeout(tmr);
      resolve(e.data);
    }
  });
}
