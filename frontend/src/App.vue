<template>
  <div class="app-container">
    <Sidebar
      :active-nav="activeNav"
      :terminal-sub-tab="terminalSubTab"
      @update:terminal-sub-tab="terminalSubTab = $event"
      :nav-items="navItems"
      @switch-nav="switchNav"
      @terminal-command="handleTerminalCommand"
    />

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
          <el-button
            v-if="activeNav === 'terminal'"
            class="tab-add"
            size="small"
            @click="addTerminalTab(defaultShell)"
            :icon="Plus"
            circle
          />
        </div>
        <div class="main-tabs-body" ref="mainTabsBodyRef">
          <TerminalTab
            v-for="tab in terminalTabs"
            :key="tab.id"
            :tab-id="tab.id"
            :shell="tab.shell || 'cmd.exe'"
            :theme="currentTheme"
            v-show="activeTabId === tab.id"
            @command-executed="handleCommandExecuted"
            @send-command="handleSendCommand"
          />
          <iframe
            v-for="tab in appTabs"
            :key="tab.id"
            v-show="activeTabId === tab.id"
            :src="tab.url"
            class="app-iframe"
            :data-tab-id="tab.id"
            sandbox="allow-scripts allow-same-origin allow-forms allow-popups"
          />
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
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, computed, watch, provide, defineAsyncComponent } from 'vue'
import { Monitor, Grid, Promotion, SetUp, Close, Plus } from '@element-plus/icons-vue'
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

const navItems = [
  { key: 'terminal', label: '终端', icon: Monitor },
  { key: 'apps', label: '我的应用', icon: Grid },
  { key: 'shortcuts', label: '快速启动', icon: Promotion },
  { key: 'tools', label: '实用工具', icon: SetUp }
]

const activeNav = ref('terminal')
const terminalSubTab = ref('shortcuts')
const quickLaunchTabRef = ref(null)
const myAppDialogsRef = ref(null)
const quickLaunchDialogsRef = ref(null)
const settingsRef = ref(null)
const closeActionDialogRef = ref(null)

const { currentTheme, changeTheme, loadTheme } = useTheme()
const { defaultShell, closeAction, changeDefaultShell, changeCloseAction, loadSettings } = useSettings()
const { sendCommand, recordHistory } = useTerminalEvent()

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

const handleTerminalCommand = (command) => {
  if (terminalTabs.value.length === 0) {
    addTerminalTab(defaultShell.value)
  }
  sendCommand(activeTabId.value, command)
}

const handleCommandExecuted = (data) => {
  if (data && data.command) {
    recordHistory(data.command)
  }
}

const handleSendCommand = (tabId, command) => {
  sendCommand(tabId, command)
}

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
      const result = searchInContainer(container, keyword)
      if (searchBarRef.value) searchBarRef.value.updateMatchInfo(result.current, result.total)
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
  if (result.dontAsk) {
    await changeCloseAction(result.action)
  }
  if (result.action === 'tray') {
    HideWindow()
  } else {
    QuitApp()
  }
}

onMounted(async () => {
  await Promise.all([loadTheme(), loadSettings()])
  addTerminalTab(defaultShell.value)
  appService.loadApps()
  qlService.loadQlCategories()
  qlService.loadQlCmds()
  window.addEventListener('keydown', handleGlobalKeyDown, true)
  window.addEventListener('tab-search-result', handleSearchResult)
  EventsOn('close-requested', handleCloseRequested)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleGlobalKeyDown, true)
  window.removeEventListener('tab-search-result', handleSearchResult)
})
</script>

<style scoped>
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
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: 4px;
  cursor: pointer;
  white-space: nowrap;
  font-size: 13px;
  color: var(--text-muted);
  background-color: var(--bg-tertiary);
  transition: all 0.15s;
  min-width: 0;
}

.main-tab-item:hover {
  background-color: var(--bg-active);
  color: var(--text-primary);
}

.main-tab-item.active {
  background-color: var(--bg-active);
  color: var(--text-primary);
  border-bottom: 2px solid var(--accent);
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

.main-tab-item:hover .tab-close {
  opacity: 1;
}

.tab-close:hover {
  background-color: var(--scrollbar-thumb);
  color: #fff;
}

.tab-add {
  flex-shrink: 0;
  background-color: var(--bg-tertiary) !important;
  border-color: var(--border-light) !important;
  color: var(--text-muted) !important;
}

.tab-add:hover {
  color: var(--accent) !important;
}

.main-tabs-body {
  flex: 1;
  overflow: hidden;
  position: relative;
}
</style>
