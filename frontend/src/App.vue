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

      <MyApp ref="myAppDialogsRef" />
      <QuickLaunchDialogs ref="quickLaunchDialogsRef" />

      <SettingsDialog
        ref="settingsRef"
        @theme-change="changeTheme"
        @shell-change="changeDefaultShell"
        @close-action-change="changeCloseAction"
      />
      <CloseActionDialog ref="closeActionDialogRef" />
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, computed, watch, provide, defineAsyncComponent } from 'vue'
import { Close } from '@element-plus/icons-vue'
import { GetBootstrapSettings } from '../wailsjs/go/main/SettingService'
import TerminalTab from './views/terminal/TerminalTab.vue'
import QuickLaunchTab from './views/quicklaunch/QuickLaunchTab.vue'
const SettingsDialog = defineAsyncComponent(() => import('./views/settings/SettingsDialog.vue'))
const NetworkPortList = defineAsyncComponent(() => import('./views/tools/NetworkPortList.vue'))
import Sidebar from './views/sidebar/Sidebar.vue'
const MyApp = defineAsyncComponent(() => import('./views/myapp/MyApp.vue'))
const QuickLaunchDialogs = defineAsyncComponent(() => import('./views/quicklaunch/QuickLaunchDialogs.vue'))
import SearchBar from './components/SearchBar.vue'
import CloseActionDialog from './components/CloseActionDialog.vue'
import { searchInContainer, findNextInContainer, findPrevInContainer, clearHighlights } from './utils/domSearch'
import { useAppTabs } from './composables/useAppTabs'
import { useAppService } from './composables/useAppService'
import { useQuickLaunch } from './composables/useQuickLaunch'
import { useTheme } from './composables/useTheme'
import { useSettings } from './composables/useSettings'
import { useTerminalEvent } from './composables/useTerminalEvent'
import { EventsOn } from '../wailsjs/runtime/runtime'
import { HideWindow, QuitApp } from '../wailsjs/go/main/App'

const appReady = ref(false)
const bootProgress = ref(0)
const bootStatus = ref('正在初始化界面')
let bootProgressTimer = null

const updateBootProgress = (value, status) => {
  bootProgress.value = Math.max(0, Math.min(100, value))
  if (status) {
    bootStatus.value = status
  }
}

const startBootProgress = () => {
  clearInterval(bootProgressTimer)
  updateBootProgress(8, '正在初始化界面')
  bootProgressTimer = window.setInterval(() => {
    if (bootProgress.value < 88) {
      bootProgress.value += bootProgress.value < 40 ? 8 : 4
    }
  }, 120)
}

const finishBootProgress = () => {
  clearInterval(bootProgressTimer)
  updateBootProgress(100, '初始化完成')
  window.setTimeout(() => {
    appReady.value = true
  }, 180)
}

const activeNav = ref('terminal')
const quickLaunchTabRef = ref(null)
const myAppDialogsRef = ref(null)
const quickLaunchDialogsRef = ref(null)
const settingsRef = ref(null)
const closeActionDialogRef = ref(null)
const { currentTheme, changeTheme, loadTheme, applyTheme } = useTheme()
const { defaultShell, allowDebug, changeDefaultShell, changeCloseAction, changeAllowDebug, loadSettings, applyBootstrapSettings } = useSettings()
const { sendCommand } = useTerminalEvent()

const {
  tabs, activeTabId, terminalTabs, appTabs, quickLaunchTab, toolTabs,
  getTabIcon, addTerminalTab, addAppTab, addQuickLaunchTab, addToolTab,
  switchTab, closeTab, closeAppTab
} = useAppTabs()

const appService = useAppService(closeAppTab)
const qlService = useQuickLaunch(addQuickLaunchTab)

provide('appService', appService)
provide('qlService', qlService)

provide('openMyAppDialog', (action, data) => {
  if (!myAppDialogsRef.value) return
  switch (action) {
    case 'import': myAppDialogsRef.value.showAppImport(); break
    case 'addWebApp': myAppDialogsRef.value.showAddWebAppDialog(); break
    case 'batchExport': myAppDialogsRef.value.showBatchExport(); break
    case 'handleCmd': myAppDialogsRef.value.handleAppCmd(data.cmd, data.app); break
  }
})

provide('openQlDialog', (action, data) => {
  if (!quickLaunchDialogsRef.value) return
  switch (action) {
    case 'add': quickLaunchDialogsRef.value.showQlAddDialog(); break
    case 'category': quickLaunchDialogsRef.value.showQlCategoryDialog(); break
    case 'edit': quickLaunchDialogsRef.value.editQlCmd(data); break
  }
})

provide('openSettings', () => {
  if (settingsRef.value) settingsRef.value.open()
})

provide('addAppTab', addAppTab)
provide('addToolTab', addToolTab)
provide('addQuickLaunchTab', addQuickLaunchTab)
provide('quickLaunchTabRef', quickLaunchTabRef)
provide('sendCommand', sendCommand)

const switchNav = (key) => {
  activeNav.value = key
  if (key === 'terminal') {
  } else if (key === 'apps') {
    appService.loadApps()
    appService.loadServerStatus()
  } else if (key === 'shortcuts') {
    qlService.loadQlCmds()
    qlService.loadQlCategories()
  }
}

const createTerminalAndRunCommands = (commandLines, title = '') => {
  const tabId = addTerminalTab(defaultShell.value, title)
  if (!tabId) {
    return
  }

  const handleReady = (event) => {
    if (event.detail.tabId !== tabId) {
      return
    }

    window.removeEventListener('terminal-ready', handleReady)
    commandLines.forEach((line) => {
      sendCommand(tabId, line)
    })
  }

  window.addEventListener('terminal-ready', handleReady)
}

const executeShortcutCommand = ({ commandLines, commandName, workDir = '', forceNewTerminal = false }) => {
  const lines = (commandLines || []).filter((line) => line && line.trim()).map((line) => line.trim())
  if (lines.length === 0) {
    return
  }

  const finalLines = workDir
    ? [`cd /d ${workDir}`, ...lines]
    : lines

  const activeTab = tabs.value.find((tab) => tab.id === activeTabId.value)
  const shouldUseCurrentTerminal = !forceNewTerminal && activeTab?.type === 'terminal'

  if (shouldUseCurrentTerminal) {
    finalLines.forEach((line) => {
      sendCommand(activeTab.id, line)
    })
    return
  }

  createTerminalAndRunCommands(finalLines, forceNewTerminal ? commandName : '')
}

const handleTerminalCommand = (command) => {
  const line = command?.trim()
  if (!line) {
    return
  }
  executeShortcutCommand({ commandLines: [line] })
}

provide('activeNav', activeNav)
provide('switchNav', switchNav)
provide('handleTerminalCommand', handleTerminalCommand)
provide('addTerminalTab', addTerminalTab)
provide('defaultShell', defaultShell)
provide('allowDebug', allowDebug)
provide('changeAllowDebug', changeAllowDebug)
provide('executeShortcutCommand', executeShortcutCommand)

const handleTabMouseDown = (event, tab) => {
  if (event.button === 1 && tab.closable !== false) {
    event.preventDefault()
    closeTab(tab.id)
  }
}

const searchBarRef = ref(null)
const mainTabsBodyRef = ref(null)
const searchVisibleMap = reactive({})
const lastSearchKeyword = reactive({})

watch(tabs, (currentTabs) => {
  const activeIds = new Set(currentTabs.map(t => t.id))
  for (const key of Object.keys(searchVisibleMap)) {
    if (!activeIds.has(key)) {
      delete searchVisibleMap[key]
    }
  }
  for (const key of Object.keys(lastSearchKeyword)) {
    if (!activeIds.has(key)) {
      delete lastSearchKeyword[key]
    }
  }
}, { deep: true })

const currentSearchVisible = computed(() => !!searchVisibleMap[activeTabId.value])

const getSearchableContainer = () => {
  const activeTab = tabs.value.find(t => t.id === activeTabId.value)
  if (!activeTab || !mainTabsBodyRef.value) return null
  if (activeTab.type === 'app') return null
  if (activeTab.type === 'terminal') return null
  return mainTabsBodyRef.value.querySelector(`[data-tab-id="${activeTabId.value}"]`)
}

const handleGlobalKeyDown = (e) => {
  if ((e.ctrlKey || e.metaKey) && e.key === 'f') {
    e.preventDefault()
    e.stopPropagation()
    const tabId = activeTabId.value
    if (!tabId) return
    if (searchVisibleMap[tabId]) {
      searchBarRef.value?.focus()
    } else {
      searchVisibleMap[tabId] = true
    }
    return
  }
  if (e.key === 'Escape' && searchVisibleMap[activeTabId.value]) {
    e.stopPropagation()
    handleSearchClose()
  }
}

const handleSearchBarVisibleChange = (val) => {
  searchVisibleMap[activeTabId.value] = val
  if (!val) {
    handleSearchCleanup()
  }
}

const handleSearchCleanup = () => {
  const activeTab = tabs.value.find(t => t.id === activeTabId.value)
  if (activeTab?.type === 'terminal') {
    const event = new CustomEvent('tab-search-close', {
      detail: { tabId: activeTabId.value }
    })
    window.dispatchEvent(event)
  } else {
    const container = getSearchableContainer()
    if (container) clearHighlights(container)
  }
}

const handleSearchInput = (keyword) => {
  lastSearchKeyword[activeTabId.value] = keyword
  const activeTab = tabs.value.find(t => t.id === activeTabId.value)
  if (!activeTab) return

  if (activeTab.type === 'terminal') {
    const event = new CustomEvent('tab-search', {
      detail: { tabId: activeTabId.value, action: 'search', keyword }
    })
    window.dispatchEvent(event)
  } else if (activeTab.type === 'app') {
    if (searchBarRef.value) searchBarRef.value.setUnsupported()
  } else {
    const container = getSearchableContainer()
    if (container) {
      if (keyword) {
        const result = searchInContainer(container, keyword)
        if (searchBarRef.value) searchBarRef.value.updateMatchInfo(result.current, result.total)
      } else {
        clearHighlights(container)
        if (searchBarRef.value) searchBarRef.value.clearMatchInfo()
      }
    }
  }
}

const handleSearchFindNext = (keyword) => {
  lastSearchKeyword[activeTabId.value] = keyword
  const activeTab = tabs.value.find(t => t.id === activeTabId.value)
  if (!activeTab) return

  if (activeTab.type === 'terminal') {
    const event = new CustomEvent('tab-search', {
      detail: { tabId: activeTabId.value, action: 'findNext', keyword }
    })
    window.dispatchEvent(event)
  } else if (activeTab.type === 'app') {
  } else {
    const container = getSearchableContainer()
    if (container) {
      const result = findNextInContainer(container)
      if (result && searchBarRef.value) searchBarRef.value.updateMatchInfo(result.current, result.total)
    }
  }
}

const handleSearchFindPrev = (keyword) => {
  lastSearchKeyword[activeTabId.value] = keyword
  const activeTab = tabs.value.find(t => t.id === activeTabId.value)
  if (!activeTab) return

  if (activeTab.type === 'terminal') {
    const event = new CustomEvent('tab-search', {
      detail: { tabId: activeTabId.value, action: 'findPrev', keyword }
    })
    window.dispatchEvent(event)
  } else if (activeTab.type === 'app') {
  } else {
    const container = getSearchableContainer()
    if (container) {
      const result = findPrevInContainer(container)
      if (result && searchBarRef.value) searchBarRef.value.updateMatchInfo(result.current, result.total)
    }
  }
}

const handleSearchClose = () => {
  searchVisibleMap[activeTabId.value] = false
  handleSearchCleanup()
}

const handleSearchResult = (event) => {
  const { tabId, resultIndex, resultCount } = event.detail
  if (tabId === activeTabId.value && searchBarRef.value) {
    searchBarRef.value.updateMatchInfo(resultIndex, resultCount)
  }
}

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

const handleContextMenu = (e) => {
  if (allowDebug.value) {
    const terminalEl = e.target?.closest?.('.terminal-tab-container')
    if (terminalEl) {
      e.preventDefault()
    }
    return
  }
  e.preventDefault()
}

const handleTabTitleChange = (event) => {
  const { tabId, host } = event.detail
  const tab = tabs.value.find(t => t.id === tabId)
  if (!tab) return
  if (host) {
    tab.title = host
  } else {
    const match = tab.id.match(/terminal-\d+-(\d+)/)
    tab.title = `终端 ${match ? match[1] : ''}`
  }
}

onMounted(async () => {
  startBootProgress()
  updateBootProgress(18, '正在加载基础设置')

  try {
    const bootstrapSettings = await GetBootstrapSettings()
    applyTheme(bootstrapSettings.theme)
    applyBootstrapSettings(bootstrapSettings)
  } catch (err) {
    console.error('加载启动设置失败:', err)
  }

  updateBootProgress(42, '正在预加载常用数据')
  await Promise.allSettled([
    loadTheme(),
    loadSettings(),
    appService.loadApps(),
    qlService.loadQlCategories(),
    qlService.loadQlCmds()
  ])

  updateBootProgress(86, '正在绑定事件')
  window.addEventListener('keydown', handleGlobalKeyDown, true)
  window.addEventListener('tab-search-result', handleSearchResult)
  window.addEventListener('tab-title-change', handleTabTitleChange)
  window.addEventListener('contextmenu', handleContextMenu)
  EventsOn('close-requested', handleCloseRequested)
  finishBootProgress()
})

onUnmounted(() => {
  clearInterval(bootProgressTimer)
  window.removeEventListener('keydown', handleGlobalKeyDown, true)
  window.removeEventListener('tab-search-result', handleSearchResult)
  window.removeEventListener('tab-title-change', handleTabTitleChange)
  window.removeEventListener('contextmenu', handleContextMenu)
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
  height: 100vh;
  width: 100%;
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
