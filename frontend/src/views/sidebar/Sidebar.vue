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
          <el-icon :size="18"><Setting /></el-icon>
        </div>
        <div class="version-info" title="oneWin v0.0.1">v0.0.1</div>
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
  background-color: var(--bg-primary);
}

.nav-column {
  width: 56px;
  display: flex;
  flex-direction: column;
  align-items: center;
  border-right: 1px solid var(--border-color);
  flex-shrink: 0;
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
  gap: 4px;
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
}

.nav-settings {
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

.nav-settings:hover {
  color: var(--text-secondary);
  background-color: var(--bg-hover);
}

.version-info {
  font-size: 10px;
  color: var(--text-dimmed);
  text-align: center;
  cursor: default;
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
