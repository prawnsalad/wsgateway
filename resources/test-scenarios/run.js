const { readdir } = require("fs/promises");
const { fork } = require("child_process");

// Run all the scripts in /scenarios/ indefinitely
const testsPath = "./scenarios";

(async () => {
  const scripts = [];

  const files = await readdir(testsPath);
  for (const file of files) {
    if (file === "bootstrap.js") {
      continue;
    }
    scripts.push(file);
  }

  await runConcurrently(scripts, 5);
})();

async function runConcurrently(scripts, numConcurrent = 10) {
  let running = 0;

  let currentScriptIdx = -1;
  while (true) {
    currentScriptIdx++;
    // keep running through all the scripts
    if (currentScriptIdx >= scripts.length) {
      currentScriptIdx = 0;
    }

    if (running >= numConcurrent) {
      await sleep(500);
      continue;
    }

    running++;
    forkScript(scripts[currentScriptIdx]).finally(() => {
      running--;
    });
  }
}

function forkScript(script) {
  return new Promise((resolve, reject) => {
    const started = Date.now();
    const child = fork(`${testsPath}/${script}`, {
      silent: true,
    });
    child.on("exit", () => {
      const msTaken = Date.now() - started;
      console.log(`Finished running ${script} in ${msTaken / 1000}s`);
      resolve();
    });
  });
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

/*
connect.. wait 0.0 - 3s.. disconnect - wait-disconnect.js
connect.. disconnect - instant-disconnect.js
connect.. send 1 message.. disconnect - one-message.js
connect.. flood 10 messages.. disconnect - flood-10-messages.js
connect.. send 1 message.. wait 1s.. send 1 message.. wait 1s.. disconnect - message-second-close.js
connect.. send 1 message + disconnect  instantly - message-disconnect.js
connect.. flood messages for 20s.. disconnect
*/
