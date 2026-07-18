// herdr feature palette — a modal, mouse-aware picker built on OpenTUI.
//
// Generic on purpose: all augie-specific logic (feature-dir resolution, config
// sourcing, the open/TablePlus/IntelliJ/notes dispatch) lives in the bash
// wrapper `herdr-feature-palette-tui`. This process only renders the menu on
// the tty and writes the chosen item's `tag` to the --out file. Bash reads that
// tag and dispatches. We can't print the result to stdout because OpenTUI needs
// the real tty there.
//
// Modes:
//   normal  single-key hop (each item's `key`), ↑↓/jk to move, ⏎ to open,
//           `/` to enter search, q/esc to cancel. Mouse: hover highlights,
//           click opens.
//   search  type to filter (non-matches dim + unselectable), ⏎ opens the
//           highlighted match, esc returns to normal.

import { createCliRenderer, BoxRenderable, TextRenderable, type KeyEvent } from "@opentui/core"
import { writeFileSync } from "node:fs"

type Item = { tag: string; key: string; icon: string; label: string; detail: string }
type Payload = { title: string; subtitle?: string; items: Array<Item> }

const outFile = process.argv[2] ?? process.env.PALETTE_OUT ?? ""
const payload: Payload = JSON.parse(process.env.PALETTE_JSON ?? "")
const items = payload.items

const C = {
  accent: "#89b4fa",
  border: "#585b70",
  panelBg: "#181825",
  headerBg: "#11111b",
  normalFg: "#cdd6f4",
  dimFg: "#6c7086",
  selFg: "#11111b",
  keyFg: "#f9e2af",
  subFg: "#a6adc8",
}

const renderer = await createCliRenderer({ exitOnCtrlC: false, useMouse: true })

let mode: "normal" | "search" = "normal"
let query = ""
let selected = 0

const isEnter = (k: KeyEvent): boolean =>
  k.name === "return" || k.name === "enter" || k.name === "linefeed" || k.sequence === "\r" || k.sequence === "\n"

const matches = (it: Item): boolean => {
  if (mode === "normal" || query === "") return true
  const q = query.toLowerCase()
  return `${it.label} ${it.detail} ${it.key} ${it.tag}`.toLowerCase().includes(q)
}
const selectableIdx = (): Array<number> => items.map((_, i) => i).filter((i) => matches(items[i]))

// ── layout ──────────────────────────────────────────────────────────────────

const screen = new BoxRenderable(renderer, {
  id: "screen",
  width: "100%",
  height: "100%",
  flexDirection: "column",
  justifyContent: "center",
  alignItems: "center",
})
renderer.root.add(screen)

const panel = new BoxRenderable(renderer, {
  id: "panel",
  flexDirection: "column",
  width: "90%",
  maxWidth: 76,
  border: true,
  borderColor: C.border,
  backgroundColor: C.panelBg,
})
screen.add(panel)

const header = new TextRenderable(renderer, {
  id: "header",
  content: payload.subtitle ? `❯ ${payload.title}   ${payload.subtitle}` : `❯ ${payload.title}`,
  fg: C.accent,
  backgroundColor: C.headerBg,
  paddingLeft: 1,
  paddingRight: 1,
})
panel.add(header)

const list = new BoxRenderable(renderer, { id: "list", flexDirection: "column", paddingTop: 1, paddingBottom: 1 })
panel.add(list)

const rows = items.map((it, i) => {
  const box = new BoxRenderable(renderer, {
    id: `row-${i}`,
    width: "100%",
    paddingLeft: 1,
    paddingRight: 1,
    backgroundColor: C.panelBg,
    onMouseDown: () => activate(items[i]),
    onMouseOver: () => {
      if (matches(it)) {
        selected = i
        render()
      }
    },
  })
  const text = new TextRenderable(renderer, { id: `row-text-${i}`, content: "", fg: C.normalFg })
  box.add(text)
  list.add(box)
  return { box, text, item: it }
})

const footer = new TextRenderable(renderer, {
  id: "footer",
  content: "",
  fg: C.dimFg,
  backgroundColor: C.headerBg,
  paddingLeft: 1,
  paddingRight: 1,
})
panel.add(footer)

// ── render ────────────────────────────────────────────────────────────────

function render(): void {
  const sel = selectableIdx()
  if (!sel.includes(selected)) selected = sel[0] ?? 0

  for (let i = 0; i < rows.length; i++) {
    const { box, text, item } = rows[i]
    const isMatch = matches(item)
    const isSel = i === selected && isMatch
    const pointer = isSel ? "▶" : " "
    const key = `(${item.key})`
    text.content = `${pointer} ${key} ${item.icon}  ${item.label.padEnd(10)}  ${item.detail}`
    box.backgroundColor = isSel ? C.accent : C.panelBg
    text.fg = isSel ? C.selFg : isMatch ? C.normalFg : C.dimFg
  }

  footer.content =
    mode === "normal"
      ? "(f/s/t/i/n) hop · ↑↓ move · ⏎ open · / search · q quit"
      : `/${query}▏   esc back · ⏎ open`
  renderer.requestRender()
}

// ── actions ─────────────────────────────────────────────────────────────────

function finish(tag: string): void {
  if (outFile) writeFileSync(outFile, tag)
  renderer.destroy()
  process.exit(0)
}
const activate = (it: Item): void => finish(it.tag)
const cancel = (): void => finish("")

function move(delta: number): void {
  const sel = selectableIdx()
  if (sel.length === 0) return
  const cur = sel.indexOf(selected)
  const next = cur === -1 ? 0 : (cur + delta + sel.length) % sel.length
  selected = sel[next]
  render()
}

// ── input ───────────────────────────────────────────────────────────────────

renderer.keyInput.on("keypress", (key: KeyEvent) => {
  if (key.ctrl && key.name === "c") return cancel()

  if (mode === "search") {
    if (key.name === "escape") {
      mode = "normal"
      query = ""
      render()
    } else if (isEnter(key)) {
      const sel = selectableIdx()
      if (sel.includes(selected)) activate(items[selected])
    } else if (key.name === "backspace") {
      query = query.slice(0, -1)
      render()
    } else if (key.name === "up") {
      move(-1)
    } else if (key.name === "down") {
      move(1)
    } else if (key.sequence && key.sequence.length === 1 && key.sequence >= " " && !key.ctrl) {
      query += key.sequence
      render()
    }
    return
  }

  // normal mode
  if (key.name === "escape" || key.name === "q") return cancel()
  if (key.sequence === "/") {
    mode = "search"
    query = ""
    render()
    return
  }
  if (isEnter(key)) {
    activate(items[selected])
    return
  }
  if (key.name === "up" || key.name === "k") return move(-1)
  if (key.name === "down" || key.name === "j") return move(1)

  const hop = items.find((it) => it.key === key.name || it.key === key.sequence)
  if (hop) activate(hop)
})

render()
