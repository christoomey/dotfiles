import { expect, test } from 'claude-code/testing'

const TEXT = [
  'Intro paragraph with **bold**.',
  '',
  '[[Done]] see [the PR](https://example.com/pr) and `code` -> ok ✓',
  '- [[Failed: 3 tests]] *italic* ~~gone~~',
  '',
  '| a | [[Done]] |',
  '|---|---|',
  '',
  '```',
  '[[Done]] in fence',
  '```',
  '**Fix:** later',
].join('\n')

const MSG = (text: string) =>
  ({
    plugin: 'emphasis',
    component: 'AssistantMessage',
    requestId: 'm1',
    surface: 'terminal',
    props: { text, isFirstOfReply: true },
  }) as const

test('pill lines are drawn by hand, the rest as markdown', async ($, on) => {
  on('state.get', () => ({ value: { value: true, version: 1 } }))
  const ui = await $.ui.mount(MSG(TEXT))
  const pill = await ui.find({ type: 'Text', text: /Failed: 3 tests/ })
  expect(pill).toBeTruthy()
  const link = await ui.find({ type: 'Link' })
  expect(link).toBeTruthy()
  const md = await ui.findAll({ type: 'Markdown' })
  console.log(JSON.stringify(md.map(m => m.props?.text ?? m.text)))
  await ui.unmount()
})

test('no pills: engine draws, icons added', async ($, on) => {
  on('state.get', () => ({ value: { value: true, version: 1 } }))
  let seen = ''
  on('ui.render', ($, e) => {
    seen = (e.props as { text: string }).text
    const { Text } = $.ui.resolve(e)
    return <Text>engine</Text>
  })
  const ui = await $.ui.mount(MSG('**Fix:** thing\n## Plan'))
  console.log(JSON.stringify(seen))
  await ui.unmount()
})
