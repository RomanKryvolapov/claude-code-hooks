// usage-limits-panel — the usage-limits status line, drawn above the prompt in the VS Code extension,
// which shows no status line.
//
// What it shows is the status line's own: the Go hook, in its `panel` mode, says what the status line
// would for the input a status-line command is handed — built here from what Claude Code hands a
// plugin — the same figures, words and colours, as data. This module lays that out with the editor's
// own elements: each gauge a bar, filled in its colour, its empty part in the colours the status line
// paints its empty cells in, faded, and the lines as one table whose columns are boxes, every gauge in
// one column as wide as the band leaves.
// The status line's text would not do: its columns hold only in a monospaced face, and the extension
// draws a plugin's text in its proportional one (2.1.319 ignores Text's `mono`).
// How long the prompt cache stays warm comes from the session's transcript, since Claude Code hands a
// plugin nothing about it; the context's fill is Claude Code's own, as the status line's is, and the
// transcript's where Claude Code has none. Before the first reply, when nothing has measured the
// context, its fill is what /context estimates, marked ≈; the status line has nothing there.
//
// The order is the status line's: CONTEXT with the cache's time left, each limit's two lines, the
// config's notes, and last the token counts.
//
// It draws in the band directly above the prompt input, and only where VS Code draws the session: a
// terminal has the status line itself. A pane would not do — the extension docks one beside the
// conversation once its panel is 800 pixels wide. The band is documented for the terminal and the
// desktop app alone, but the extension draws it (2.1.319).
//
// It is loaded from this folder for every session on the machine, by naming the folder in
// CLAUDE_CODE_PLUGIN_DIRS in the `env` of ~/.claude/settings.json, and draws in a project only where
// the project has the hook's binary, which it runs. It is not kept in a project's .claude/skills/,
// where Claude Code would load it by itself: a plugin there loads only in a folder trusted under the
// exact spelling of its path, and the VS Code extension spells a Windows drive letter in lower case.
// The band is drawn again when Claude Code measures the session (after each turn, and when a limit
// moves a whole point) and every 30 seconds.

import { atom, read, update } from 'claude-code'
import type {
  BoxProps,
  ElementConstructor,
  EngineInterface,
  Register,
  RenderChildren,
  RenderElement,
  TextProps,
  Timer,
} from 'claude-code'

// The status line's own refreshInterval in .claude/settings.json: the reset countdowns, the time
// gauges and the cache's time left move with the clock, not with any event.
const REFRESH_MS = 30_000

// The injected block's timeout in .claude/settings.json: a run that has to ask the usage API waits
// for its answer, and any other is over in a fraction of a second.
const RUN_TIMEOUT_MS = 20_000

// Where a project keeps the hook's binaries, searched in this order: the same folders, and the same
// binary per platform, as the launcher in .claude/hooks/usage-limits. A plugin cannot run that
// launcher — it starts commands without a shell, and on Windows there is no sh to hand it to.
const FOLDERS = ['scripts', 'scripts/claude-code', '.claude/bin']

// Every gauge takes what its line leaves in the band, the same for all of them, and no less than this
// many cells.
const GAUGE_MIN_CELLS = 10

// How much of their colour a gauge's empty cells show, as hex opacity: a quarter, as the status line's
// light-shade cell paints a quarter of itself.
const TRACK_OPACITY = '40'

// A colour that paints nothing: a bar's border, there only to round its ends.
const CLEAR = '#00000000'

const transcriptPath = atom({ plugin: 'usage-limits-panel', key: 'transcriptPath' } as const, '')
const refreshes = atom({ plugin: 'usage-limits-panel', key: 'refreshes' } as const, 0)

// What the hook's panel mode prints; panel.go in the hook's sources says what each field holds.
type Picture = {
  context: { tokens: string; pct: number; colour: string; estimated: boolean; empty: Run[] } | null
  cache: { text: string; colour: string } | null
  tokens: string
  limits: Limit[] | null
  notes: { label: string; text: string }[]
  figures_colour: string
}
type Limit = { label: string; pct: number; colour: string; empty: Run[]; reset: string; window: LimitWindow | null }
type LimitWindow = { caption: string; pct: number; colour: string; empty: Run[]; after: string }
// A stretch of a gauge's empty cells in one colour, out of the status line's hundred.
type Run = { colour: string; cells: number }

// A run of the hook: the picture, or a problem to show in its place.
type Painted = { picture: Picture } | { problem: string }

// The two elements the band is drawn with, which every surface's table has.
type Elements = { Box: ElementConstructor<BoxProps>; Text: ElementConstructor<TextProps> }

// The binary's name on this platform, worked out once per load of the module.
let binaryName: Promise<string> | undefined
// The last run of the hook, under what it was run for: one run per refresh, transcript and binary,
// however often the band is drawn in between.
let painted: { key: string; painted: Promise<Painted> } | undefined
let refreshing: Timer | undefined

export const register: Register = on => {
  // A reload of this module runs this again, and VS Code may be attached already.
  on('session.start', async ($, e, next) => {
    if ((await $.session.surfaces()).includes('vscode')) {
      startRefreshing($)
    }
    return next(e)
  })

  // The VS Code extension, like every remote surface, attaches after the session has started.
  on('session.attach', async ($, e, next) => {
    if (e.surface === 'vscode') {
      startRefreshing($)
    }
    return next(e)
  })

  // Every settings-hook event names the session's transcript, and these two come first.
  on('classic.SessionStart', async ($, e, next) => {
    await remember($, e.transcript_path)
    return next(e)
  })
  on('classic.UserPromptSubmit', async ($, e, next) => {
    await remember($, e.transcript_path)
    return next(e)
  })

  on('session.measure', async ($, e, next) => {
    await refresh($)
    return next(e)
  })

  on('ui.render', { component: 'AbovePrompt' }, async ($, e, next) => {
    // A terminal draws the status line itself, and a survey holds the band while it asks.
    if (e.surface !== 'vscode' || e.props.hasSurvey) {
      return next(e)
    }
    const done = await paintedFor($)
    // Loaded for every session on the machine, the plugin is also in projects that carry no hook.
    if (done === undefined) {
      return next(e)
    }
    const ui: Elements = $.ui.resolve(e)
    if ('problem' in done) {
      return <ui.Text color="error">{done.problem}</ui.Text>
    }
    return band(ui, done.picture)
  })
}

function startRefreshing($: EngineInterface): void {
  refreshing ??= $.clock.every(REFRESH_MS, () => void refresh($))
}

async function refresh($: EngineInterface): Promise<void> {
  await update($, refreshes, n => n + 1)
}

async function remember($: EngineInterface, path: string): Promise<void> {
  if (path !== '' && path !== (await read($, transcriptPath))) {
    await update($, transcriptPath, () => path)
  }
}

// Reading the refresh count and the transcript here, while the band is drawn, is what draws it again
// when either is written. Undefined in a project without the hook's binary.
async function paintedFor($: EngineInterface): Promise<Painted | undefined> {
  const [count, transcript, root] = await Promise.all([read($, refreshes), read($, transcriptPath), $.session.root()])
  const binary = await findBinary($, root)
  if (binary === undefined) {
    return undefined
  }
  const key = `${count} ${transcript} ${binary}`
  if (painted?.key !== key) {
    painted = { key, painted: paint($, binary, root, transcript) }
  }
  return painted.painted
}

// Runs the hook. Anything that stops it is shown where the figures would be, rather than a band left
// empty.
async function paint($: EngineInterface, binary: string, root: string, transcript: string): Promise<Painted> {
  try {
    const ran = await $.process.run([binary, '--mode', 'panel'], {
      stdin: await statusInput($, transcript),
      env: { CLAUDE_PROJECT_DIR: root },
      timeoutMs: RUN_TIMEOUT_MS,
    })
    if (ran.exitCode !== 0) {
      return { problem: `usage-limits exited with ${ran.exitCode}: ${ran.stderr.trim()}` }
    }
    // A binary older than the panel mode prints nothing for it, as for every mode it does not know,
    // and so does one that failed: the hook never takes a session down.
    if (ran.stdout.trim() === '') {
      return {
        problem: `${binary} printed nothing: it is older than this plugin, or it failed. Copy in the current binaries.`,
      }
    }
    const picture = readPicture(ran.stdout)
    if (picture === undefined) {
      return { problem: `${binary} printed what this plugin cannot read: the two are of different versions.` }
    }
    return { picture }
  } catch (error) {
    return { problem: `usage-limits could not run: ${error instanceof Error ? error.message : String(error)}` }
  }
}

async function findBinary($: EngineInterface, root: string): Promise<string | undefined> {
  binaryName ??= platformBinary($)
  const name = await binaryName
  for (const folder of FOLDERS) {
    const path = `${root}/${folder}/${name}`
    if (await $.fs.exists(path)) {
      return path
    }
  }
  return undefined
}

async function platformBinary($: EngineInterface): Promise<string> {
  if ((await $.env.get('OS')) === 'Windows_NT') {
    return 'usage-limits-windows-amd64.exe'
  }
  const { stdout } = await $.process.run(['uname', '-s'])
  return stdout.startsWith('Darwin') ? 'usage-limits-darwin-arm64' : 'usage-limits-linux-amd64'
}

// The input a status-line command is handed, from what Claude Code hands a plugin. The context's fill
// is Claude Code's own wherever it has one, as a status line's is: it has none in a fresh session or
// one just compacted, until the next response — when a status line is handed none either and reads
// the transcript, as the hook then does. Then, too, the input carries what /context estimates the
// context holds, which the hook takes only where the transcript has no reading. The prompt cache is
// read off the transcript always: nothing a plugin is handed says how long it stays warm.
async function statusInput($: EngineInterface, transcript: string): Promise<string> {
  const [usage, model] = await Promise.all([$.session.usage(), $.session.model()])
  const tokens = usage.context.tokens
  // Counted locally, so a summary breakdown sends no request. An estimate that cannot be had leaves the
  // band without a CONTEXT line until the first response, as the status line is, and no worse.
  const estimate =
    tokens === undefined
      ? await $.session
          .usage({ breakdown: 'summary' })
          .then(counted => counted.context.breakdown?.totalTokens)
          .catch(() => undefined)
      : undefined
  const window = (kind: string) => {
    const limit = usage.rateLimits.find(one => one.kind === kind)
    const resetsAt = limit?.resetsAt === undefined ? Number.NaN : Date.parse(limit.resetsAt) / 1000
    return limit === undefined || Number.isNaN(resetsAt)
      ? undefined
      : { used_percentage: limit.percentUsed, resets_at: resetsAt }
  }
  return JSON.stringify({
    transcript_path: transcript === '' ? undefined : transcript,
    model: { id: model },
    // A status line is handed the usage the reading came from; the hook asks only that there is one.
    context_window: {
      context_window_size: usage.context.window,
      ...(tokens !== undefined && { total_input_tokens: tokens, current_usage: {} }),
    },
    context_estimate: estimate,
    rate_limits: { five_hour: window('five_hour'), seven_day: window('seven_day') },
  })
}

function readPicture(stdout: string): Picture | undefined {
  let value: unknown
  try {
    value = JSON.parse(stdout)
  } catch {
    return undefined
  }
  const is = (v: unknown, type: string) => typeof v === type
  const record = (v: unknown): v is Record<string, unknown> => typeof v === 'object' && v !== null
  const isEmpty = (v: unknown) =>
    Array.isArray(v) && v.every(run => record(run) && is(run.colour, 'string') && is(run.cells, 'number'))
  const isLimit = (v: unknown) =>
    record(v) &&
    is(v.label, 'string') &&
    is(v.pct, 'number') &&
    is(v.colour, 'string') &&
    isEmpty(v.empty) &&
    is(v.reset, 'string') &&
    (v.window === null ||
      (record(v.window) &&
        is(v.window.caption, 'string') &&
        is(v.window.pct, 'number') &&
        is(v.window.colour, 'string') &&
        isEmpty(v.window.empty) &&
        is(v.window.after, 'string')))
  const fits =
    record(value) &&
    (value.context === null ||
      (record(value.context) &&
        is(value.context.tokens, 'string') &&
        is(value.context.pct, 'number') &&
        is(value.context.colour, 'string') &&
        is(value.context.estimated, 'boolean') &&
        isEmpty(value.context.empty))) &&
    (value.cache === null ||
      (record(value.cache) && is(value.cache.text, 'string') && is(value.cache.colour, 'string'))) &&
    is(value.tokens, 'string') &&
    (value.limits === null || (Array.isArray(value.limits) && value.limits.every(isLimit))) &&
    Array.isArray(value.notes) &&
    value.notes.every(n => record(n) && is(n.label, 'string') && is(n.text, 'string')) &&
    is(value.figures_colour, 'string')
  return fits ? (value as Picture) : undefined
}

// One line of the table: what stands in each of its columns, its gauge included.
type Row = {
  label: RenderChildren
  // Right-aligned: CONTEXT's counter, or a limit's percentage.
  figure: RenderChildren
  // Right-aligned: CONTEXT's percentage, or LIMIT, TIME, WORK.
  caption: RenderChildren
  gauge: { colour: string; empty: Run[] }
  // The cache's time left, or a limit's reset.
  after: RenderChildren
}

// The status line's picture: CONTEXT and each limit's two lines as one table, every gauge in one
// column as wide as the band leaves, then the config's notes and, last, the token counts.
function band(ui: Elements, picture: Picture): RenderElement {
  const { Box, Text } = ui
  const figures = picture.figures_colour
  const cache =
    picture.cache === null ? null : (
      <Text>
        CACHE <Text color={picture.cache.colour}>{picture.cache.text}</Text>
      </Text>
    )
  const rows: Row[] = []
  const context = picture.context
  if (context !== null) {
    rows.push({
      label: <Text>CONTEXT</Text>,
      // ≈ while it is an estimate.
      figure: (
        <Text color={context.colour}>
          {context.estimated ? '≈' : ''}
          {context.tokens}
        </Text>
      ),
      caption: <Text color={context.colour}>{context.pct} %</Text>,
      gauge: { colour: context.colour, empty: context.empty },
      after: cache,
    })
  }
  for (const limit of picture.limits ?? []) {
    rows.push({
      label: <Text>{limit.label}</Text>,
      figure: <Text color={figures}>{limit.pct} %</Text>,
      caption: <Text>LIMIT</Text>,
      gauge: { colour: limit.colour, empty: limit.empty },
      after:
        limit.reset === '' ? null : (
          <Text>
            RESET AT <Text color={figures}>{limit.reset}</Text>
          </Text>
        ),
    })
    const window = limit.window
    if (window !== null) {
      rows.push({
        label: null,
        figure: <Text color={figures}>{window.pct} %</Text>,
        caption: <Text>{window.caption}</Text>,
        gauge: { colour: window.colour, empty: window.empty },
        after: window.after === '' ? null : <Text>RESET AFTER {figuresIn(ui, window.after, figures)}</Text>,
      })
    }
  }

  const lines: RenderChildren[] = []
  // Without a context reading, the cache's time left heads the band alone.
  if (context === null && cache !== null) {
    lines.push(cache)
  }
  if (rows.length > 0) {
    lines.push(table(ui, rows))
  }
  if (picture.limits === null) {
    lines.push(<Text>{'LIMITS -> N/A'}</Text>)
  }
  for (const note of picture.notes) {
    lines.push(
      <Box columnGap={2}>
        <Text>{note.label}</Text>
        <Text>{note.text}</Text>
      </Box>,
    )
  }
  if (picture.tokens !== '') {
    lines.push(<Text>{picture.tokens}</Text>)
  }
  return <Box flexDirection="column">{lines}</Box>
}

// The rows as columns of boxes, each cell exactly a line high so that a row stays level across them.
// The text columns are as wide as their widest cell; the gauges' column takes the rest of the band, so
// every gauge is as wide as the others; the last column gives way, clipped, where the band is too
// narrow for it.
function table(ui: Elements, rows: Row[]): RenderElement {
  const { Box } = ui
  const column = (props: BoxProps, cell: (row: Row) => RenderChildren) => (
    <Box flexDirection="column" {...props}>
      {rows.map(row => (
        <Box height={1} alignItems="center" flexWrap="nowrap">
          {cell(row)}
        </Box>
      ))}
    </Box>
  )
  return (
    <Box columnGap={2} flexWrap="nowrap">
      {column({ flexShrink: 0 }, row => row.label)}
      {column({ flexShrink: 0, alignItems: 'flex-end' }, row => row.figure)}
      {column({ flexShrink: 0, alignItems: 'flex-end' }, row => row.caption)}
      {column({ flexGrow: 1, minWidth: GAUGE_MIN_CELLS }, row => gauge(ui, row.gauge.colour, row.gauge.empty))}
      {column({ flexShrink: 1, overflow: 'hidden' }, row => row.after)}
    </Box>
  )
}

// A gauge as a bar a line high, as wide as what holds it: its empty part as the status line paints its
// empty cells, a cell a percent of the bar, each run in its colour faded, and the filled part, in the
// gauge's colour, the rest — as long as the status line's, since the empty part is.
function gauge({ Box }: Elements, colour: string, empty: Run[]): RenderElement {
  return (
    <Box width="100%" height={1} alignItems="center" flexWrap="nowrap">
      <Box width="100%" height="50%" borderStyle="round" borderColor={CLEAR} overflow="hidden" flexWrap="nowrap">
        <Box flexGrow={1} height="100%" backgroundColor={colour} />
        {empty.map(run => (
          <Box width={`${run.cells}%`} flexShrink={0} height="100%" backgroundColor={`${run.colour}${TRACK_OPACITY}`} />
        ))}
      </Box>
    </Box>
  )
}

// The figures of a wait such as "6 days 19:49" in their colour, its words left plain, as the status
// line paints them.
function figuresIn({ Text }: Elements, text: string, colour: string): RenderChildren[] {
  return text
    .split(' ')
    .flatMap((word, i) => [i > 0 ? ' ' : '', /\d/.test(word) ? <Text color={colour}>{word}</Text> : word])
}
