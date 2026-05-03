<template>
  <div class="terminal-container">
    <div class="terminal-header">
      <span class="terminal-title">
        <el-icon><Monitor /></el-icon>
        终端
      </span>
      <div class="terminal-actions">
        <el-button-group>
          <el-button size="small" @click="startTerminal" :disabled="isRunning">
            <el-icon><VideoPlay /></el-icon>
            启动
          </el-button>
          <el-button size="small" @click="stopTerminal" :disabled="!isRunning">
            <el-icon><VideoPause /></el-icon>
            停止
          </el-button>
          <el-button size="small" @click="clearTerminal">
            <el-icon><Delete /></el-icon>
            清空
          </el-button>
        </el-button-group>
      </div>
    </div>
    <div ref="terminalRef" class="terminal-content"></div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import { WebLinksAddon } from 'xterm-addon-web-links'
import 'xterm/css/xterm.css'
import { Monitor, VideoPlay, VideoPause, Delete } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { Start, Write, Stop, IsRunning, Resize } from '../../wailsjs/go/main/PtyService'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

const terminalRef = ref(null)
const isRunning = ref(false)

let terminal = null
let fitAddon = null
let onDataDisposable = null
let onResizeDisposable = null

/**
 * 初始化 xterm.js 终端实例
 */
const initTerminal = () => {
  terminal = new Terminal({
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
  terminal.loadAddon(fitAddon)
  terminal.loadAddon(new WebLinksAddon())

  terminal.open(terminalRef.value)

  setTimeout(() => {
    fitAddon.fit()
  }, 100)

  onDataDisposable = terminal.onData((data) => {
    if (isRunning.value) {
      Write(data).catch((err) => {
        terminal.writeln('\r\n\x1b[31m写入失败: ' + err + '\x1b[0m')
      })
    }
  })

  onResizeDisposable = terminal.onResize(({ cols, rows }) => {
    if (isRunning.value) {
      Resize(cols, rows).catch(() => {})
    }
  })

  window.addEventListener('resize', handleResize)
}

/**
 * 处理窗口大小变化，自适应终端尺寸
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
    const { cols, rows } = terminal
    await Start('cmd.exe', cols, rows)
    isRunning.value = true
    ElMessage.success('终端已启动')
  } catch (err) {
    ElMessage.error('启动终端失败: ' + err)
    terminal.writeln('\r\n\x1b[31m启动失败: ' + err + '\x1b[0m')
  }
}

/**
 * 停止伪终端进程
 */
const stopTerminal = async () => {
  try {
    await Stop()
    isRunning.value = false
    terminal.writeln('\r\n\x1b[33m终端已停止\x1b[0m')
    ElMessage.warning('终端已停止')
  } catch (err) {
    ElMessage.error('停止终端失败: ' + err)
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
 * 处理后端 PTY 输出数据，写入终端
 */
const handlePtyOutput = (data) => {
  if (terminal) {
    terminal.write(data)
  }
}

onMounted(() => {
  initTerminal()

  EventsOn('pty-output', handlePtyOutput)

  IsRunning().then((running) => {
    isRunning.value = running
  })
})

onUnmounted(() => {
  EventsOff('pty-output')

  if (isRunning.value) {
    Stop()
  }

  window.removeEventListener('resize', handleResize)

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
</script>

<style scoped>
.terminal-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background-color: #1e1e1e;
  border-radius: 8px;
  overflow: hidden;
}

.terminal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 15px;
  background-color: #2d2d2d;
  border-bottom: 1px solid #3d3d3d;
}

.terminal-title {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #e5e5e5;
  font-size: 14px;
  font-weight: 500;
}

.terminal-content {
  flex: 1;
  padding: 10px;
  overflow: hidden;
}

.terminal-content :deep(.xterm) {
  height: 100%;
}

.terminal-content :deep(.xterm-viewport) {
  overflow-y: auto !important;
}
</style>
