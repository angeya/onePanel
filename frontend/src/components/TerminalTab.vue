<template>
  <div class="terminal-tab-container">
    <div class="terminal-toolbar">
      <div class="toolbar-left">
        <el-button size="small" @click="clearTerminal" plain>
          <el-icon><Delete /></el-icon>
          清空
        </el-button>
      </div>
      <div class="toolbar-right">
        <el-button size="small" @click="toggleSearch" plain>
          <el-icon><Search /></el-icon>
          搜索
        </el-button>
      </div>
    </div>
    <div v-if="showSearch" class="search-bar">
      <el-input
        v-model="searchKeyword"
        size="small"
        placeholder="搜索终端内容..."
        @input="handleSearch"
        clearable
      >
        <template #append>
          <el-button-group>
            <el-button size="small" @click="findPrevious">上一个</el-button>
            <el-button size="small" @click="findNext">下一个</el-button>
          </el-button-group>
        </template>
      </el-input>
      <el-button size="small" @click="toggleSearch" class="search-close" plain>
        <el-icon><Close /></el-icon>
      </el-button>
    </div>
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
import { Delete, Search, Close } from '@element-plus/icons-vue'
import { Start, Write, Stop, Resize } from '../../wailsjs/go/main/PtyService'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

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
const showSearch = ref(false)
const searchKeyword = ref('')

let terminal = null
let fitAddon = null
let searchAddon = null
let onDataDisposable = null
let onResizeDisposable = null
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
    terminal.writeln('\r\n\x1b[31m启动失败: ' + err + '\x1b[0m')
  }
}

/**
 * 清空终端显示内容
 */
const clearTerminal = () => {
  if (terminal) {
    terminal.clear()
  }
}

/**
 * 切换搜索栏显示
 */
const toggleSearch = () => {
  showSearch.value = !showSearch.value
  if (showSearch.value && searchKeyword.value) {
    handleSearch()
  } else if (!showSearch.value && searchAddon) {
    searchAddon.clearDecorations()
  }
}

/**
 * 执行搜索
 */
const handleSearch = () => {
  if (!searchAddon) return
  if (searchKeyword.value) {
    searchAddon.findNext(searchKeyword.value, { decorations: SEARCH_DECORATIONS })
  } else {
    searchAddon.clearDecorations()
  }
}

/**
 * 查找上一个匹配项
 */
const findPrevious = () => {
  if (searchAddon && searchKeyword.value) {
    searchAddon.findPrevious(searchKeyword.value, { decorations: SEARCH_DECORATIONS })
  }
}

/**
 * 查找下一个匹配项
 */
const findNext = () => {
  if (searchAddon && searchKeyword.value) {
    searchAddon.findNext(searchKeyword.value, { decorations: SEARCH_DECORATIONS })
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

onMounted(() => {
  initTerminal()

  window.addEventListener('terminal-send-command', handleSendCommandEvent)
})

onUnmounted(() => {
  window.removeEventListener('terminal-send-command', handleSendCommandEvent)

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
  background-color: #1e1e1e;
}

.terminal-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 12px;
  background-color: #2d2d2d;
  border-bottom: 1px solid #3d3d3d;
}

.toolbar-left,
.toolbar-right {
  display: flex;
  gap: 6px;
}

.search-bar {
  display: flex;
  align-items: center;
  padding: 6px 12px;
  background-color: #2d2d2d;
  border-bottom: 1px solid #3d3d3d;
  gap: 8px;
}

.search-bar .el-input {
  flex: 1;
}

.search-close {
  flex-shrink: 0;
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
