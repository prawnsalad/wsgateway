const EventEmitter = require("events");
const querystring = require("querystring");
const Redis = require("ioredis");

// redis://user:password@redis-service.com:6379/0
module.exports.WsGateway = class WsGateway extends EventEmitter {
  constructor(gatewayPath, redisUrl) {
    super();
    this.gatewayPath = gatewayPath;
    this.redis = new Redis(redisUrl);
    this.websockets = new Map();
  }

  async listen() {
    const queueName = "connectionevents";
    const consumerGroup = "ws_consumers";
    const consumerName = "main_consumer";

    try {
      await this.redis.xgroup(
        "CREATE",
        queueName,
        consumerGroup,
        "0",
        "MKSTREAM"
      );
    } catch (err) {
      // Ignore if the group already exists
    }

    while (true) {
      const events = await this.redis.xreadgroup(
        "GROUP",
        consumerGroup,
        consumerName,
        "COUNT",
        "1",
        "BLOCK",
        0,
        "STREAMS",
        queueName,
        ">"
      );

      for (const [streamName, messages] of events) {
        for (const message of messages) {
          const [id, rawData] = message;

          if (!Array.isArray(rawData)) {
            // We're only expecting array redis hashes. Anything else we don't know what to do with
            continue;
          }
          const data = redisHashToObject(rawData);
          this.handleStreamMessage(data);
          this.redis.xack(queueName, consumerGroup, id);
        }
      }
    }
  }

  handleStreamMessage(message) {
    if (message.action === "open") {
      const ws = this.getOrAddConnection(message.connection, message.tags);
      this.emit("open", ws);
    } else if (message.action === "close") {
      const ws = this.getOrAddConnection(message.connection, message.tags);
      this.websockets.delete(message.connection);
      this.emit("close", ws);
      ws.emit("close");
    } else if (message.action === "message") {
      const ws = this.getOrAddConnection(message.connection, message.tags);
      this.emit("message", ws, message.message);
      ws.emit("message", message.message);
    }
  }

  getOrAddConnection(id, rawTagString) {
    let ws = this.websockets.get(id);
    if (!ws) {
      ws = new WebSocket(this, id);

      const tags = querystring.parse(rawTagString);
      if (tags) {
        ws.tags = new Map(Object.entries(tags));
      }

      this.websockets.set(id, ws);
    }

    return ws;
  }

  async sendToIds(websocketIds, message) {
    const idList = websocketIds.join(",");
    await fetch(`${this.gatewayPath}/send?id=${idList}`, {
      method: "POST",
      headers: {
        "Content-Type": "text/plain",
      },
      body: message,
    });
  }

  async sendToTags(tags, message) {
    await fetch(`${this.gatewayPath}/send?${querystring.stringify(tags)}`, {
      method: "POST",
      headers: {
        "Content-Type": "text/plain",
      },
      body: message,
    });
  }
};

class WebSocket extends EventEmitter {
  constructor(wsGateway, id) {
    super();
    this.wsGateway = wsGateway;
    this.id = id || "";
    this.tags = new Map();
  }

  send(message) {
    return this.wsGateway.sendToIds([this.id], message);
  }

  close() {}
}

function redisHashToObject(hash) {
  const obj = {};
  for (let i = 0; i < hash.length; i += 2) {
    obj[hash[i]] = hash[i + 1];
  }
  return obj;
}
