import { spawn, spawnSync } from "node:child_process";
import { homedir } from "node:os";
import { join } from "node:path";
import type { ExtensionAPI } from "@earendil-works/pi-coding-agent";

const inHerdr = () => process.env.HERDR_ENV === "1" && !!process.env.HERDR_PANE_ID;
const bin = (name: string) => join(homedir(), "bin", name);

function spawnDetached(name: string, args: Array<string>) {
  const child = spawn(bin(name), args, { detached: true, stdio: "ignore" });
  child.on("error", () => {});
  child.unref();
}

export default function (pi: ExtensionAPI) {
  pi.on("session_start", async () => {
    if (!inHerdr()) return;
    spawnDetached("herdr-claude-sidebar", ["--pi"]);
    spawnDetached("herdr-claude-layout", ["start"]);
  });

  pi.on("session_shutdown", async (event) => {
    if (!inHerdr() || event.reason !== "quit") return;
    spawnSync(bin("herdr-claude-layout"), ["end"], { input: "{}", stdio: ["pipe", "ignore", "ignore"], timeout: 5000 });
  });
}
