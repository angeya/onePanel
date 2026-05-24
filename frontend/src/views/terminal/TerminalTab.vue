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
import { EventsOn, EventsOff, ClipboardSetText, ClipboardGetText } from '../../../wailsjs/runtime/runtime'
import {
  TERMINAL_THEMES,
  SEARCH_DECORATIONS,
  SSH_CONNECT_INDICATORS,
  SSH_DISCONNECT_PATTERNS
} from './terminalConstants'
import { detectTerminalSshState } from './terminalHelpers'
import { applyKeywordHighlight } from './terminalHighlight'

const props = defineProps({
  tabId: { type: String, required: true },
  shell: { type: String, default: 'cmd.exe' },
  theme: { type: String, default: 'dark' }
})

const terminalRef = ref(null)
const isRunning = ref(false)

let terminal = null
let fitAddon = null
let searchAddon = null
let onDataDisposable = null
let onResizeDisposable = null
let searchResultsDisposable = null
let resizeObserver = null
let ptyId = ''
let sshHost = null

/**
 * emitTabTitleChange 通知根组件更新终端标签标题。
 * SSH 登录成功后展示主机名，连接结束后恢复默认标题。
 */
const emitTabTitleChange = (host) => {
  window.dispatchEvent(new CustomEvent('tab-title-change', {
    detail: { tabId: props.tabId, host }
  }))
}

/**
 * fitTerminal 统一处理终端窗口尺寸自适应。
 * 用于窗口大小变化、容器变化和首次初始化后的尺寸修正。
 */
const fitTerminal = () => {
  if (fitAddon) {
    fitAddon.fit()
  }
}

/**
 * handleResize 响应浏览器窗口大小变化。
 */
const handleResize = () => {
  fitTerminal()
}

/**
 * handleSearchResultsChange 将 xterm 搜索结果同步给顶部 SearchBar。
 */
const handleSearchResultsChange = ({ resultIndex, resultCount }) => {
  window.dispatchEvent(new CustomEvent('tab-search-result', {
    detail: {
      tabId: props.tabId,
      resultIndex: resultIndex >= 0 ? resultIndex + 1 : 0,
      resultCount
    }
  }))
}

/**
 * copySelection 将终端选中文本复制到系统剪贴板。
 * 优先使用 Wails 运行时提供的操作系统级剪贴板接口，
 * 若不可用则降级到浏览器 navigator.clipboard API，
 * 最后兜底使用 document.execCommand 方式。
 */
const copySelection = async (text) => {
  if (!text) {
    return
  }
  try {
    await ClipboardSetText(text)
    return
  } catch {
    // Wails 剪贴板接口不可用时降级
  }
  try {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      await navigator.clipboard.writeText(text)
      return
    }
  } catch {
    // 浏览器 Clipboard API 不可用时降级
  }
  try {
    const textarea = document.createElement('textarea')
    textarea.value = text
    textarea.style.position = 'fixed'
    textarea.style.opacity = '0'
    textarea.style.left = '-9999px'
    document.body.appendChild(textarea)
    textarea.select()
    document.execCommand('copy')
    document.body.removeChild(textarea)
  } catch {
    // 所有方案均失败，静默处理
  }
}

/**
 * pasteFromClipboard 从系统剪贴板读取文本并写入终端。
 * 优先使用 Wails 运行时接口，不可用时降级到浏览器 API。
 */
const pasteFromClipboard = async () => {
  let text = ''
  try {
    text = await ClipboardGetText()
  } catch {
    try {
      if (navigator.clipboard && navigator.clipboard.readText) {
        text = await navigator.clipboard.readText()
      }
    } catch {
      return
    }
  }
  if (text && isRunning.value && ptyId) {
    Write(ptyId, text).catch(() => {})
  }
}

/**
 * handleTerminalContextMenu 支持右键快速粘贴剪贴板内容。
 * 为了避免与宿主右键菜单冲突，这里直接接管 contextmenu 行为。
 */
const handleTerminalContextMenu = (event) => {
  event.preventDefault()
  event.stopPropagation()
  pasteFromClipboard()
}

/**
 * handleTerminalInput 将用户输入写入 PTY 进程。
 */
const handleTerminalInput = (data) => {
  if (!isRunning.value || !ptyId) {
    return
  }
  Write(ptyId, data).catch(() => {})
}

/**
 * handleTerminalResize 将 xterm 当前行列数同步回后端 PTY。
 */
const handleTerminalResize = ({ cols, rows }) => {
  if (isRunning.value && ptyId) {
    Resize(ptyId, cols, rows).catch(() => {})
  }
}

/**
 * detectSshState 根据终端输出推断 SSH 连接状态。
 * 状态判断逻辑已提取到独立辅助函数，便于单独维护。
 */
const detectSshState = (data) => {
  const sshState = detectTerminalSshState({
    sshHost,
    data,
    disconnectPatterns: SSH_DISCONNECT_PATTERNS,
    connectIndicators: SSH_CONNECT_INDICATORS
  })

  if (sshState === 'disconnected') {
    sshHost = null
    emitTabTitleChange(null)
    return
  }

  if (sshState === 'connected') {
    emitTabTitleChange('connected')
  }
}

/**
 * handlePtyOutput 将后端 PTY 输出写入终端，并顺带检测 SSH 状态变化。
 * 注意：SSH 状态检测必须使用原始数据，高亮处理仅影响终端显示。
 */
const handlePtyOutput = (data) => {
  detectSshState(data)
  if (terminal) {
    terminal.write(applyKeywordHighlight(data))
  }
}

/**
 * handlePtyExit 处理后端 PTY 退出事件。
 */
const handlePtyExit = () => {
  isRunning.value = false
  ptyId = ''
}

/**
 * startTerminal 启动后端 PTY 并绑定输出事件。
 */
const startTerminal = async () => {
  try {
    fitTerminal()
    const id = await Start({
      shell: String(props.shell),
      cols: Number(terminal.cols),
      rows: Number(terminal.rows)
    })
    ptyId = id
    isRunning.value = true
    EventsOn('pty-output-' + ptyId, handlePtyOutput)
    EventsOn('pty-exit-' + ptyId, handlePtyExit)
    window.dispatchEvent(new CustomEvent('terminal-ready', {
      detail: { tabId: props.tabId }
    }))
  } catch (err) {
    terminal?.writeln('\r\n\x1b[31m' + err + '\x1b[0m')
  }
}

/**
 * handleCustomKeyEvent 自定义键盘事件处理。
 * 返回 false 表示拦截该事件（不传递给终端），true 表示放行。
 * 规则：
 * - Ctrl+F：拦截，由搜索组件处理
 * - Ctrl+C 有选区：复制选中内容到剪贴板，拦截
 * - Ctrl+C 无选区：放行，发送 SIGINT 中断信号
 * - Ctrl+Shift+C：强制复制选中内容
 * - Ctrl+V：粘贴剪贴板内容到终端
 * - Ctrl+Shift+V：同 Ctrl+V，粘贴剪贴板内容
 */
const handleCustomKeyEvent = (event) => {
  if (event.ctrlKey && !event.shiftKey && !event.altKey && event.key === 'f') {
    return false
  }

  if (event.ctrlKey && !event.shiftKey && !event.altKey && event.key === 'c') {
    if (terminal.hasSelection()) {
      copySelection(terminal.getSelection())
      return false
    }
    return true
  }

  if (event.ctrlKey && event.shiftKey && !event.altKey && event.key === 'C') {
    if (terminal.hasSelection()) {
      copySelection(terminal.getSelection())
    }
    return false
  }

  if (event.ctrlKey && !event.shiftKey && !event.altKey && event.key === 'v') {
    pasteFromClipboard()
    return false
  }

  if (event.ctrlKey && event.shiftKey && !event.altKey && event.key === 'V') {
    pasteFromClipboard()
    return false
  }

  return true
}

/**
 * handleTerminalKeyDown 终端 DOM 级别的键盘事件拦截。
 * 在捕获阶段优先于 xterm 内部处理，确保 Ctrl+C/V 复制粘贴可靠生效。
 * 仅拦截需要自定义处理的快捷键，其余事件正常传递给 xterm。
 */
const handleTerminalKeyDown = (event) => {
  if (!terminal) {
    return
  }

  if (event.ctrlKey && !event.shiftKey && !event.altKey && event.key === 'c') {
    if (terminal.hasSelection()) {
      event.preventDefault()
      event.stopPropagation()
      copySelection(terminal.getSelection())
    }
    return
  }

  if (event.ctrlKey && event.shiftKey && !event.altKey && event.key === 'C') {
    if (terminal.hasSelection()) {
      event.preventDefault()
      event.stopPropagation()
      copySelection(terminal.getSelection())
    }
    return
  }

  if (event.ctrlKey && !event.shiftKey && !event.altKey && event.key === 'v') {
    event.preventDefault()
    event.stopPropagation()
    pasteFromClipboard()
    return
  }

  if (event.ctrlKey && event.shiftKey && !event.altKey && event.key === 'V') {
    event.preventDefault()
    event.stopPropagation()
    pasteFromClipboard()
    return
  }
}

/**
 * initializeTerminal 创建并初始化 xterm 实例。
 * 该方法集中处理主题、插件、事件和首次启动流程。
 */
const initializeTerminal = () => {
  terminal = new Terminal({
    allowProposedApi: true,
    customKeyEventHandler: handleCustomKeyEvent,
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

  searchResultsDisposable = searchAddon.onDidChangeResults(handleSearchResultsChange)

  terminal.open(terminalRef.value)
  terminal.element?.addEventListener('contextmenu', handleTerminalContextMenu)
  terminal.element?.addEventListener('keydown', handleTerminalKeyDown, true)

  onDataDisposable = terminal.onData(handleTerminalInput)
  onResizeDisposable = terminal.onResize(handleTerminalResize)

  resizeObserver = new ResizeObserver(() => {
    fitTerminal()
  })
  resizeObserver.observe(terminalRef.value)
  window.addEventListener('resize', handleResize)

  nextTick(() => {
    fitTerminal()
    startTerminal()
  })
}

/**
 * handleSshConnect 接收外部指定的 SSH 主机信息。
 * 由服务器列表触发，便于终端在后续输出中更新标签标题。
 */
const handleSshConnect = (event) => {
  if (event.detail.tabId === props.tabId) {
    sshHost = event.detail.host
    emitTabTitleChange(sshHost)
  }
}

/**
 * handleSendCommandEvent 监听外部发送命令事件。
 */
const handleSendCommandEvent = (event) => {
  if (event.detail.tabId === props.tabId && isRunning.value && ptyId) {
    Write(ptyId, event.detail.command + '\r').catch(() => {})
  }
}

/**
 * handleSearchEvent 响应终端页签的搜索、上一个和下一个动作。
 */
const handleSearchEvent = (event) => {
  if (event.detail.tabId !== props.tabId || !searchAddon) return

  const { action, keyword } = event.detail
  switch (action) {
    case 'search':
      if (keyword) {
        searchAddon.findNext(keyword, {
          decorations: SEARCH_DECORATIONS,
          incremental: true
        })
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
 * handleSearchCloseEvent 在搜索条关闭时清理终端搜索高亮。
 */
const handleSearchCloseEvent = (event) => {
  if (event.detail.tabId === props.tabId && searchAddon) {
    searchAddon.clearDecorations()
  }
}

/**
 * bindWindowEvents 统一绑定终端页签依赖的全局事件。
 */
const bindWindowEvents = () => {
  window.addEventListener('terminal-send-command', handleSendCommandEvent)
  window.addEventListener('tab-ssh-connect', handleSshConnect)
  window.addEventListener('tab-search', handleSearchEvent)
  window.addEventListener('tab-search-close', handleSearchCloseEvent)
}

/**
 * unbindWindowEvents 统一解绑终端页签依赖的全局事件。
 */
const unbindWindowEvents = () => {
  window.removeEventListener('terminal-send-command', handleSendCommandEvent)
  window.removeEventListener('tab-ssh-connect', handleSshConnect)
  window.removeEventListener('tab-search', handleSearchEvent)
  window.removeEventListener('tab-search-close', handleSearchCloseEvent)
  window.removeEventListener('resize', handleResize)
}

/**
 * disposeTerminalResources 释放终端实例、观察器和 PTY 相关资源。
 */
const disposeTerminalResources = () => {
  if (ptyId) {
    EventsOff('pty-output-' + ptyId)
    EventsOff('pty-exit-' + ptyId)
    Stop(ptyId).catch(() => {})
    ptyId = ''
    isRunning.value = false
  }

  if (resizeObserver) {
    resizeObserver.disconnect()
    resizeObserver = null
  }

  if (onResizeDisposable) {
    onResizeDisposable.dispose()
    onResizeDisposable = null
  }

  if (onDataDisposable) {
    onDataDisposable.dispose()
    onDataDisposable = null
  }

  if (searchResultsDisposable) {
    searchResultsDisposable.dispose()
    searchResultsDisposable = null
  }

  if (terminal) {
    terminal.element?.removeEventListener('keydown', handleTerminalKeyDown, true)
    terminal.element?.removeEventListener('contextmenu', handleTerminalContextMenu)
    terminal.dispose()
    terminal = null
  }

  fitAddon = null
  searchAddon = null
}

onMounted(() => {
  initializeTerminal()
  bindWindowEvents()
})

/**
 * 主题变化时仅更新 xterm 配色，不重建终端实例。
 */
watch(() => props.theme, (newTheme) => {
  if (terminal) {
    terminal.options.theme = TERMINAL_THEMES[newTheme] || TERMINAL_THEMES.dark
  }
})

onUnmounted(() => {
  unbindWindowEvents()
  disposeTerminalResources()
})
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

.terminal-content :deep(.xterm-viewport::-webkit-scrollbar) {
  width: 6px;
}

.terminal-content :deep(.xterm-viewport::-webkit-scrollbar-track) {
  background: transparent;
}

.terminal-content :deep(.xterm-viewport::-webkit-scrollbar-thumb) {
  background-color: var(--scrollbar-thumb);
  border-radius: 3px;
}

.terminal-content :deep(.xterm-viewport::-webkit-scrollbar-thumb:hover) {
  background-color: var(--text-dimmed);
}
</style>
