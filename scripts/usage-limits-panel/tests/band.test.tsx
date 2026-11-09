import type { On, RenderSurface, SessionContextBreakdown } from 'claude-code'
import { expect, mock, test } from 'claude-code/testing'

// The band above the prompt as a surface asks for it, at a width in cells.
const band = (bodyColumns: number, hasSurvey = false) =>
  ({
    plugin: 'usage-limits-panel',
    component: 'AbovePrompt',
    props: {
      hasSurvey,
      isCollapsed: false,
      isWorking: false,
      maxRows: 12,
      bodyColumns,
      scroll: { offset: 0, bodyRows: 12 },
      view: {},
    },
  }) as const

// What the hook's panel mode prints for a session with a context reading and a warm cache, on a plan
// with a 5-hour window, whose spend runs ahead of its clock, and the 7-day one on working time. Each
// gauge's empty part comes as runs of one colour, out of the status line's hundred cells.
const PICTURE = {
  context: {
    tokens: '308k/1m',
    pct: 31,
    colour: '#a3f224',
    estimated: false,
    empty: [
      { colour: '#c6f224', cells: 30 },
      { colour: '#f2a324', cells: 30 },
      { colour: '#f22424', cells: 9 },
    ],
  },
  cache: { text: '1:00', colour: '#24f224' },
  tokens: 'SESSION: FRESH 2m CACHED 108m, ACCOUNT 5H: FRESH 19m CACHED 1.1b, 7D: FRESH 20m CACHED 1.1b',
  limits: [
    {
      label: '5 HOURS',
      pct: 58,
      colour: '#f22424',
      empty: [
        { colour: '#2458f2', cells: 20 },
        { colour: '#a424f2', cells: 12 },
        { colour: '#f22424', cells: 10 },
      ],
      reset: '23:50',
      window: { caption: 'TIME', pct: 40, colour: '#f2f2f2', empty: [{ colour: '#f2f2f2', cells: 60 }], after: '3:00' },
    },
    {
      label: '7 DAYS',
      pct: 6,
      colour: '#2458f2',
      empty: [
        { colour: '#2458f2', cells: 50 },
        { colour: '#f22424', cells: 44 },
      ],
      reset: 'MON 17:00',
      window: {
        caption: 'WORK',
        pct: 7,
        colour: '#f2f2f2',
        empty: [{ colour: '#f2f2f2', cells: 93 }],
        after: '6 days 19:49',
      },
    },
  ],
  notes: [],
  figures_colour: '#ffb400',
}

const BINARY = 'D:/project/scripts/claude-code/usage-limits-windows-amd64.exe'

type Run = { argv: readonly string[]; stdin?: string; env?: Record<string, string> }

// What Claude Code knows of the context: what the last response was answered over, once there was
// one, and what /context estimates, or that it cannot; `breakdowns` counts the times the estimate was
// asked for.
type Context = { tokens?: number; estimate?: number; cannotEstimate?: boolean; breakdowns: number }

// The engine beneath the plugin: a Windows project, a binary only where `isThere` says, and a hook
// that prints `stdout`, or what it says at the time. The engine hands a path on in the platform's own
// spelling.
const project = (
  on: On,
  runs: Run[],
  isThere: (path: string) => boolean,
  stdout: string | (() => string) = JSON.stringify(PICTURE),
  context: Context = { breakdowns: 0 },
) => {
  mock.env(on, { OS: 'Windows_NT' })
  // The engine's own band, empty while no survey asks there.
  on('ui.render', ($, e) => {
    const { Box } = $.ui.resolve(e)
    return <Box />
  })
  on('session.root', () => ({ value: 'D:/project' }))
  on('fs.exists', (_, e) => ({ value: isThere(e.path.replaceAll('\\', '/')) }))
  on('session.model', () => ({ value: 'claude-opus-5-5' }))
  on('session.usage', (_, e) => {
    const breakdown = e?.breakdown !== undefined
    if (breakdown) {
      context.breakdowns += 1
      if (context.cannotEstimate) {
        throw new Error('no session is bound')
      }
    }
    return {
      value: {
        startedAt: 0,
        context: {
          window: 1_000_000,
          ...(context.tokens !== undefined && { tokens: context.tokens }),
          ...(breakdown &&
            context.estimate !== undefined && {
              breakdown: { totalTokens: context.estimate } as SessionContextBreakdown,
            }),
        },
        rateLimits: [
          { kind: 'five_hour', percentUsed: 58, resetsAt: '2026-11-09T21:50:00Z' },
          { kind: 'spend_limit', percentUsed: 3 },
        ],
      },
    }
  })
  on('process.run', (_, e) => {
    runs.push({ argv: e.argv, stdin: e.init?.stdin, env: e.init?.env })
    const printed = typeof stdout === 'function' ? stdout() : stdout
    return {
      value: { exitCode: 0, stdout: printed, stderr: '', isStdoutTruncated: false, isStderrTruncated: false },
    }
  })
}

test('the band draws the status line in VS Code alone, its gauges as bars', async ($, on) => {
  const runs: Run[] = []
  // In the second of the three folders the launcher searches.
  project(on, runs, path => path === BINARY)
  on('session.measure', (_, e) => ({ changed: e.changed }))
  on('classic.UserPromptSubmit', () => ({}))

  // A terminal draws the status line itself, and a survey holds the band: nothing drawn, nothing run.
  for (const target of [
    { ...band(100), surface: 'terminal' },
    { ...band(100), surface: 'desktop' },
    { ...band(100, true), surface: 'vscode' },
  ] as const) {
    const ui = await $.ui.mount(target)
    expect(await ui.find({ type: 'Text', text: 'CONTEXT' })).toBeUndefined()
    await ui.unmount()
  }
  expect(runs.length).toBe(0)

  const ui = await $.ui.mount({ ...band(100), surface: 'vscode' })
  // CONTEXT's counter and percentage in its gauge's colour, then the cache's time left in its own.
  expect((await ui.find({ type: 'Text', text: /^308k\/1m$/ }))?.props.color).toBe('#a3f224')
  expect((await ui.find({ type: 'Text', text: /^31 %$/ }))?.props.color).toBe('#a3f224')
  expect((await ui.find({ type: 'Text', text: /^CACHE / }))?.text).toBe('CACHE 1:00')
  expect((await ui.find({ type: 'Text', text: /^1:00$/ }))?.props.color).toBe('#24f224')
  // The token counts are the band's last line, under the limits.
  const drawn = await ui.drawn()
  const lines = 'children' in drawn ? (drawn.children ?? []) : []
  expect(lines.at(-1)).toMatchObject({ type: 'Text', children: [PICTURE.tokens] })
  expect(JSON.stringify(lines.slice(0, -1))).not.toContain('SESSION')

  // Every gauge is a bar: its empty part the hook's runs, a percent of the bar a cell, each in its
  // colour faded, and the filled part, in the gauge's colour, the rest — CONTEXT, then each limit's
  // LIMIT gauge over the gauge of its window. The bar's own border paints nothing.
  const boxes = await ui.findAll({ type: 'Box' })
  const bars = boxes.filter(box => box.props.height === '50%')
  expect(bars.map(bar => [bar.props.backgroundColor, bar.props.borderColor])).toEqual(
    Array(5).fill([undefined, '#00000000']),
  )
  const parts = boxes.filter(box => box.props.height === '100%')
  expect(
    parts.map(box => [box.props.flexGrow === 1 ? 'the rest' : box.props.width, box.props.backgroundColor]),
  ).toEqual([
    ['the rest', '#a3f224'],
    ['30%', '#c6f22440'],
    ['30%', '#f2a32440'],
    ['9%', '#f2242440'],
    ['the rest', '#f22424'],
    ['20%', '#2458f240'],
    ['12%', '#a424f240'],
    ['10%', '#f2242440'],
    ['the rest', '#f2f2f2'],
    ['60%', '#f2f2f240'],
    ['the rest', '#2458f2'],
    ['50%', '#2458f240'],
    ['44%', '#f2242440'],
    ['the rest', '#f2f2f2'],
    ['93%', '#f2f2f240'],
  ])

  // The figures are amber; a wait's words are not.
  expect((await ui.findAll({ type: 'Text', text: /^(58|40|6|7) %$/ })).map(text => text.props.color)).toEqual([
    '#ffb400',
    '#ffb400',
    '#ffb400',
    '#ffb400',
  ])
  expect((await ui.findAll({ type: 'Text', text: /^(LIMIT|TIME|WORK)$/ })).map(text => text.text)).toEqual([
    'LIMIT',
    'TIME',
    'LIMIT',
    'WORK',
  ])
  expect((await ui.findAll({ type: 'Text', text: /^RESET / })).map(text => text.text)).toEqual([
    'RESET AT 23:50',
    'RESET AFTER 3:00',
    'RESET AT MON 17:00',
    'RESET AFTER 6 days 19:49',
  ])
  expect((await ui.find({ type: 'Text', text: /^MON 17:00$/ }))?.props.color).toBe('#ffb400')
  expect((await ui.find({ type: 'Text', text: /^19:49$/ }))?.props.color).toBe('#ffb400')
  expect(await ui.find({ type: 'Text', text: /^days$/ })).toBeUndefined()
  await ui.unmount()

  // One run of the hook, with the input a status line gets.
  expect(runs.length).toBe(1)
  expect(runs[0]?.argv).toEqual([BINARY, '--mode', 'panel'])
  expect(runs[0]?.env).toEqual({ CLAUDE_PROJECT_DIR: 'D:/project' })
  const input = JSON.parse(runs[0]?.stdin ?? '{}')
  expect(input.model).toEqual({ id: 'claude-opus-5-5' })
  expect(input.context_window).toEqual({ context_window_size: 1_000_000 })
  expect(input.rate_limits).toEqual({
    five_hour: { used_percentage: 58, resets_at: Date.parse('2026-11-09T21:50:00Z') / 1000 },
  })
  expect(input.transcript_path).toBeUndefined()

  // Once a settings-hook event has named the transcript, and once Claude Code has measured the
  // session, the hook is run again, and handed the transcript.
  await $.classic.UserPromptSubmit({ prompt: 'hi', transcript_path: 'D:/t/s.jsonl' })
  await $.session.measure({ context: { window: 1_000_000 }, rateLimits: [], changed: ['context'] })
  const again = await $.ui.mount({ ...band(100), surface: 'vscode' })
  await again.find({ type: 'Text', text: 'CONTEXT' })
  await again.unmount()
  expect(runs.length).toBe(2)
  expect(JSON.parse(runs[1]?.stdin ?? '{}').transcript_path).toBe('D:/t/s.jsonl')
})

test('before the first reply the band shows what /context estimates, marked as an estimate', async ($, on) => {
  const runs: Run[] = []
  const context: Context = { estimate: 48_000, breakdowns: 0 }
  let stdout = JSON.stringify({
    ...PICTURE,
    context: {
      tokens: '48k/1m',
      pct: 5,
      colour: '#38f224',
      estimated: true,
      empty: [{ colour: '#3df224', cells: 95 }],
    },
  })
  project(
    on,
    runs,
    path => path === BINARY,
    () => stdout,
    context,
  )
  on('session.measure', (_, e) => ({ changed: e.changed }))

  // No response has measured the context: the hook is handed no reading of Claude Code's but the
  // estimate, and the band marks it.
  let ui = await $.ui.mount({ ...band(100), surface: 'vscode' })
  expect((await ui.find({ type: 'Text', text: /48k\/1m$/ }))?.text).toBe('≈48k/1m')
  await ui.unmount()
  let input = JSON.parse(runs[0]?.stdin ?? '{}')
  expect(input.context_window).toEqual({ context_window_size: 1_000_000 })
  expect(input.context_estimate).toBe(48_000)
  expect(context.breakdowns).toBe(1)

  // Once a response has, the hook is handed Claude Code's own reading, as a status line is, and
  // nothing is estimated.
  context.tokens = 52_000
  stdout = JSON.stringify(PICTURE)
  await $.session.measure({ context: { window: 1_000_000, tokens: 52_000 }, rateLimits: [], changed: ['context'] })
  ui = await $.ui.mount({ ...band(100), surface: 'vscode' })
  expect((await ui.find({ type: 'Text', text: /308k\/1m$/ }))?.text).toBe('308k/1m')
  await ui.unmount()
  input = JSON.parse(runs[1]?.stdin ?? '{}')
  expect(input.context_window).toEqual({
    context_window_size: 1_000_000,
    total_input_tokens: 52_000,
    current_usage: {},
  })
  expect(input.context_estimate).toBeUndefined()
  expect(context.breakdowns).toBe(1)
})

test('an estimate Claude Code cannot make leaves the band drawn, without it', async ($, on) => {
  const runs: Run[] = []
  const context: Context = { cannotEstimate: true, breakdowns: 0 }
  project(on, runs, path => path === BINARY, JSON.stringify({ ...PICTURE, context: null }), context)

  const ui = await $.ui.mount({ ...band(100), surface: 'vscode' })
  expect(await ui.find({ type: 'Text', text: /could not run/ })).toBeUndefined()
  expect((await ui.findAll({ type: 'Text', text: /^RESET AT / })).length).toBe(2)
  await ui.unmount()
  expect(context.breakdowns).toBe(1)
  expect(JSON.parse(runs[0]?.stdin ?? '{}').context_estimate).toBeUndefined()
})

test('every gauge stands in one column, which takes what the band leaves', async ($, on) => {
  project(on, [], path => path === BINARY)

  for (const columns of [100, 60]) {
    const ui = await $.ui.mount({ ...band(columns), surface: 'vscode' })
    const drawn = await ui.drawn()
    const lines = 'children' in drawn ? (drawn.children ?? []) : []
    // One table, CONTEXT's line and each limit's two lines its rows, then the tokens.
    expect(lines).toHaveLength(2)
    const table = lines[0] as {
      props: Record<string, unknown>
      children: { props: Record<string, unknown>; children: unknown[] }[]
    }
    expect(table.props.flexWrap).toBe('nowrap')
    expect(table.children.map(column => column.children.length)).toEqual([5, 5, 5, 5, 5])
    // The gauges' column grows into the band whatever its width; the others keep their text's width.
    expect(table.children.map(column => [column.props.flexGrow, column.props.minWidth])).toEqual([
      [undefined, undefined],
      [undefined, undefined],
      [undefined, undefined],
      [1, 10],
      [undefined, undefined],
    ])
    // So every gauge is as wide as its cell, none a width of its own.
    const gauges = (await ui.findAll({ type: 'Box' })).filter(
      box => box.props.height === 1 && box.props.width === '100%',
    )
    expect(gauges).toHaveLength(5)
    await ui.unmount()
  }
})

test('a project without the hook draws nothing', async ($, on) => {
  const runs: Run[] = []
  project(on, runs, () => false)

  const ui = await $.ui.mount({ ...band(100), surface: 'vscode' })
  expect(await ui.find({ type: 'Text' })).toBeUndefined()
  expect(runs.length).toBe(0)
  await ui.unmount()
})

test('a binary that prints nothing, or what the plugin cannot read, is said where the figures would be', async ($, on) => {
  let stdout = ''
  project(
    on,
    [],
    path => path === BINARY,
    () => stdout,
  )
  on('session.measure', (_, e) => ({ changed: e.changed }))

  let ui = await $.ui.mount({ ...band(100), surface: 'vscode' })
  let said = await ui.find({ type: 'Text', text: /printed nothing/ })
  expect(said?.props.color).toBe('error')
  expect(said?.text).toBe(
    `${BINARY} printed nothing: it is older than this plugin, or it failed. Copy in the current binaries.`,
  )
  await ui.unmount()

  // The panel mode of an earlier build of this plugin painted the status line's text.
  stdout = 'CONTEXT  \x1b[38;2;242;61;12m877k/1m\x1b[0m'
  await $.session.measure({ context: { window: 1_000_000 }, rateLimits: [], changed: ['context'] })
  ui = await $.ui.mount({ ...band(100), surface: 'vscode' })
  said = await ui.find({ type: 'Text', text: /cannot read/ })
  expect(said?.props.color).toBe('error')
  expect(said?.text).toBe(`${BINARY} printed what this plugin cannot read: the two are of different versions.`)
  await ui.unmount()

  // A hook older than the gauges' empty parts printed the picture without them, wherever a gauge stands.
  const before = <T extends { empty: unknown }>({ empty: _, ...rest }: T) => rest
  for (const picture of [
    { ...PICTURE, context: before(PICTURE.context) },
    { ...PICTURE, limits: PICTURE.limits.map((limit, i) => (i === 0 ? before(limit) : limit)) },
    { ...PICTURE, limits: PICTURE.limits.map(limit => ({ ...limit, window: before(limit.window) })) },
  ]) {
    stdout = JSON.stringify(picture)
    await $.session.measure({ context: { window: 1_000_000 }, rateLimits: [], changed: ['context'] })
    ui = await $.ui.mount({ ...band(100), surface: 'vscode' })
    expect((await ui.find({ type: 'Text', text: /cannot read/ }))?.text).toBe(
      `${BINARY} printed what this plugin cannot read: the two are of different versions.`,
    )
    await ui.unmount()
  }
})

test('no figures at all read as the status line writes them, and the config notes follow', async ($, on) => {
  project(
    on,
    [],
    path => path === BINARY,
    JSON.stringify({
      context: null,
      cache: null,
      tokens: '',
      limits: null,
      notes: [{ label: 'CONFIG', text: 'time_zone "Mars/Olympus" is not a zone; the machine\'s own is used' }],
      figures_colour: '#ffb400',
    }),
  )

  const ui = await $.ui.mount({ ...band(100), surface: 'vscode' })
  expect(await ui.find({ type: 'Text', text: 'CONTEXT' })).toBeUndefined()
  expect((await ui.find({ type: 'Text', text: /N\/A/ }))?.text).toBe('LIMITS -> N/A')
  expect((await ui.find({ type: 'Text', text: 'CONFIG' }))?.text).toBe('CONFIG')
  expect(await ui.find({ type: 'Text', text: /Mars\/Olympus/ })).toBeDefined()
  await ui.unmount()
})

test('once VS Code attaches the band is drawn again every 30 seconds', async ($, on) => {
  const runs: Run[] = []
  project(on, runs, path => path === BINARY)
  const clock = mock.clock(on)
  let surfaces: RenderSurface[] = []
  on('session.start', (_, e) => ({ cwd: e.cwd }))
  on('session.attach', (_, e) => ({ clientId: e.clientId }))
  on('session.surfaces', () => ({ value: surfaces }))

  // The extension attaches after the session has started.
  await $.session.start({ cwd: '/', surface: null, isInteractive: false })
  surfaces = ['vscode']
  await $.session.attach({ surface: 'vscode', clientId: 'vscode:default' })
  const ui = await $.ui.mount({ ...band(100), surface: 'vscode' })
  await ui.find({ type: 'Text', text: 'CONTEXT' })
  expect(runs.length).toBe(1)

  await clock.advance(29_000)
  await ui.find({ type: 'Text', text: 'CONTEXT' })
  expect(runs.length).toBe(1)
  await clock.advance(1_000)
  await ui.find({ type: 'Text', text: 'CONTEXT' })
  expect(runs.length).toBe(2)
  await ui.unmount()
})
