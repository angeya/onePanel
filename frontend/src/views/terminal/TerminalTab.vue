<template>
  <div class="terminal-tab-container">
    <div ref="terminalRef" class="terminal-content"></div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import { SearchAddon } from 'xterm-addon-search'
import { WebLinksAddon } from 'xterm-addon-web-links'
import 'xterm/css/xterm.css'
import { Start, Write, Stop, Resize } from '../../../wailsjs/go/main/PtyService'
import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime'

const SEARCH_DECORATIONS = {
  matchBackground: '#FFC800',
  matchOverviewRuler: '#FFC800',
  activeMatchBackground: '#FFA000',
  activeMatchColorOverviewRuler: '#FFA000'
}

const TERMINAL_THEMES = {
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

const props = defineProps({
  tabId: { type: String, required: true },
  shell: { type: String, default: 'cmd.exe' },
  theme: { type: String, default: 'dark' }
})

const emit = defineEmits([])

const terminalRef = ref(null)
const isRunning = ref(false)

let terminal = null
let fitAddon = null
let searchAddon = null
let onDataDisposable = null
let onResizeDisposable = null
let searchResultsDisposable = null
let ptyId = ''
let currentLine = ''
let resizeObserver = null
let sshHost = null

/**
 * 初始化 xterm.js 终端实例
 */
const initTerminal = () => {
  terminal = new Terminal({
    allowProposedApi: true,
    customKeyEventHandler: (event) => {
      if ((event.ctrlKey || event.metaKey) && event.key === 'f') {
        return false
      }
      return true
    },
    fontFamily: 'Consolas, "Courier New", monospace',
    fontSize: 14,
    lineHeight: 1.2,
    theme: TERMINAL_THEMES[props.theme] || TERMINAL_THEMES.dark,
    cursorBlink: true,
    cursorStyle: 'block',
    scrollback: 5000,
    tabStopWidth: 4
  })

  fitAddon = new FitAddon()
  searchAddon = new SearchAddon()

  terminal.loadAddon(fitAddon)
  terminal.loadAddon(searchAddon)
  terminal.loadAddon(new WebLinksAddon())

  searchResultsDisposable = searchAddon.onDidChangeResults(({ resultIndex, resultCount }) => {
    const event = new CustomEvent('tab-search-result', {
      detail: {
        tabId: props.tabId,
        resultIndex: resultIndex >= 0 ? resultIndex + 1 : 0,
        resultCount
      }
    })
    window.dispatchEvent(event)
  })

  terminal.open(terminalRef.value)

  terminal.element.addEventListener('contextmenu', (e) => {
    e.preventDefault()
    e.stopPropagation()
    navigator.clipboard.readText().then(text => {
      if (text && isRunning.value && ptyId) {
        Write(ptyId, text).catch(() => {})
      }
    }).catch(() => {})
  })

  nextTick(() => {
    fitAddon.fit()
    startTerminal()
  })

  onDataDisposable = terminal.onData((data) => {
    if (isRunning.value) {
      if (data === '\r') {
        currentLine = ''
      } else if (data === '\x7f') {
        currentLine = currentLine.slice(0, -1)
      } else if (data.length === 1 && data.charCodeAt(0) >= 32) {
        currentLine += data
      }
      Write(ptyId, data).catch(() => {})
    }
  })

  onResizeDisposable = terminal.onResize(({ cols, rows }) => {
    if (isRunning.value && ptyId) {
      Resize(ptyId, cols, rows).catch(() => {})
    }
  })

  resizeObserver = new ResizeObserver(() => {
    if (fitAddon) {
      fitAddon.fit()
    }
  })
  resizeObserver.observe(terminalRef.value)

  window.addEventListener('resize', handleResize)
}

/**
 * 处理窗口大小变化
 */
const handleResize = () => {
  if (fitAddon) {
    fitAddon.fit()
  }
}

/**
 * 启动伪终端进程
 */
const startTerminal = async () => {
  try {
    fitAddon.fit()
    const cols = terminal.cols
    const rows = terminal.rows
    const id = await Start({ shell: String(props.shell), cols: Number(cols), rows: Number(rows) })
    ptyId = id
    isRunning.value = true
    EventsOn('pty-output-' + ptyId, handlePtyOutput)
    EventsOn('pty-exit-' + ptyId, handlePtyExit)
  } catch (err) {
    terminal.writeln('\r\n\x1b[31m' + err + '\x1b[0m')
  }
}

/**
 * 处理后端 PTY 输出数据
 * 同时检测 SSH 连接/断开状态以更新标签页标题
 */
const handlePtyOutput = (data) => {
  if (terminal) {
    terminal.write(data)
  }
  detectSshState(data)
}

/**
 * 检测 SSH 连接/断开状态
 * 通过终端输出中的特征字符串判断
 */
const detectSshState = (data) => {
  if (sshHost) {
    const disconnectPatterns = [
      'Connection to ',
      'connection closed',
      'Disconnected from',
      'Network error: Connection',
      'Connection timed out',
      'Connection refused',
      'No route to host',
      'broken pipe'
    ]
    const lowerData = data.toLowerCase()
    for (const pattern of disconnectPatterns) {
      if (lowerData.includes(pattern.toLowerCase())) {
        sshHost = null
        emitTabTitleChange(null)
        return
      }
    }
  } else if (sshHost === null) {
    const connectIndicators = [
      'Last login',
      'Welcome to'
    ]
    for (const indicator of connectIndicators) {
      if (data.includes(indicator)) {
        emitTabTitleChange('connected')
        return
      }
    }
  }
}

/**
 * 接收外部指定的 SSH 主机地址
 * 由 ServerListPanel 在发送 SSH 命令时触发
 */
const handleSshConnect = (event) => {
  if (event.detail.tabId === props.tabId) {
    sshHost = event.detail.host
    emitTabTitleChange(sshHost)
  }
}

/**
 * 通知 App.vue 更新标签页标题
 * host 为 null 时恢复默认标题
 */
const emitTabTitleChange = (host) => {
  const event = new CustomEvent('tab-title-change', {
    detail: { tabId: props.tabId, host }
  })
  window.dispatchEvent(event)
}

/**
 * 处理 PTY 进程退出事件
 */
const handlePtyExit = () => {
  isRunning.value = false
  ptyId = ''
}

/**
 * 监听外部发送命令事件
 */
const handleSendCommandEvent = (event) => {
  if (event.detail.tabId === props.tabId && isRunning.value && ptyId) {
    Write(ptyId, event.detail.command + '\r').catch(() => {})
  }
}

/**
 * 监听搜索事件（从SearchBar发出）
 */
const handleSearchEvent = (event) => {
  if (event.detail.tabId !== props.tabId || !searchAddon) return
  const { action, keyword } = event.detail
  switch (action) {
    case 'search':
      if (keyword) {
        searchAddon.findNext(keyword, { decorations: SEARCH_DECORATIONS, incremental: true })
      } else {
        searchAddon.clearDecorations()
      }
      break
    case 'findNext':
      if (keyword) {
        searchAddon.findNext(keyword, { decorations: SEARCH_DECORATIONS })
      }
      break
    case 'findPrev':
      if (keyword) {
        searchAddon.findPrevious(keyword, { decorations: SEARCH_DECORATIONS })
      }
      break
  }
}

/**
 * 监听搜索关闭事件
 */
const handleSearchCloseEvent = (event) => {
  if (event.detail.tabId !== props.tabId || !searchAddon) return
  searchAddon.clearDecorations()
}

onMounted(() => {
  initTerminal()
  window.addEventListener('terminal-send-command', handleSendCommandEvent)
  window.addEventListener('tab-ssh-connect', handleSshConnect)
  window.addEventListener('tab-search', handleSearchEvent)
  window.addEventListener('tab-search-close', handleSearchCloseEvent)
})

watch(() => props.theme, (newTheme) => {
  if (terminal) {
    terminal.options.theme = TERMINAL_THEMES[newTheme] || TERMINAL_THEMES.dark
  }
})

onUnmounted(() => {
  window.removeEventListener('terminal-send-command', handleSendCommandEvent)
  window.removeEventListener('tab-ssh-connect', handleSshConnect)
  window.removeEventListener('tab-search', handleSearchEvent)
  window.removeEventListener('tab-search-close', handleSearchCloseEvent)

  if (ptyId) {
    EventsOff('pty-output-' + ptyId)
    EventsOff('pty-exit-' + ptyId)
    Stop(ptyId).catch(() => {})
    ptyId = ''
    isRunning.value = false
  }

  window.removeEventListener('resize', handleResize)

  if (resizeObserver) {
    resizeObserver.disconnect()
  }

  if (onResizeDisposable) {
    onResizeDisposable.dispose()
  }

  if (onDataDisposable) {
    onDataDisposable.dispose()
  }

  if (searchResultsDisposable) {
    searchResultsDisposable.dispose()
  }

  if (terminal) {
    terminal.dispose()
  }
})

defineExpose({})
</script>

<style scoped>
.terminal-tab-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background-color: var(--bg-primary);
}

.terminal-content {
  flex: 1;
  padding: 4px;
  overflow: hidden;
}

.terminal-content :deep(.xterm) {
  height: 100%;
}

.terminal-content :deep(.xterm-viewport) {
  overflow-y: auto !important;
}
</style>
