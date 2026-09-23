#!/usr/bin/env node
/**
 * uninstall.js — Removes the hush binary and deployed skills.
 * The vault (~/.hush/vault) is kept so reinstalling loses nothing.
 */

const fs = require("fs");
const path = require("path");
const os = require("os");

const BIN_PATH = path.join(os.homedir(), ".local", "bin", "hush");
const SKILL_DIRS = [
  [".claude", "skills", "hush"],
  [".codex", "skills", "hush"],
  [".cursor", "skills", "hush"],
  [".agents", "skills", "hush"],
];

function main() {
  console.log("\n🗑  hush uninstaller\n");

  if (fs.existsSync(BIN_PATH)) {
    fs.unlinkSync(BIN_PATH);
    console.log("  ✓ Removed binary:", BIN_PATH);
  } else {
    console.log("  - Binary not found");
  }

  for (const segs of SKILL_DIRS) {
    const dir = path.join(os.homedir(), ...segs);
    if (fs.existsSync(dir)) {
      fs.rmSync(dir, { recursive: true, force: true });
      console.log("  ✓ Removed skill:", dir);
    }
  }

  console.log("\n  ✅ Done. Vault kept at ~/.hush/vault (delete it manually to wipe secrets).\n");
}

main();
