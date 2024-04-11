const { WsGateway } = require("./wsgateway");

const serverId = Math.random().toString(36).substring(2);

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
    console.log("Received message", ws.id, message);
    ws.send(`Echo from ${serverId}: ` + message);
  });
})();
