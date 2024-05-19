const { WsGateway } = require("./wsgateway");

(async function () {
  const gateway = new WsGateway(
    "http://localhost:5000",
    "redis://localhost:6379/0"
  );

  gateway.listen();

  gateway.on("open", (ws) => {
    console.log("Connection opened", ws.id, ws.tags);
  });
  gateway.on("close", (ws) => {
    console.log("Connection closed", ws.id);
  });
  gateway.on("message", (ws, message) => {
    const parts = message.split(" ");
    if (parts[0] === "echo") {
      ws.send(parts.slice(1).join(" "));
      return;
    }
    if (parts[0] === "settag") {
      ws.setTags({ [parts[1]]: parts[2] });
      ws.send("settag ok");
      return;
    }
    if (parts[0] === "gettag") {
      ws.send(ws.tags.get(parts[1]));
      return;
    }
    if (parts[0] === "send") {
      gateway.sendToTags({ foo: parts[1] }, parts.slice(2).join(" "));
      return;
    }

    console.log("Received message", ws.id, message);
  });
})();
