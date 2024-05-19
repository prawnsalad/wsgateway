const { WsGateway } = require("./wsgateway");

const gateway1 = new WsGateway(
  "http://localhost:5000",
  "redis://localhost:6379/0"
);
const gateway2 = new WsGateway(
  "http://localhost:5000",
  "redis://localhost:6379/0"
);
const gateway3 = new WsGateway(
  "http://localhost:5000",
  "redis://localhost:6379/0"
);
const gateway4 = new WsGateway(
  "http://localhost:5000",
  "redis://localhost:6379/0"
);

addlisteners(gateway1, "gw1");
addlisteners(gateway2, "gw2");
addlisteners(gateway3, "gw3");
addlisteners(gateway4, "gw4");

let lastMessage = "0";
let missedMessage = 0;
let recievedMessages = 0;

// Report if we've been good for 5 seconds
setInterval(() => {
  if (!missedMessage) {
    console.log(
      new Date().toISOString(),
      `Recieved: ${recievedMessages}, Missed: ${missedMessage}`
    );
  }

  recievedMessages = 0;
  missedMessage = 0;
}, 5000);

function addlisteners(gateway, label) {
  gateway.on("open", (ws) => {
    log("Connection opened", ws.id, ws.tags);
  });
  gateway.on("close", (ws) => {
    log("Connection closed", ws.id);
  });
  gateway.on("message", (ws, message) => {
    // log("Received message", ws.id, message);
    recievedMessages++;
    if (parseInt(lastMessage) + 1 !== parseInt(message)) {
      missedMessage++;
      log(`Missed message. Last: "${lastMessage}" New: "${message}"`);
    }
    lastMessage = message;
    ws.send(`Echo: ` + message);
  });

  function log(...args) {
    console.log(`[${label}]`, ...args);
  }
}

gateway1.listen();
gateway2.listen();
gateway3.listen();
gateway4.listen();
