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
const { execFileSync } = require("child_process");

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

function download(url, dest, redirects = 0) {
  return new Promise((resolve, reject) => {
    const { protocol } = new URL(url);
    if (protocol !== "https:") {
      reject(new Error(`refusing non-https URL: ${url}`));
      return;
    }
    if (redirects > 5) {
      reject(new Error(`too many redirects: ${url}`));
      return;
    }
    const https = require("https");
    const file = fs.createWriteStream(dest);
    https
      .get(url, (res) => {
        if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
          file.close();
          fs.rmSync(dest, { force: true });
          return download(res.headers.location, dest, redirects + 1).then(resolve).catch(reject);
        }
        if (res.statusCode !== 200) {
          file.close();
          fs.rmSync(dest, { force: true });
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
        fs.rmSync(dest, { force: true });
        reject(err);
      });
  });
}

function binaryInstalled() {
  if (!fs.existsSync(BIN_PATH)) return false;
  try {
    const out = execFileSync(BIN_PATH, ["version"], { encoding: "utf-8" });
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
    compressed = await downloadRelease(assetName);
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

async function downloadRelease(assetName) {
  const { randomBytes } = require("crypto");
  const url = `${REPO}/releases/download/${VERSION}/${assetName}`;
  const workdir = fs.mkdtempSync(path.join(os.tmpdir(), "hush-"));
  const dest = path.join(workdir, assetName);
  console.log(`  ⬇ Downloading ${assetName}...`);
  try {
    await download(url, dest);
    const compressed = fs.readFileSync(dest);
    await verifyChecksum(assetName, compressed);
    return compressed;
  } catch (err) {
    console.error(`  ❌ Binary download failed: ${err.message}`);
    console.error("     Build from source instead: git clone ... && go build -o hush ./cmd/hush");
    process.exit(1);
  } finally {
    fs.rmSync(workdir, { recursive: true, force: true });
  }
}

async function verifyChecksum(assetName, compressed) {
  const { createHash } = require("crypto");
  const sumsUrl = `${REPO}/releases/download/${VERSION}/checksums.txt`;
  const workdir = fs.mkdtempSync(path.join(os.tmpdir(), "hush-sums-"));
  try {
    const sumsDest = path.join(workdir, "checksums.txt");
    await download(sumsUrl, sumsDest);
    const sums = fs.readFileSync(sumsDest, "utf-8");
    const line = sums.split("\n").find((l) => l.trim().endsWith(`  ${assetName}`));
    if (!line) throw new Error(`checksum for ${assetName} not found`);
    const expected = line.split(/\s+/)[0];
    const actual = createHash("sha256").update(compressed).digest("hex");
    if (actual.toLowerCase() !== expected.toLowerCase()) {
      throw new Error(`checksum mismatch for ${assetName}`);
    }
    console.log("  ✓ Checksum verified");
  } finally {
    fs.rmSync(workdir, { recursive: true, force: true });
  }
}

function runSetup() {  try {
    const out = execFileSync(BIN_PATH, ["setup", "--yes"], { encoding: "utf-8" });
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
