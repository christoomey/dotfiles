export type Toggle = boolean

declare module 'claude-code' {
  interface PluginState {
    emphasis: { isEnabled: Toggle }
  }
}
