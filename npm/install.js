#!/usr/bin/env node

// Postinstall script: downloads the correct binary from GitHub Releases.
// No external dependencies — uses Node.js built-in https/fs/child_process.

const https = require("https");
const fs = require("fs");
const path = require("path");
const { execSync } = require("child_process");
const { createGunzip } = require("zlib");

const REPO = "SpyrosBou/dotenvx-tui";
const NAME = "dotenvx-tui";
const VERSION = require("./package.json").version;

const PLATFORM_MAP = {
  darwin: "darwin",
  linux: "linux",
};

const ARCH_MAP = {
  x64: "amd64",
  arm64: "arm64",
};

function getTarget() {
  const platform = PLATFORM_MAP[process.platform];
  const arch = ARCH_MAP[process.arch];

  if (!platform || !arch) {
    console.error(
      `Unsupported platform: ${process.platform}-${process.arch}\n` +
        `Supported: darwin-arm64, darwin-x64, linux-arm64, linux-x64\n` +
        `Build from source: go install github.com/${REPO}@latest`
    );
    process.exit(1);
  }

  return { platform, arch };
}

function downloadFile(url) {
  return new Promise((resolve, reject) => {
    https
      .get(url, (res) => {
        if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
          return downloadFile(res.headers.location).then(resolve, reject);
        }
        if (res.statusCode !== 200) {
          return reject(new Error(`Download failed: HTTP ${res.statusCode}`));
        }
        const chunks = [];
        res.on("data", (chunk) => chunks.push(chunk));
        res.on("end", () => resolve(Buffer.concat(chunks)));
        res.on("error", reject);
      })
      .on("error", reject);
  });
}

function extractTarGz(buffer, destDir) {
  // Write to temp file and use system tar (available on macOS + Linux)
  const tmpFile = path.join(destDir, "_archive.tar.gz");
  fs.writeFileSync(tmpFile, buffer);
  execSync(`tar -xzf "${tmpFile}" -C "${destDir}" "${NAME}"`, { stdio: "ignore" });
  fs.unlinkSync(tmpFile);
}

async function main() {
  const { platform, arch } = getTarget();
  const archiveName = `${NAME}_${VERSION}_${platform}_${arch}.tar.gz`;
  const url = `https://github.com/${REPO}/releases/download/v${VERSION}/${archiveName}`;
  const binDir = path.join(__dirname, "bin");

  // Skip if binary already exists
  const binPath = path.join(binDir, NAME);
  if (fs.existsSync(binPath)) {
    return;
  }

  console.log(`Downloading ${NAME} v${VERSION} for ${platform}/${arch}...`);

  try {
    const buffer = await downloadFile(url);
    fs.mkdirSync(binDir, { recursive: true });
    extractTarGz(buffer, binDir);
    fs.chmodSync(binPath, 0o755);
    console.log(`Installed ${NAME} to ${binPath}`);
  } catch (err) {
    console.error(
      `Failed to download ${NAME}: ${err.message}\n\n` +
        `You can install manually:\n` +
        `  go install github.com/${REPO}@latest\n` +
        `  # or download from: https://github.com/${REPO}/releases`
    );
    process.exit(1);
  }
}

main();
