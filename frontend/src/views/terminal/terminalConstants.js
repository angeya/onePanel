/**
 * TERMINAL_THEMES 定义终端支持的主题配色。
 * 保持与系统主题选项一致，便于前端直接按 theme key 取值。
 */
export const TERMINAL_THEMES = {
  dark: {
    background: '#1e1e1e',
    foreground: '#d4d4d4',
    cursor: '#ffffff',
    cursorAccent: '#000000',
    selection: 'rgba(255, 255, 255, 0.3)',
    black: '#000000',
    red: '#cd3131',
    green: '#0dbc79',
    yellow: '#e5e510',
    blue: '#2472c8',
    magenta: '#bc3fbc',
    cyan: '#11a8cd',
    white: '#e5e5e5',
    brightBlack: '#666666',
    brightRed: '#f14c4c',
    brightGreen: '#23d18b',
    brightYellow: '#f5f543',
    brightBlue: '#3b8eea',
    brightMagenta: '#d670d6',
    brightCyan: '#29b8db',
    brightWhite: '#e5e5e5'
  },
  light: {
    background: '#ffffff',
    foreground: '#383a42',
    cursor: '#526eff',
    cursorAccent: '#ffffff',
    selection: 'rgba(82, 110, 255, 0.25)',
    black: '#383a42',
    red: '#e45649',
    green: '#50a14f',
    yellow: '#c18401',
    blue: '#4078f2',
    magenta: '#a626a4',
    cyan: '#0184bc',
    white: '#a0a1a7',
    brightBlack: '#4f525e',
    brightRed: '#e06c75',
    brightGreen: '#98c379',
    brightYellow: '#e5c07b',
    brightBlue: '#61afef',
    brightMagenta: '#c678dd',
    brightCyan: '#56b6c2',
    brightWhite: '#ffffff'
  },
  blue: {
    background: '#0d1b2a',
    foreground: '#e0e8f0',
    cursor: '#4da6ff',
    cursorAccent: '#0d1b2a',
    selection: 'rgba(77, 166, 255, 0.3)',
    black: '#0d1b2a',
    red: '#ff6b6b',
    green: '#5cb85c',
    yellow: '#f0ad4e',
    blue: '#4da6ff',
    magenta: '#c678dd',
    cyan: '#56b6c2',
    white: '#e0e8f0',
    brightBlack: '#3d5a78',
    brightRed: '#ff8a8a',
    brightGreen: '#7dce7d',
    brightYellow: '#ffc85a',
    brightBlue: '#79bfff',
    brightMagenta: '#d8a0ff',
    brightCyan: '#7dd8e0',
    brightWhite: '#ffffff'
  },
  green: {
    background: '#1a2e1a',
    foreground: '#d8e8d8',
    cursor: '#67c23a',
    cursorAccent: '#1a2e1a',
    selection: 'rgba(103, 194, 58, 0.3)',
    black: '#1a2e1a',
    red: '#e06c75',
    green: '#67c23a',
    yellow: '#e6a23c',
    blue: '#61afef',
    magenta: '#c678dd',
    cyan: '#56b6c2',
    white: '#d8e8d8',
    brightBlack: '#3d6a3d',
    brightRed: '#f08890',
    brightGreen: '#85ce5f',
    brightYellow: '#f5c462',
    brightBlue: '#79bfff',
    brightMagenta: '#d8a0ff',
    brightCyan: '#7dd8e0',
    brightWhite: '#f0f8f0'
  }
}

/**
 * SEARCH_DECORATIONS 定义 xterm 搜索结果高亮样式。
 */
export const SEARCH_DECORATIONS = {
  matchBackground: '#FFC800',
  matchOverviewRuler: '#FFC800',
  activeMatchBackground: '#FFA000',
  activeMatchColorOverviewRuler: '#FFA000'
}

/**
 * SSH_CONNECT_INDICATORS 表示常见的 SSH 登录成功输出特征。
 */
export const SSH_CONNECT_INDICATORS = ['Last login', 'Welcome to']

/**
 * SSH_DISCONNECT_PATTERNS 表示常见的 SSH 断开提示关键字。
 */
export const SSH_DISCONNECT_PATTERNS = [
  'Connection to ',
  'connection closed',
  'Disconnected from',
  'Network error: Connection',
  'Connection timed out',
  'Connection refused',
  'No route to host',
  'broken pipe'
]
