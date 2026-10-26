#!/usr/bin/env node

const fs = require("node:fs");
const path = require("node:path");
const { execFileSync } = require("node:child_process");

const root = path.resolve(__dirname, "..");
const pkg = require(path.join(root, "package.json"));
const binaries = [
  "spaceship_darwin_amd64",
  "spaceship_darwin_arm64",
  "spaceship_linux_amd64",
  "spaceship_linux_arm64",
  "spaceship_windows_amd64.exe"
];

if (pkg.license !== "Apache-2.0") {
  throw new Error("npm license metadata must match LICENSE (Apache-2.0)");
}
for (const file of binaries) {
  const filePath = path.join(root, "bin", file);
  if (!fs.existsSync(filePath) || !fs.statSync(filePath).isFile() || fs.statSync(filePath).size === 0) {
    throw new Error(`missing or empty built binary: ${file}`);
  }
}

fs.mkdirSync(path.join(root, "pkg"), { recursive: true });
const metadata = JSON.parse(execFileSync("npm", ["pack", "--json", "--pack-destination", "pkg"], { cwd: root, encoding: "utf8" }))[0];
const packagedFiles = new Set(metadata.files.map((file) => file.path));
for (const file of ["bin/spaceship.js", "scripts/postinstall.js", "LICENSE", ...binaries.map((name) => `bin/${name}`)]) {
  if (!packagedFiles.has(file)) {
    throw new Error(`archive is missing required file: ${file}`);
  }
}
if (metadata.version !== pkg.version) {
  throw new Error(`archive version mismatch: ${metadata.version} != ${pkg.version}`);
}
console.log(`Validated ${metadata.filename}: all ${binaries.length} platform binaries, Apache-2.0 license, and version ${pkg.version}`);
