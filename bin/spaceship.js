#!/usr/bin/env node

const fs = require("node:fs");
const path = require("node:path");
const { spawnSync } = require("node:child_process");

const executableName = process.platform === "win32" ? "spaceship.exe" : "spaceship-bin";
const executablePath = path.join(__dirname, executableName);

if (!fs.existsSync(executablePath)) {
  console.error("spaceship binary is not installed. Re-run: npm rebuild -g spaceship-domains-cli");
  process.exit(1);
}

const child = spawnSync(executablePath, process.argv.slice(2), {
  stdio: "inherit"
});

if (child.error) {
  console.error(child.error.message);
  process.exit(1);
}

if (child.signal) {
  process.kill(process.pid, child.signal);
}

process.exit(child.status ?? 1);
