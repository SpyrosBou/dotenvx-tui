#!/usr/bin/env node

const { spawn } = require("child_process");
const path = require("path");
const fs = require("fs");

const NAME = "dotenvx-tui";
const binPath = path.join(__dirname, NAME);

if (!fs.existsSync(binPath)) {
  console.error(
    `${NAME} binary not found. Try reinstalling:\n` +
      `  npm install -g dotenvx-tui`
  );
  process.exit(1);
}

const child = spawn(binPath, process.argv.slice(2), {
  stdio: "inherit",
});

child.on("error", (err) => {
  console.error(`Failed to start ${NAME}: ${err.message}`);
  process.exit(1);
});

child.on("close", (code) => {
  process.exit(code ?? 1);
});
