#!/usr/bin/env node

const { spawn, execFileSync } = require("child_process");
const https = require("https");
const path = require("path");
const fs = require("fs");

const NAME = "dotenvx-tui";
const REPO = "warui1/dotenvx-tui";
const VERSION = require("../package.json").version;
const binDir = __dirname;
const binPath = path.join(binDir, NAME);
const versionPath = `${binPath}.version`;

const PLATFORM_MAP = { darwin: "darwin", linux: "linux" };
const ARCH_MAP = { x64: "amd64", arm64: "arm64" };

function downloadFile(url) {
  return new Promise((resolve, reject) => {
    https
      .get(url, (res) => {
        if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
          return downloadFile(res.headers.location).then(resolve, reject);
        }
        if (res.statusCode !== 200) {
          return reject(new Error(`HTTP ${res.statusCode}`));
        }
        const total = parseInt(res.headers["content-length"] || "0", 10);
        const chunks = [];
        let downloaded = 0;
        res.on("data", (chunk) => {
          chunks.push(chunk);
          downloaded += chunk.length;
          const mb = (downloaded / 1024 / 1024).toFixed(1);
          if (total > 0) {
            const pct = Math.round((downloaded / total) * 100);
            process.stderr.write(`\r  downloading ${NAME}... ${mb} MB (${pct}%)`);
          } else {
            process.stderr.write(`\r  downloading ${NAME}... ${mb} MB`);
          }
        });
        res.on("end", () => {
          process.stderr.write("\n");
          resolve(Buffer.concat(chunks));
        });
        res.on("error", reject);
      })
      .on("error", reject);
  });
}

async function ensureBinary() {
  if (fs.existsSync(binPath) && fs.existsSync(versionPath)) {
    const installedVersion = fs.readFileSync(versionPath, "utf8").trim();
    if (installedVersion === VERSION) {
      return;
    }
  }

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

  const archiveName = `${NAME}_${VERSION}_${platform}_${arch}.tar.gz`;
  const url = `https://github.com/${REPO}/releases/download/v${VERSION}/${archiveName}`;

  process.stderr.write(`${NAME} v${VERSION} — first run setup\n`);

  let tmpFile = "";
  try {
    const buffer = await downloadFile(url);
    tmpFile = path.join(binDir, `_${NAME}-${process.pid}.tar.gz`);
    fs.writeFileSync(tmpFile, buffer);
    execFileSync("tar", ["-xzf", tmpFile, "-C", binDir, NAME], { stdio: "ignore" });
    fs.chmodSync(binPath, 0o755);
    fs.writeFileSync(versionPath, `${VERSION}\n`);
    process.stderr.write(`  ready!\n\n`);
  } catch (err) {
    console.error(
      `\nFailed to download ${NAME}: ${err.message}\n\n` +
        `Install manually:\n` +
        `  go install github.com/${REPO}@latest\n` +
        `  # or: https://github.com/${REPO}/releases`
    );
    process.exit(1);
  } finally {
    if (tmpFile) {
      fs.rmSync(tmpFile, { force: true });
    }
  }
}

async function main() {
  await ensureBinary();

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
}

main();
