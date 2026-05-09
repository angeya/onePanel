<template>
  <div class="terminal-tab-container">
    <div ref="terminalRef" class="terminal-content"></div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
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

const props = defineProps({
  tabId: { type: String, required: true },
  shell: { type: String, default: 'cmd.exe' }
})

const emit = defineEmits(['commandExecuted', 'sendCommand'])

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

/**
 * 初始化 xterm.js 终端实例
 */
const initTerminal = () => {
  terminal = new Terminal({
    allowProposedApi: true,
    fontFamily: 'Consolas, "Courier New", monospace',
    fontSize: 14,
    lineHeight: 1.2,
    theme: {
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
    cursorBlink: true,
    cursorStyle: 'block',
    scrollback: 10000,
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

  nextTick(() => {
    fitAddon.fit()
    startTerminal()
  })

  onDataDisposable = terminal.onData((data) => {
    if (isRunning.value) {
      if (data === '\r') {
        if (currentLine.trim()) {
          emit('commandExecuted', { command: currentLine.trim() })
        }
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
 */
const handlePtyOutput = (data) => {
  if (terminal) {
    terminal.write(data)
  }
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
  window.addEventListener('tab-search', handleSearchEvent)
  window.addEventListener('tab-search-close', handleSearchCloseEvent)
})

onUnmounted(() => {
  window.removeEventListener('terminal-send-command', handleSendCommandEvent)
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
