import { defineHexConfig } from "@hex/commands"
import { execFile } from "node:child_process"
import { promisify } from "node:util"

const run = promisify(execFile)

const submitToClaude = async (text: string) => {
  const { stdout } = await run("/Users/christoomey/bin/hex-submit-claude", [], { input: text })
  if (stdout.trim() === "sent") {
    return ""
  } else {
    return text
  }
}

export default defineHexConfig({
  transformations: {
    "submit-to-claude": {
      name: "Submit to Claude",
      description: "Send to the focused Claude pane in herdr, otherwise paste",
      timeout: 3000,
      transform: submitToClaude,
    },
  },
})
