// What the usage-limits-panel plugin keeps for the session ($.state), and so what outlives a reload
// of its module. The band's drawing reads both values, so writing either draws it again.

// The session's transcript, which the hook counts the session's tokens and reads the prompt cache
// from: empty until a settings-hook event has named it.
export type TranscriptPath = string

declare module 'claude-code' {
  interface PluginState {
    // refreshes: how many times the band has been asked to draw itself again.
    'usage-limits-panel': { transcriptPath: TranscriptPath; refreshes: number }
  }
}
