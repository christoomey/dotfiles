// Claude Code port of https://github.com/johnlindquist/pi-emphasis (MIT, John Lindquist).
import { atom, read, update } from 'claude-code'
import type { Register } from 'claude-code'

const isEnabled = atom({ plugin: 'emphasis', key: 'isEnabled' } as const, true)

const LEFT_CAP = ''
const RIGHT_CAP = ''

type Tone = { bg: string; fg: string }

const TONES = {
  red: { bg: '#f38ba8', fg: '#1e1e2e' },
  yellow: { bg: '#f9e2af', fg: '#1e1e2e' },
  green: { bg: '#a6e3a1', fg: '#1e1e2e' },
  danger: { bg: '#d20f39', fg: '#ffffff' },
  cyan: { bg: '#94e2d5', fg: '#1e1e2e' },
  owner: { bg: '#1e66f5', fg: '#ffffff' },
  context: { bg: '#89b4fa', fg: '#1e1e2e' },
} satisfies Record<string, Tone>

type ToneName = keyof typeof TONES

const PILLS: Array<{ tone: ToneName; words: Array<[string, string]> }> = [
  {
    tone: 'red',
    words: [
      ['Failed', ''],
      ['Blocked', ''],
      ['Broken', ''],
      ['Unsafe', ''],
      ['Rejected', ''],
      ['Down', ''],
    ],
  },
  {
    tone: 'yellow',
    words: [
      ['Needs you', ''],
      ['Decision', ''],
      ['Unverified', ''],
      ['Partial', ''],
      ['Skipped', ''],
      ['In progress', ''],
      ['Pending', ''],
      ['At risk', ''],
      ['Review', ''],
      ['Todo', ''],
      ['Stale', ''],
      ['Draft', ''],
    ],
  },
  {
    tone: 'green',
    words: [
      ['Done', ''],
      ['Passed', ''],
      ['Observed', ''],
      ['Fixed', ''],
      ['Live', ''],
      ['Recommended', ''],
      ['Keep', ''],
      ['Ready', ''],
      ['Merged', ''],
      ['Safe', ''],
      ['Verified', ''],
      ['Shipped', ''],
      ['Approved', ''],
      ['Healthy', ''],
    ],
  },
  {
    tone: 'danger',
    words: [
      ['Costs money', ''],
      ['Destructive', ''],
      ['Irreversible', ''],
    ],
  },
  {
    tone: 'cyan',
    words: [
      ['Brainstorm', ''],
      ['Investigate', ''],
      ['Preview', ''],
      ['Implement', ''],
      ['Qualify', ''],
    ],
  },
  {
    tone: 'owner',
    words: [
      ['Owned', ''],
      ['Shared', ''],
      ['Foreign', ''],
    ],
  },
  {
    tone: 'context',
    words: [
      ['Inferred', ''],
      ['New', ''],
      ['Changed', ''],
      ['Optional', ''],
    ],
  },
]

const SYNONYMS: Record<string, string> = {
  completed: 'Done',
  complete: 'Done',
  ok: 'Done',
  finished: 'Done',
  fail: 'Failed',
  failing: 'Failed',
  error: 'Failed',
  pass: 'Passed',
  waiting: 'Pending',
  'needs review': 'Review',
}

const DEFAULT_ICON = ''

type PillWord = { word: string; icon: string; tone: ToneName }

const PILL_WORDS: Array<PillWord> = PILLS.flatMap(group =>
  group.words.map(([word, icon]) => ({ word, icon, tone: group.tone })),
)

const LABEL_ICONS: Record<string, string> = {
  finding: '',
  result: '',
  status: '',
  evidence: '',
  cause: '',
  effect: '',
  fix: '',
  risk: '',
  limits: '',
  cost: '',
  decision: '',
  options: '',
  plan: '',
  next: '',
  why: '',
  how: '',
  rule: '',
  example: '',
  checked: '',
  verdict: '',
  rollback: '',
  undo: '',
  review: '',
  terminal: '',
  folder: '',
  file: '',
  branch: '',
  ideas: '',
  'tl;dr': '',
}

const ACTION_ICONS: Record<string, string> = {
  check: '',
  do: '',
  next: '',
  unblock: '',
  left: '',
  run: '',
  guard: '',
  undo: '',
  see: '',
}

const PILL_PATTERN = /\[\[([^[\]\n]{1,32})\]\]/

function resolvePill(raw: string): { icon: string; tone: ToneName; text: string } {
  const label = raw.trim()
  const lower = label.toLowerCase()

  if (/^[-]/.test(label)) {
    return { icon: '', tone: 'context', text: label }
  }

  for (const [synonym, word] of Object.entries(SYNONYMS)) {
    if (lower === synonym || lower.startsWith(`${synonym} `) || lower.startsWith(`${synonym}:`)) {
      const known = PILL_WORDS.find(p => p.word === word)

      if (known) {
        return { icon: known.icon, tone: known.tone, text: known.word + label.slice(synonym.length) }
      }
    }
  }

  const byLength = [...PILL_WORDS].sort((a, b) => b.word.length - a.word.length)

  for (const pill of byLength) {
    const w = pill.word.toLowerCase()

    if (lower === w || lower.startsWith(`${w} `) || lower.startsWith(`${w}:`)) {
      return { icon: pill.icon, tone: pill.tone, text: label }
    }
  }

  return { icon: DEFAULT_ICON, tone: 'context', text: label }
}

const GLYPH_START = /^[-]/

function addLabelIcons(line: string): string {
  let out = line.replace(
    /^(#{1,6}\s+)([A-Za-z;]+)(:?)(?=\s|$)/,
    (match, hashes: string, word: string, colon: string) => {
      const icon = LABEL_ICONS[word.toLowerCase()]

      return icon ? `${hashes}${icon} ${word}${colon}` : match
    },
  )

  out = out.replace(
    /^(\s*(?:[-*+]\s+|\d+[.)]\s+)?)\*\*([A-Za-z;]+)(:?)\*\*/,
    (match, lead: string, word: string, colon: string) => {
      const icon = LABEL_ICONS[word.toLowerCase()] ?? (colon ? ACTION_ICONS[word.toLowerCase()] : undefined)

      return icon ? `${lead}**${icon} ${word}${colon}**` : match
    },
  )

  out = out.replace(/\*\*([A-Za-z]+):\*\*/g, (match, word: string) => {
    const icon = ACTION_ICONS[word.toLowerCase()]

    return icon ? `**${icon} ${word}:**` : match
  })

  return out
}

type Segment =
  | { kind: 'markdown'; lines: Array<string> }
  | { kind: 'pills'; line: string }
  | { kind: 'blank' }

function isTableRow(line: string): boolean {
  return /^\s*\|/.test(line)
}

function segment(text: string): Array<Segment> {
  const segments: Array<Segment> = []
  let fence: string | null = null
  let buffer: Array<string> = []

  const flush = () => {
    while (buffer.length > 0 && buffer[0]!.trim() === '') {
      buffer.shift()
    }

    while (buffer.length > 0 && buffer[buffer.length - 1]!.trim() === '') {
      buffer.pop()
    }

    if (buffer.length > 0) {
      segments.push({ kind: 'markdown', lines: buffer })
    }

    buffer = []
  }

  for (const line of text.split('\n')) {
    const fenceMatch = /^\s*(```|~~~)/.exec(line)

    if (fence !== null) {
      buffer.push(line)

      if (fenceMatch && fenceMatch[1] === fence) {
        fence = null
      }
    } else if (fenceMatch) {
      fence = fenceMatch[1]!
      buffer.push(line)
    } else if (PILL_PATTERN.test(stripCode(line)) && !isTableRow(line) && !/^ {4,}/.test(line)) {
      const hadContent = buffer.some(l => l.trim() !== '')
      const endedBlank = buffer.length > 0 && buffer[buffer.length - 1]!.trim() === ''
      flush()

      if (hadContent && endedBlank) {
        segments.push({ kind: 'blank' })
      }

      segments.push({ kind: 'pills', line: addLabelIcons(line) })
    } else {
      const previous = segments[segments.length - 1]

      if (buffer.length === 0 && line.trim() === '' && previous?.kind === 'pills') {
        segments.push({ kind: 'blank' })
      } else if (buffer.length > 0 || line.trim() !== '') {
        buffer.push(addLabelIcons(line))
      }
    }
  }

  flush()

  return segments.filter(
    (s, i, all) => !(s.kind === 'blank' && (i === 0 || i === all.length - 1 || all[i - 1]!.kind === 'blank')),
  )
}

function stripCode(line: string): string {
  return line.replace(/`[^`]*`/g, '')
}

type Style = { bold?: boolean; italic?: boolean; strikethrough?: boolean }

type Inline =
  | { kind: 'text'; text: string; style: Style }
  | { kind: 'code'; text: string }
  | { kind: 'pill'; raw: string }
  | { kind: 'link'; href: string; label: string; style: Style }
  | { kind: 'mark'; text: string; color: string; dim?: boolean }

const INLINE =
  /(`[^`]+`)|\[\[([^[\]\n]{1,32})\]\]|\[([^\]]+)\]\(([^)\s]+)\)|<(https?:\/\/[^>\s]+)>|(https?:\/\/[^\s<>)\]]+)|\*\*(.+?)\*\*|~~(.+?)~~|(?<![\w*])\*(?!\s)([^*]+?)\*(?![\w*])|(?<!\w)_(?!\s)([^_]+?)_(?!\w)|(✓)|(✗)|(→|->)|(├──|└──|│)/g

function tokenize(text: string, style: Style = {}): Array<Inline> {
  const out: Array<Inline> = []
  let last = 0

  for (const m of text.matchAll(INLINE)) {
    const at = m.index ?? 0

    if (at > last) {
      out.push({ kind: 'text', text: text.slice(last, at), style })
    }

    if (m[1]) {
      out.push({ kind: 'code', text: m[1].slice(1, -1) })
    } else if (m[2]) {
      out.push({ kind: 'pill', raw: m[2] })
    } else if (m[3] && m[4]) {
      out.push({ kind: 'link', label: m[3], href: m[4], style })
    } else if (m[5]) {
      out.push({ kind: 'link', label: m[5], href: m[5], style })
    } else if (m[6]) {
      out.push({ kind: 'link', label: m[6], href: m[6], style })
    } else if (m[7]) {
      out.push(...tokenize(m[7], { ...style, bold: true }))
    } else if (m[8]) {
      out.push(...tokenize(m[8], { ...style, strikethrough: true }))
    } else if (m[9] || m[10]) {
      out.push(...tokenize((m[9] ?? m[10])!, { ...style, italic: true }))
    } else if (m[11]) {
      out.push({ kind: 'mark', text: m[11], color: 'success' })
    } else if (m[12]) {
      out.push({ kind: 'mark', text: m[12], color: 'error' })
    } else if (m[13]) {
      out.push({ kind: 'mark', text: m[13], color: 'suggestion' })
    } else if (m[14]) {
      out.push({ kind: 'mark', text: m[14], color: 'subtle', dim: true })
    }

    last = at + m[0].length
  }

  if (last < text.length) {
    out.push({ kind: 'text', text: text.slice(last), style })
  }

  return out
}

function splitPrefix(line: string): { indent: string; marker: string; isHeading: boolean; isQuote: boolean; body: string } {
  const m = /^(\s*)(#{1,6}\s+|[-*+]\s+|\d+[.)]\s+|>\s?)?(.*)$/.exec(line)!
  const raw = m[2] ?? ''
  const isHeading = raw.startsWith('#')
  const isQuote = raw.startsWith('>')
  let marker = ''

  if (/^[-*+]/.test(raw)) {
    marker = '- '
  } else if (/^\d/.test(raw)) {
    marker = raw.trimEnd() + ' '
  } else if (isQuote) {
    marker = '▎ '
  }

  return { indent: m[1] ?? '', marker, isHeading, isQuote, body: m[3] ?? '' }
}

export const register: Register = on => {
  on('session.start', async ($, e, next) => {
    await $.command.register({
      name: 'emphasis',
      description: 'Turn colored reply pills on or off: /emphasis on|off',
      argumentHint: 'on|off',
    })

    return next(e)
  })

  on('command.run', { command: 'emphasis' }, async ($, e) => {
    const arg = e.args.trim().toLowerCase()
    const current = await read($, isEnabled)
    const wanted = arg === 'on' ? true : arg === 'off' ? false : !current
    await update($, isEnabled, () => wanted)

    return { text: `Emphasis ${wanted ? 'on' : 'off'}.` }
  })

  on('prompt.compose', async ($, e, next) => {
    const result = await next(e)

    const inPaseo = (await $.env.get('PASEO_AGENT_ID')) !== undefined

    if (!(e.surfaces.includes('terminal') || inPaseo) || !(await read($, isEnabled))) {
      return result
    }

    const sections = [...result.sections, { id: 'emphasis:reply-style', text: guidance(), scope: 'session' as const }]

    if (inPaseo) {
      sections.push({ id: 'emphasis:collapsible', text: collapsibleGuidance(), scope: 'session' as const })
    }

    return { sections }
  })

  on('ui.render', { component: 'AssistantMessage' }, async ($, e, next) => {
    if (e.surface !== 'terminal' || !(await read($, isEnabled))) {
      return next(e)
    }

    const text = e.props.text
    const segments = segment(text)

    if (!segments.some(s => s.kind === 'pills')) {
      const withIcons = rewriteIcons(text)

      return withIcons === text ? next(e) : next({ ...e, props: { ...e.props, text: withIcons } })
    }

    const { Box, Text, Link, Markdown } = $.ui.resolve(e)

    const renderInline = (part: Inline) => {
      switch (part.kind) {
        case 'text':
          return (
            <Text bold={part.style.bold} italic={part.style.italic} strikethrough={part.style.strikethrough}>
              {part.text}
            </Text>
          )
        case 'code':
          return <Text color="permission">{part.text}</Text>
        case 'link':
          return (
            <Text bold={part.style.bold} italic={part.style.italic} color="suggestion">
              <Link href={part.href}>{part.label}</Link>
            </Text>
          )
        case 'mark':
          return (
            <Text color={part.color} dimColor={part.dim}>
              {part.text}
            </Text>
          )
        case 'pill': {
          const pill = resolvePill(part.raw)
          const tone = TONES[pill.tone]
          const label = pill.icon ? `${pill.icon} ${pill.text}` : pill.text

          return (
            <Text>
              <Text color={tone.bg}>{LEFT_CAP}</Text>
              <Text backgroundColor={tone.bg} color={tone.fg}>
                {label}
              </Text>
              <Text color={tone.bg}>{RIGHT_CAP}</Text>
            </Text>
          )
        }
      }
    }

    const rows = segments.map(s => {
      if (s.kind === 'markdown') {
        return <Markdown text={s.lines.join('\n')} />
      }

      if (s.kind === 'blank') {
        return <Text> </Text>
      }

      const { indent, marker, isHeading, isQuote, body } = splitPrefix(s.line)
      const prefix = indent + marker
      const parts = tokenize(body).map(renderInline)

      return (
        <Box flexDirection="row">
          {prefix.length > 0 && (
            <Box flexShrink={0} width={prefix.length}>
              <Text dimColor={isQuote}>{prefix}</Text>
            </Box>
          )}
          <Box flexGrow={1}>
            <Text bold={isHeading}>{parts}</Text>
          </Box>
        </Box>
      )
    })

    return (
      <Box flexDirection="row">
        <Box flexShrink={0} width={2}>
          <Text>{e.props.isFirstOfReply ? '⏺' : ' '}</Text>
        </Box>
        <Box flexDirection="column" flexGrow={1}>
          {rows}
        </Box>
      </Box>
    )
  })
}

function rewriteIcons(text: string): string {
  let fence: string | null = null

  return text
    .split('\n')
    .map(line => {
      const fenceMatch = /^\s*(```|~~~)/.exec(line)

      if (fence !== null) {
        if (fenceMatch && fenceMatch[1] === fence) {
          fence = null
        }

        return line
      }

      if (fenceMatch) {
        fence = fenceMatch[1]!

        return line
      }

      return addLabelIcons(line)
    })
    .join('\n')
}

function guidance(): string {
  const byTone = (tone: ToneName) =>
    PILL_WORDS.filter(p => p.tone === tone)
      .map(p => p.word)
      .join(', ')

  return `# Reply style: pills

The UI draws \`[[Word]]\` as a colored pill with an icon. Use pills to make status scannable; don't decorate every line.

- Red (wrong): ${byTone('red')}
- Yellow (look here): ${byTone('yellow')}
- Green (settled): ${byTone('green')}
- Solid red (consequence): ${byTone('danger')}
- Cyan (kind of work): ${byTone('cyan')}
- Blue (ownership): ${byTone('owner')}
- Soft blue (context): ${byTone('context')}, or any other short label

Rules:
- Put the pill right before the thing it describes. Detail may follow the word inside: \`[[Passed 12/12]]\`, \`[[Failed: 3 tests]]\`.
- Every failure, every ask of the user, and every claim not checked gets its pill ([[Failed]], [[Needs you]], [[Unverified]]). Follow an attention pill with the action it calls for, e.g. \`[[Unverified]] X. **Check:** run Y.\`
- Write plain words; the renderer adds icons. Never put a pill inside a table or a code block.
- Bold lead labels (**Fix:**, **Cause:**, **Next:**, **Check:**, **Do:**) and ✓ ✗ -> are styled too; use them where natural.
- These pills replace nothing else in the user's own formatting rules; follow those as before.`
}

function collapsibleGuidance(): string {
  return `# Reply style: collapsible detail

Paseo draws \`<details><summary>Title</summary>body</details>\` as a tap-to-expand block, collapsed by default.

What shows before any tap should read in about 10 seconds on a phone: the answer, the outcome, or the decision needed. Supporting material goes in \`<details>\`: evidence, file refs, reasoning, diffs, long lists, command output.

- Short replies (under ~10 lines) need no structure. Just answer.
- Longer replies: lead with the top-level points, as many as the content really has (usually 1 to 3), each one line with a bold label. Put each point's depth in a \`<details>\` right after it, with a few-word summary ("Why", "Evidence", "Diff").
- Never collapse a question, a decision, an ask of the user, or a "What I need from you" block. Those stay visible.
- Blank line after \`</summary>\` and before \`</details>\`; no nesting.`
}
