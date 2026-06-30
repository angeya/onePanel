<template>
  <div class="app-shell">
    <div v-if="!appReady" class="app-boot-screen">
      <div class="boot-card">
        <div class="boot-title">oneWin</div>
        <div class="boot-subtitle">{{ bootStatus }}</div>
        <div class="boot-progress-track">
          <div class="boot-progress-bar" :style="{ width: bootProgress + '%' }"></div>
        </div>
        <div class="boot-progress-text">{{ bootProgress }}%</div>
      </div>
    </div>

    <div v-show="appReady" class="app-container">
      <AppTitleBar />
      <div class="app-workspace">
        <Sidebar />

        <div class="right-panel">
          <div v-if="tabs.length === 0" class="empty-main">
            <el-empty description="选择左侧功能开始使用" :image-size="80" />
          </div>
          <template v-else>
            <div class="main-tabs-header">
              <div class="tabs-list">
                <div
                  v-for="tab in tabs"
                  :key="tab.id"
                  class="main-tab-item"
                  :class="{ active: activeTabId === tab.id }"
                  @click="switchTab(tab.id)"
                  @mouseup="handleTabMouseDown($event, tab)"
                >
                  <el-icon size="12"><component :is="getTabIcon(tab)" /></el-icon>
                  <span class="tab-name">{{ tab.title }}</span>
                  <el-icon
                    v-if="tab.closable !== false"
                    class="tab-close"
                    size="12"
                    @click.stop="closeTab(tab.id)"
                  >
                    <Close />
                  </el-icon>
                </div>
              </div>
            </div>
            <div class="main-tabs-body" ref="mainTabsBodyRef">
              <TerminalTab
                v-for="tab in terminalTabs"
                :key="tab.id"
                :tab-id="tab.id"
                :shell="tab.shell || 'cmd.exe'"
                :theme="currentTheme"
                v-show="activeTabId === tab.id"
              />
              <div
                v-for="tab in appTabs"
                :key="tab.id"
                v-show="activeTabId === tab.id"
                class="app-iframe-wrapper"
                :data-tab-id="tab.id"
              >
                <iframe
                  :src="tab.url"
                  class="app-iframe"
                  sandbox="allow-scripts allow-same-origin allow-forms allow-popups"
                />
              </div>
              <QuickLaunchTab
                v-if="quickLaunchTab"
                v-show="activeTabId === quickLaunchTab.id"
                ref="quickLaunchTabRef"
                :data-tab-id="quickLaunchTab.id"
              />
              <NetworkPortList
                v-for="tab in toolTabs"
                :key="tab.id"
                v-show="activeTabId === tab.id"
                :data-tab-id="tab.id"
                :embedded="true"
              />
              <SearchBar
                ref="searchBarRef"
                :visible="currentSearchVisible"
                @update:visible="handleSearchBarVisibleChange"
                @search="handleSearchInput"
                @find-next="handleSearchFindNext"
                @find-prev="handleSearchFindPrev"
                @close="handleSearchClose"
              />
            </div>
          </template>
        </div>
      </div>

      <MyApp ref="myAppDialogsRef" />
      <QuickLaunchDialogs ref="quickLaunchDialogsRef" />

      <SettingsDialog
        ref="settingsRef"
        @theme-change="changeTheme"
        @shell-change="changeDefaultShell"
        @close-action-change="changeCloseAction"
        @hotkey-change="changeGlobalHotkey"
      />
      <CloseActionDialog ref="closeActionDialogRef" />
    </div>
  </div>
</template>

<script setup>
import { ref, defineAsyncComponent, onMounted, onUnmounted } from 'vue'
import { Close } from '@element-plus/icons-vue'
import TerminalTab from './views/terminal/TerminalTab.vue'
import QuickLaunchTab from './views/quicklaunch/QuickLaunchTab.vue'
import Sidebar from './views/sidebar/Sidebar.vue'
import SearchBar from './components/SearchBar.vue'
import AppTitleBar from './components/AppTitleBar.vue'
import CloseActionDialog from './components/CloseActionDialog.vue'
import { useAppTabs } from './composables/useAppTabs'
import { useAppService } from './composables/useAppService'
import { useQuickLaunch } from './composables/useQuickLaunch'
import { useTheme } from './composables/useTheme'
import { useSettings } from './composables/useSettings'
import { useTerminalEvent } from './composables/useTerminalEvent'
import { useBootProgress } from './composables/useBootProgress'
import { useTabSearch } from './composables/useTabSearch'
import { useAppProviders } from './composables/useAppProviders'
import { useAppBootstrap } from './composables/useAppBootstrap'
import { useTerminalCommandExecution } from './composables/useTerminalCommandExecution'
import { HideWindow, QuitApp } from '../wailsjs/go/main/App'

const SettingsDialog = defineAsyncComponent(() => import('./views/settings/SettingsDialog.vue'))
const NetworkPortList = defineAsyncComponent(() => import('./views/tools/NetworkPortList.vue'))
const MyApp = defineAsyncComponent(() => import('./views/myapp/MyApp.vue'))
const QuickLaunchDialogs = defineAsyncComponent(() => import('./views/quicklaunch/QuickLaunchDialogs.vue'))

const activeNav = ref('terminal')
const mainTabsBodyRef = ref(null)
const searchBarRef = ref(null)
const quickLaunchTabRef = ref(null)
const myAppDialogsRef = ref(null)
const quickLaunchDialogsRef = ref(null)
const settingsRef = ref(null)
const closeActionDialogRef = ref(null)

const {
  appReady,
  bootProgress,
  bootStatus,
  updateBootProgress,
  startBootProgress,
  finishBootProgress,
  cleanupBootProgress
} = useBootProgress()

const { currentTheme, changeTheme, loadTheme, applyTheme } = useTheme()
const {
  defaultShell,
  changeDefaultShell,
  changeCloseAction,
  changeGlobalHotkey,
  loadSettings,
  applyBootstrapSettings
} = useSettings()
const { sendCommand } = useTerminalEvent()

const {
  tabs,
  activeTabId,
  terminalTabs,
  appTabs,
  quickLaunchTab,
  toolTabs,
  getTabIcon,
  addTerminalTab,
  addAppTab,
  addQuickLaunchTab,
  addToolTab,
  switchTab,
  closeTab,
  closeAppTab
} = useAppTabs()

const appService = useAppService(closeAppTab)
const qlService = useQuickLaunch(addQuickLaunchTab)

/**
 * switchNav 负责左侧导航切换时的懒加载。
 * 仅在切换到对应模块时拉取必要数据，避免应用启动时做无效请求。
 */
const switchNav = (key) => {
  activeNav.value = key
  if (key === 'apps') {
    appService.loadApps()
    appService.loadServerStatus()
    return
  }
  if (key === 'shortcuts') {
    qlService.loadQlCmds()
    qlService.loadQlCategories()
  }
}

const {
  executeShortcutCommand,
  handleTerminalCommand
} = useTerminalCommandExecution({
  tabs,
  activeTabId,
  defaultShell,
  addTerminalTab,
  sendCommand
})

const {
  currentSearchVisible,
  handleGlobalKeyDown,
  handleSearchBarVisibleChange,
  handleSearchInput,
  handleSearchFindNext,
  handleSearchFindPrev,
  handleSearchClose,
  handleSearchResult
} = useTabSearch({
  tabs,
  activeTabId,
  mainTabsBodyRef,
  searchBarRef
})

/**
 * handleTabMouseDown 支持使用鼠标中键关闭可关闭的页签。
 */
const handleTabMouseDown = (event, tab) => {
  if (event.button === 1 && tab.closable !== false) {
    event.preventDefault()
    closeTab(tab.id)
  }
}

/**
 * handleCloseRequested 处理宿主窗口关闭请求。
 * 用户确认后，根据选择决定隐藏窗口还是退出应用。
 */
const handleCloseRequested = async () => {
  const result = await closeActionDialogRef.value.open()
  if (result.action === 'cancel') {
    return
  }
  if (result.dontAsk) {
    await changeCloseAction(result.action)
  }
  if (result.action === 'tray') {
    HideWindow()
  } else {
    QuitApp()
  }
}

/**
 * handleTabTitleChange 根据终端或 SSH 状态更新页签标题。
 */
const handleTabTitleChange = (event) => {
  const { tabId, host } = event.detail
  const tab = tabs.value.find(item => item.id === tabId)
  if (!tab) return

  if (host) {
    tab.title = host
    return
  }

  const match = tab.id.match(/terminal-\d+-(\d+)/)
  tab.title = `终端 ${match ? match[1] : ''}`
}

const { registerProviders } = useAppProviders({
  activeNav,
  switchNav,
  myAppDialogsRef,
  quickLaunchDialogsRef,
  settingsRef,
  quickLaunchTabRef,
  addAppTab,
  addToolTab,
  addQuickLaunchTab,
  addTerminalTab,
  defaultShell,
  executeShortcutCommand,
  handleTerminalCommand,
  sendCommand,
  appService,
  qlService
})

registerProviders()

const { bootstrapApp, cleanupBootstrapEffects } = useAppBootstrap({
  appService,
  qlService,
  loadTheme,
  loadSettings,
  applyTheme,
  applyBootstrapSettings,
  startBootProgress,
  updateBootProgress,
  finishBootProgress,
  handleGlobalKeyDown,
  handleSearchResult,
  handleTabTitleChange,
  handleCloseRequested
})

onMounted(async () => {
  await bootstrapApp()
})

onUnmounted(() => {
  cleanupBootProgress()
  cleanupBootstrapEffects()
})
</script>

<style scoped>
.app-shell {
  width: 100%;
  height: 100vh;
  overflow: hidden;
  background-color: var(--bg-secondary);
}

.app-boot-screen {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  background:
    radial-gradient(circle at top, rgba(64, 158, 255, 0.18), transparent 45%),
    var(--bg-secondary);
}

.boot-card {
  width: min(420px, calc(100% - 48px));
  padding: 28px 24px;
  border-radius: 16px;
  background-color: rgba(255, 255, 255, 0.04);
  box-shadow: 0 16px 40px rgba(0, 0, 0, 0.12);
  backdrop-filter: blur(8px);
}

.boot-title {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
}

.boot-subtitle {
  margin-top: 8px;
  font-size: 13px;
  color: var(--text-muted);
}

.boot-progress-track {
  margin-top: 22px;
  width: 100%;
  height: 8px;
  border-radius: 999px;
  background-color: var(--bg-hover);
  overflow: hidden;
}

.boot-progress-bar {
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, #409eff, #67c23a);
  transition: width 0.18s ease;
}

.boot-progress-text {
  margin-top: 10px;
  text-align: right;
  font-size: 12px;
  color: var(--text-faint);
}

.app-container {
  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 100%;
  overflow: hidden;
}

.app-workspace {
  flex: 1;
  min-height: 0;
  display: flex;
  overflow: hidden;
}

.right-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background-color: var(--bg-secondary);
  min-width: 0;
}

.empty-main {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.main-tabs-header {
  display: flex;
  align-items: center;
  padding: 4px 8px;
  background-color: var(--bg-primary);
  border-bottom: 1px solid var(--border-color);
  gap: 4px;
  flex-shrink: 0;
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
  background-color: var(--scrollbar-thumb);
  border-radius: 2px;
}

.main-tab-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
  max-width: 240px;
  padding: 6px 10px;
  border-radius: 6px 6px 0 0;
  cursor: pointer;
  color: var(--text-muted);
  background-color: transparent;
  user-select: none;
}

.main-tab-item.active {
  color: var(--text-primary);
  background-color: var(--bg-secondary);
}

.tab-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tab-close {
  flex-shrink: 0;
  border-radius: 4px;
}

.tab-close:hover {
  background-color: var(--bg-hover);
}

.main-tabs-body {
  flex: 1;
  min-height: 0;
  position: relative;
}

.app-iframe-wrapper {
  width: 100%;
  height: 100%;
}

.app-iframe {
  width: 100%;
  height: 100%;
  border: none;
  background-color: #fff;
}
</style>
