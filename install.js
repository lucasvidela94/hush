#!/usr/bin/env node
/**
 * install.js — install the hush binary and deploy the skill into your harnesses.
 *
 * Idempotent. Safe to run more than once. Used as npm postinstall.
 *
 * The npm tarball already ships the platform binaries, so by default nothing
 * is downloaded: assets are unpacked from the local package.
 * GitHub Releases is only a fallback for development checkouts without bin/.
 */

const fs = require("fs");
const os = require("os");
const path = require("path");
const zlib = require("zlib");
const { execFileSync, execSync } = require("child_process");

const PKG = require("./package.json");
const REPO = PKG.repository.url.replace("git+", "").replace(".git", "");
const VERSION = "v" + PKG.version;

const PKG_DIR = __dirname;
const BIN_DIR = path.join(os.homedir(), ".local", "bin");
const BIN_PATH = path.join(BIN_DIR, "hush");

function platform() {
  const arch = os.arch();
  const plat = os.platform();

  if (plat === "linux" && arch === "x64") return "linux-amd64";
  if (plat === "linux" && arch === "arm64") return "linux-arm64";
  if (plat === "darwin" && arch === "x64") return "darwin-amd64";
  if (plat === "darwin" && arch === "arm64") return "darwin-arm64";

  console.error(`Unsupported platform: ${plat}-${arch}`);
  process.exit(1);
}

function download(url, dest) {
  return new Promise((resolve, reject) => {
    const https = require("https");
    const file = fs.createWriteStream(dest);
    https
      .get(url, (res) => {
        if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
          file.close();
          fs.unlinkSync(dest);
          return download(res.headers.location, dest).then(resolve).catch(reject);
        }
        if (res.statusCode !== 200) {
          file.close();
          fs.unlinkSync(dest);
          reject(new Error(`HTTP ${res.statusCode}: ${url}`));
          return;
        }
        res.pipe(file);
        file.on("finish", () => {
          file.close();
          resolve();
        });
      })
      .on("error", (err) => {
        file.close();
        fs.unlinkSync(dest, () => {});
        reject(err);
      });
  });
}

function binaryInstalled() {
  if (!fs.existsSync(BIN_PATH)) return false;
  try {
    const out = execSync(`"${BIN_PATH}" version 2>&1 || true`, { encoding: "utf-8" });
    return out.includes(VERSION) || out.includes(PKG.version);
  } catch {
    return false;
  }
}

async function ensureBinary() {
  if (binaryInstalled()) {
    console.log(`  ✓ Binary already up-to-date (${VERSION})`);
    return;
  }

  const assetName = `hush-${platform()}.gz`;
  const localSrc = path.join(PKG_DIR, "bin", assetName);

  let compressed;
  if (fs.existsSync(localSrc)) {
    console.log(`  ✓ Using bundled binary ${assetName}`);
    compressed = fs.readFileSync(localSrc);
  } else {
    const url = `${REPO}/releases/download/${VERSION}/${assetName}`;
    const dest = path.join(os.tmpdir(), assetName);
    console.log(`  ⬇ Downloading ${assetName}...`);
    try {
      await download(url, dest);
      compressed = fs.readFileSync(dest);
      fs.unlinkSync(dest);
    } catch (err) {
      console.error(`  ❌ Binary download failed: ${err.message}`);
      console.error("     Build from source instead: git clone ... && go build -o hush ./cmd/hush");
      process.exit(1);
    }
  }

  fs.mkdirSync(BIN_DIR, { recursive: true });
  const binary = zlib.gunzipSync(compressed);
  // Atomic rename: a running hush keeps its old inode, next launch picks up the new binary.
  const tmpPath = `${BIN_PATH}.tmp-${process.pid}`;
  fs.writeFileSync(tmpPath, binary, { mode: 0o755 });
  try {
    fs.renameSync(tmpPath, BIN_PATH);
  } catch (err) {
    try {
      fs.unlinkSync(tmpPath);
    } catch {}
    throw err;
  }
  console.log(`  ✓ Installed to ${BIN_PATH}`);
}

function runSetup() {
  try {
    const out = execFileSync(BIN_PATH, ["setup"], { encoding: "utf-8" });
    console.log(out);
  } catch (err) {
    console.error(`  ❌ hush setup failed: ${err.message}`);
    process.exit(1);
  }
}

async function main() {
  console.log("\n📦 hush installer");
  console.log("================\n");

  await ensureBinary();
  runSetup();

  console.log("  ─────────────────────────────────────");
  console.log("  ✅ hush installed!");
  console.log("");
  console.log("  Next step (in YOUR terminal, never in chat):");
  console.log("    hush set MY_API_KEY");
  console.log("  ─────────────────────────────────────\n");
}

main().catch((err) => {
  console.error("Install failed:", err.message);
  process.exit(1);
});
