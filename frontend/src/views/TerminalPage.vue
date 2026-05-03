<template>
  <div class="terminal-page">
    <div class="terminal-tabs-area">
      <div class="tabs-header">
        <div class="tabs-list">
          <div
            v-for="tab in tabs"
            :key="tab.id"
            class="tab-item"
            :class="{ active: activeTabId === tab.id }"
            @click="switchTab(tab.id)"
          >
            <el-icon size="12"><Monitor /></el-icon>
            <span class="tab-name">{{ tab.name }}</span>
            <el-icon
              class="tab-close"
              size="12"
              @click.stop="closeTab(tab.id)"
            >
              <Close />
            </el-icon>
          </div>
        </div>
        <el-button class="tab-add" size="small" @click="addTab" :icon="Plus" circle />
      </div>
    </div>
    <div class="terminal-body">
      <div class="terminal-main">
        <TerminalTab
          v-for="tab in tabs"
          :key="tab.id"
          :tab-id="tab.id"
          :shell="tab.shell"
          v-show="activeTabId === tab.id"
          @command-executed="handleCommandExecuted"
          @send-command="handleSendCommand"
        />
      </div>
      <div class="terminal-sidebar" :class="{ collapsed: sidebarCollapsed }">
        <div class="sidebar-toggle" @click="sidebarCollapsed = !sidebarCollapsed">
          <el-icon v-if="sidebarCollapsed"><DArrowLeft /></el-icon>
          <el-icon v-else><DArrowRight /></el-icon>
        </div>
        <div class="sidebar-content" v-show="!sidebarCollapsed">
          <el-tabs v-model="sidebarTab" class="sidebar-tabs">
            <el-tab-pane label="快捷命令" name="shortcuts">
              <ShortcutPanel @execute-command="handleExecuteShortcut" />
            </el-tab-pane>
            <el-tab-pane label="历史" name="history">
              <HistoryPanel @execute-command="handleExecuteFromHistory" />
            </el-tab-pane>
          </el-tabs>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, nextTick } from 'vue'
import { Monitor, Close, Plus, DArrowLeft, DArrowRight } from '@element-plus/icons-vue'
import TerminalTab from '../components/TerminalTab.vue'
import ShortcutPanel from '../components/ShortcutPanel.vue'
import HistoryPanel from '../components/HistoryPanel.vue'

const tabs = ref([])
const activeTabId = ref('')
const sidebarCollapsed = ref(false)
const sidebarTab = ref('shortcuts')
let tabCounter = 0

/**
 * 新增终端标签页
 */
const addTab = (shell = 'cmd.exe') => {
  tabCounter++
  const id = `tab-${Date.now()}-${tabCounter}`
  const tab = {
    id,
    name: `终端 ${tabCounter}`,
    shell
  }
  tabs.value.push(tab)
  activeTabId.value = id
}

/**
 * 切换终端标签页
 */
const switchTab = (id) => {
  activeTabId.value = id
}

/**
 * 关闭终端标签页
 */
const closeTab = (id) => {
  const index = tabs.value.findIndex((t) => t.id === id)
  if (index === -1) return

  tabs.value.splice(index, 1)

  if (tabs.value.length === 0) {
    addTab()
  } else if (activeTabId.value === id) {
    const newIndex = Math.min(index, tabs.value.length - 1)
    activeTabId.value = tabs.value[newIndex].id
  }
}

/**
 * 处理命令执行完成事件，记录历史
 */
const handleCommandExecuted = (data) => {
  if (data && data.command) {
    AddHistory(data.command, 'cmd.exe', '').catch(() => {})
  }
}

/**
 * 处理快捷命令执行
 */
const handleExecuteShortcut = (command) => {
  handleSendCommand(activeTabId.value, command)
}

/**
 * 处理从历史记录执行命令
 */
const handleExecuteFromHistory = (command) => {
  handleSendCommand(activeTabId.value, command)
}

/**
 * 向指定终端发送命令
 */
const handleSendCommand = (tabId, command) => {
  const event = new CustomEvent('terminal-send-command', {
    detail: { tabId, command }
  })
  window.dispatchEvent(event)
}

import { AddHistory } from '../../wailsjs/go/main/HistoryService'

addTab()
</script>

<style scoped>
.terminal-page {
  display: flex;
  flex-direction: column;
  height: 100%;
  background-color: #252526;
}

.terminal-tabs-area {
  background-color: #1e1e1e;
  border-bottom: 1px solid #2d2d2d;
}

.tabs-header {
  display: flex;
  align-items: center;
  padding: 4px 8px;
  gap: 4px;
}

.tabs-list {
  display: flex;
  gap: 2px;
  flex: 1;
  overflow-x: auto;
}

.tabs-list::-webkit-scrollbar {
  height: 3px;
}

.tabs-list::-webkit-scrollbar-thumb {
  background-color: #555;
  border-radius: 2px;
}

.tab-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: 4px;
  cursor: pointer;
  white-space: nowrap;
  font-size: 13px;
  color: #a0a0a0;
  background-color: #2d2d2d;
  transition: all 0.15s;
  min-width: 0;
}

.tab-item:hover {
  background-color: #3d3d3d;
  color: #e5e5e5;
}

.tab-item.active {
  background-color: #3d3d3d;
  color: #e5e5e5;
  border-bottom: 2px solid #409eff;
}

.tab-name {
  overflow: hidden;
  text-overflow: ellipsis;
}

.tab-close {
  border-radius: 50%;
  padding: 2px;
  opacity: 0;
  transition: opacity 0.15s;
}

.tab-item:hover .tab-close {
  opacity: 1;
}

.tab-close:hover {
  background-color: #555;
  color: #fff;
}

.tab-add {
  flex-shrink: 0;
  background-color: #2d2d2d !important;
  border-color: #3d3d3d !important;
  color: #a0a0a0 !important;
}

.tab-add:hover {
  color: #409eff !important;
}

.terminal-body {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.terminal-main {
  flex: 1;
  overflow: hidden;
}

.terminal-sidebar {
  width: 320px;
  background-color: #1e1e1e;
  border-left: 1px solid #2d2d2d;
  display: flex;
  position: relative;
  transition: width 0.2s;
  overflow: hidden;
}

.terminal-sidebar.collapsed {
  width: 24px;
}

.sidebar-toggle {
  position: absolute;
  top: 8px;
  left: 4px;
  cursor: pointer;
  color: #a0a0a0;
  z-index: 10;
  padding: 2px;
  border-radius: 4px;
}

.sidebar-toggle:hover {
  color: #e5e5e5;
  background-color: #3d3d3d;
}

.sidebar-content {
  flex: 1;
  overflow: hidden;
  margin-left: 24px;
}

.sidebar-tabs :deep(.el-tabs__header) {
  margin: 0;
}

.sidebar-tabs :deep(.el-tabs__nav-wrap::after) {
  background-color: #2d2d2d;
}

.sidebar-tabs :deep(.el-tabs__item) {
  color: #a0a0a0;
  font-size: 13px;
}

.sidebar-tabs :deep(.el-tabs__item.is-active) {
  color: #409eff;
}

.sidebar-tabs :deep(.el-tabs__content) {
  padding: 8px;
  overflow-y: auto;
  height: calc(100% - 40px);
}
</style>
