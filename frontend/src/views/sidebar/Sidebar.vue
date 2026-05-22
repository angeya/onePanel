<template>
  <div class="left-panel">
    <div class="nav-column">
      <div class="nav-menu">
        <div
          v-for="item in navItems"
          :key="item.key"
          class="nav-item"
          :class="{ active: activeNav === item.key }"
          :title="item.label"
          @click="switchNav(item.key)"
        >
          <el-icon :size="20"><component :is="item.icon" /></el-icon>
        </div>
      </div>
      <div class="nav-bottom">
        <div class="nav-settings" title="系统设置" @click="openSettings">
          <el-icon :size="20"><Setting /></el-icon>
        </div>
        <div class="github-entry" title="打开 GitHub" @click="openGithub">
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <path d="M12 2C6.477 2 2 6.59 2 12.253c0 4.53 2.865 8.374 6.839 9.73.5.096.683-.222.683-.493 0-.244-.009-.89-.014-1.747-2.782.62-3.369-1.37-3.369-1.37-.455-1.183-1.11-1.498-1.11-1.498-.908-.636.069-.624.069-.624 1.004.072 1.532 1.056 1.532 1.056.893 1.566 2.341 1.114 2.91.852.091-.67.35-1.114.636-1.37-2.22-.26-4.555-1.14-4.555-5.074 0-1.12.39-2.036 1.029-2.754-.103-.261-.446-1.31.097-2.73 0 0 .84-.277 2.75 1.052A9.297 9.297 0 0 1 12 6.836a9.27 9.27 0 0 1 2.504.35c1.909-1.329 2.748-1.052 2.748-1.052.545 1.42.202 2.469.1 2.73.64.718 1.027 1.634 1.027 2.754 0 3.944-2.339 4.81-4.566 5.065.359.319.679.948.679 1.91 0 1.379-.012 2.49-.012 2.83 0 .273.18.593.688.492C19.138 20.623 22 16.781 22 12.253 22 6.59 17.523 2 12 2Z" />
          </svg>
        </div>
      </div>
    </div>

    <div
      class="sub-panel"
      :class="{ collapsed: panelCollapsed }"
      :style="{ width: panelWidth + 'px' }"
    >
      <div class="sub-panel-body">
        <TerminalPanel
          v-if="activeNav === 'terminal'"
        />
        <MyAppPanel v-if="activeNav === 'apps'" />
        <QuickLaunchPanel v-if="activeNav === 'shortcuts'" />
        <ToolsPanel v-if="activeNav === 'tools'" />
      </div>

      <div class="resize-handle"
        @mousedown.prevent="startResize"
      />
    </div>

    <div class="collapse-btn"
      :title="panelCollapsed ? '展开面板' : '收起面板'"
      @click="toggleCollapse"
    >
      <el-icon :size="12">
        <DArrowLeft v-if="!panelCollapsed" />
        <DArrowRight v-else />
      </el-icon>
    </div>
  </div>
</template>

<script setup>
import { ref, inject, onUnmounted } from 'vue'
import { Setting, DArrowLeft, DArrowRight, Monitor, Grid, Promotion, SetUp } from '@element-plus/icons-vue'
import { BrowserOpenURL } from '../../../wailsjs/runtime/runtime'
import TerminalPanel from './TerminalPanel.vue'
import MyAppPanel from './MyAppPanel.vue'
import QuickLaunchPanel from './QuickLaunchPanel.vue'
import ToolsPanel from './ToolsPanel.vue'

const DEFAULT_PANEL_WIDTH = 240
const MIN_PANEL_WIDTH = 120
const MAX_PANEL_WIDTH = DEFAULT_PANEL_WIDTH

const navItems = [
  { key: 'terminal', label: '终端', icon: Monitor },
  { key: 'apps', label: '我的应用', icon: Grid },
  { key: 'shortcuts', label: '快速启动', icon: Promotion },
  { key: 'tools', label: '实用工具', icon: SetUp }
]

const activeNav = inject('activeNav')
const switchNavFromParent = inject('switchNav')
const openSettings = inject('openSettings')

const openGithub = () => {
  BrowserOpenURL('https://github.com/angeya')
}

const panelWidth = ref(DEFAULT_PANEL_WIDTH)
const panelCollapsed = ref(false)
const savedWidthBeforeCollapse = ref(DEFAULT_PANEL_WIDTH)

let resizing = false
let startX = 0
let startWidth = 0


/**
 * 切换导航面板
 * @param {string} navKey - 导航项键名
 */
const switchNav = (navKey) => {
  switchNavFromParent(navKey)
  if (panelCollapsed.value) {
    toggleCollapse()
  }
}

/**
 * 开始调整面板宽度
 * @param {MouseEvent} e - 鼠标事件
 */
const startResize = (e) => {
  if (panelCollapsed.value) {
    return
  }
  resizing = true
  startX = e.clientX
  startWidth = panelWidth.value
  document.addEventListener('mousemove', onResize)
  document.addEventListener('mouseup', stopResize)
  document.body.style.cursor = 'col-resize'
  document.body.style.userSelect = 'none'
}

/**
 * 调整面板宽度
 * @param {MouseEvent} e - 鼠标事件
 */
const onResize = (e) => {
  if (!resizing) return
  const delta = startX - e.clientX
  let newWidth = startWidth - delta
  newWidth = Math.max(MIN_PANEL_WIDTH, Math.min(MAX_PANEL_WIDTH, newWidth))
  panelWidth.value = newWidth
}

/**
 * 停止调整面板宽度
 */
const stopResize = () => {
  resizing = false
  document.removeEventListener('mousemove', onResize)
  document.removeEventListener('mouseup', stopResize)
  document.body.style.cursor = ''
  document.body.style.userSelect = ''
}

/**
 * 切换面板折叠状态
 */
const toggleCollapse = () => {
  if (panelCollapsed.value) {
    panelWidth.value = savedWidthBeforeCollapse.value
    panelCollapsed.value = false
  } else {
    savedWidthBeforeCollapse.value = panelWidth.value
    panelCollapsed.value = true
    panelWidth.value = 0
  }
}

/**
 * 组件卸载时移除事件监听
 */
onUnmounted(() => {
  document.removeEventListener('mousemove', onResize)
  document.removeEventListener('mouseup', stopResize)
})
</script>

<style scoped>
.left-panel {
  display: flex;
  flex-shrink: 0;
  height: 100%;
  background-color: var(--bg-primary);
  position: relative;
}

.nav-column {
  width: 56px;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  border-right: 1px solid var(--border-color);
  flex-shrink: 0;
  position: relative;
}

.nav-logo {
  height: 50px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid var(--border-color);
  width: 100%;
}

.logo-icon {
  width: 28px;
  height: 28px;
  object-fit: contain;
}

.nav-menu {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 8px 0;
  padding-bottom: 70px;
  gap: 4px;
  min-height: 0;
  overflow: hidden;
}

.nav-item {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  cursor: pointer;
  color: var(--text-dimmed);
  transition: all 0.15s;
}

.nav-item:hover {
  color: var(--text-secondary);
  background-color: var(--bg-hover);
}

.nav-item.active {
  color: var(--accent);
  background-color: var(--bg-hover);
}

.nav-bottom {
  padding: 8px 0 12px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  background-color: var(--bg-primary);
}

.nav-settings {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  cursor: pointer;
  color: var(--text-muted);
  transition: all 0.15s;
}

.nav-settings:hover {
  color: var(--text-secondary);
  background-color: var(--bg-hover);
}

.github-entry {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  cursor: pointer;
  color: var(--text-dimmed);
  transition: all 0.15s;
}

.github-entry:hover {
  color: var(--text-primary);
  background-color: var(--bg-hover);
}

.github-entry svg {
  width: 18px;
  height: 18px;
  fill: currentColor;
}

.sub-panel {
  display: flex;
  overflow: hidden;
  border-right: 1px solid var(--border-color);
  transition: width 0.2s ease;
  flex-shrink: 0;
}

.sub-panel.collapsed {
  width: 0 !important;
  border-right: none;
}

.sub-panel-body {
  flex: 1;
  overflow: hidden;
  min-width: 0;
}

.resize-handle {
  width: 4px;
  cursor: col-resize;
  flex-shrink: 0;
  background: transparent;
  transition: background-color 0.15s;
  border-radius: 2px;
}

.resize-handle:hover {
  background-color: var(--accent);
}

.collapse-btn {
  width: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: var(--text-dimmed);
  background-color: var(--bg-primary);
  border-right: 1px solid var(--border-color);
  flex-shrink: 0;
  transition: all 0.15s;
}

.collapse-btn:hover {
  color: var(--text-primary);
  background-color: var(--bg-hover);
}
</style>
