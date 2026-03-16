#!/usr/bin/env node

const fs = require("node:fs");
const path = require("node:path");
const https = require("node:https");
const { spawnSync } = require("node:child_process");

const pkg = require("../package.json");
const repoOwner = "amxv";
const repoName = "spaceship-domains-cli";

const goosMap = {
  darwin: "darwin",
  linux: "linux",
  win32: "windows"
};

const goarchMap = {
  x64: "amd64",
  arm64: "arm64"
};

const goos = goosMap[process.platform];
const goarch = goarchMap[process.arch];

const rootDir = path.resolve(__dirname, "..");
const binDir = path.join(rootDir, "bin");
const binaryName = process.platform === "win32" ? "spaceship.exe" : "spaceship-bin";
const destination = path.join(binDir, binaryName);

async function main() {
  if (!goos || !goarch) {
    console.warn(`Unsupported platform/arch: ${process.platform}/${process.arch}`);
    fallbackBuildOrExit();
    return;
  }

  fs.mkdirSync(binDir, { recursive: true });

  const version = pkg.version;
  const assetName = `spaceship_${goos}_${goarch}${goos === "windows" ? ".exe" : ""}`;
  const assetURL = `https://github.com/${repoOwner}/${repoName}/releases/download/v${version}/${assetName}`;

  try {
    await downloadToFile(assetURL, destination);
    if (goos !== "windows") {
      fs.chmodSync(destination, 0o755);
    }
    console.log(`Installed spaceship binary from ${assetURL}`);
  } catch (err) {
    console.warn(`Failed to download prebuilt binary: ${err.message}`);
    fallbackBuildOrExit();
  }
}

function fallbackBuildOrExit() {
  const hasGo = spawnSync("go", ["version"], { stdio: "ignore" }).status === 0;
  const hasSource = fs.existsSync(path.join(rootDir, "cmd", "spaceship", "main.go"));

  if (!hasGo || !hasSource) {
    console.error("Unable to install spaceship binary. Missing release asset and local Go build fallback.");
    process.exit(1);
  }

  const result = spawnSync(
    "go",
    ["build", "-o", destination, "./cmd/spaceship"],
    { cwd: rootDir, stdio: "inherit" }
  );

  if (result.status !== 0) {
    process.exit(result.status || 1);
  }

  if (goos !== "windows") {
    fs.chmodSync(destination, 0o755);
  }

  console.log("Installed spaceship binary by building from source.");
}

function downloadToFile(url, destinationPath) {
  return new Promise((resolve, reject) => {
    const request = https.get(url, (response) => {
      if (response.statusCode >= 300 && response.statusCode < 400 && response.headers.location) {
        response.resume();
        downloadToFile(response.headers.location, destinationPath).then(resolve).catch(reject);
        return;
      }

      if (response.statusCode !== 200) {
        response.resume();
        reject(new Error(`HTTP ${response.statusCode}`));
        return;
      }

      const file = fs.createWriteStream(destinationPath);
      response.pipe(file);

      file.on("finish", () => {
        file.close((err) => {
          if (err) {
            reject(err);
            return;
          }
          resolve();
        });
      });

      file.on("error", (err) => {
        fs.rmSync(destinationPath, { force: true });
        reject(err);
      });
    });

    request.on("error", reject);
  });
}

main().catch((err) => {
  console.error(err.message);
  process.exit(1);
});
