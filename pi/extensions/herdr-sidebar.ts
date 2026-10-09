import { spawn } from "node:child_process";
import { homedir } from "node:os";
import { join } from "node:path";
import type { ExtensionAPI } from "@earendil-works/pi-coding-agent";

export default function (pi: ExtensionAPI) {
  pi.on("session_start", async () => {
    if (process.env.HERDR_ENV !== "1" || !process.env.HERDR_PANE_ID) return;
    const child = spawn(join(homedir(), "bin", "herdr-claude-sidebar"), ["--pi"], {
      detached: true,
      stdio: "ignore",
    });
    child.on("error", () => {});
    child.unref();
  });
}
