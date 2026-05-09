<template>
  <div class="app-container">
    <AppSidebar
      :active-nav="activeNav"
      :active-tab-id="activeTabId"
      :terminal-sub-tab="terminalSubTab"
      @update:terminal-sub-tab="terminalSubTab = $event"
      :nav-items="navItems"
      :apps="apps"
      :apps-loading="appsLoading"
      :get-app-icon-url="getAppIconUrl"
      :ql-groups="qlGroups"
      :ql-cmds="qlCmds"
      :expanded-ql-groups="expandedQlGroups"
      :ungrouped-ql-cmds="ungroupedQlCmds"
      :get-ql-cmds-by-group="getQlCmdsByGroup"
      :get-ql-cmd-count="getQlCmdCount"
      @switch-nav="switchNav"
      @terminal-ql-exec="handleTerminalQlExec"
      @terminal-history-exec="handleTerminalHistoryExec"
      @show-app-settings="showAppSettings"
      @show-app-import="showAppImport"
      @show-add-web-app="showAddWebAppDialog"
      @refresh-apps="refreshApps"
      @open-app="openAppHandler"
      @handle-app-cmd="handleAppCmd"
      @show-ql-add-dialog="showQlAddDialog"
      @show-ql-group-dialog="showQlGroupDialog"
      @toggle-ql-group="toggleQlGroup"
      @execute-ql-cmd="executeQlCmdWithRef"
      @edit-ql-cmd="editQlCmd"
      @delete-ql-cmd="deleteQlCmd"
      @open-tool="openTool"
      @open-settings="openSettings"
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
          <ToolsPage
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

    <AppDialogs
      ref="appDialogsRef"
      :app-settings-visible="appSettingsVisible"
      :static-dir="staticDir"
      :app-import-visible="appImportVisible"
      :app-import-tab="appImportTab"
      :import-zip-path="importZipPath"
      :import-dir-path="importDirPath"
      :import-app-name="importAppName"
      :app-edit-name-visible="appEditNameVisible"
      :app-edit-name-value="appEditNameValue"
      :app-rename-dir-visible="appRenameDirVisible"
      :app-rename-dir-value="appRenameDirValue"
      :web-app-dialog-visible="webAppDialogVisible"
      :is-editing-web-app="isEditingWebApp"
      :web-app-form="webAppForm"
      :ql-cmd-dialog-visible="qlCmdDialogVisible"
      :is-editing-ql-cmd="isEditingQlCmd"
      :ql-cmd-form="qlCmdForm"
      :ql-groups="qlGroups"
      :ql-group-dialog-visible="qlGroupDialogVisible"
      :new-group-name="newGroupName"
      @update:app-settings-visible="appSettingsVisible = $event"
      @update:static-dir="staticDir = $event"
      @update:app-import-visible="appImportVisible = $event"
      @update:app-import-tab="appImportTab = $event"
      @update:import-zip-path="importZipPath = $event"
      @update:import-dir-path="importDirPath = $event"
      @update:import-app-name="importAppName = $event"
      @update:app-edit-name-visible="appEditNameVisible = $event"
      @update:app-edit-name-value="appEditNameValue = $event"
      @update:app-rename-dir-visible="appRenameDirVisible = $event"
      @update:app-rename-dir-value="appRenameDirValue = $event"
      @update:web-app-dialog-visible="webAppDialogVisible = $event"
      @update-web-app-form="updateWebAppForm"
      @save-web-app="saveWebApp"
      @update:ql-cmd-dialog-visible="qlCmdDialogVisible = $event"
      @update:ql-group-dialog-visible="qlGroupDialogVisible = $event"
      @update:new-group-name="newGroupName = $event"
      @update-ql-cmd-form="handleUpdateQlCmdForm"
      @select-directory="selectDirectory"
      @save-static-dir="saveStaticDir"
      @select-zip-file="selectZipFile"
      @select-import-dir="selectImportDir"
      @do-import-zip="doImportZip"
      @do-import-dir="doImportDir"
      @save-app-display-name="saveAppDisplayName"
      @save-app-dir-name="saveAppDirName"
      @handle-icon-upload="handleIconUpload"
      @select-work-dir="selectWorkDir"
      @save-ql-cmd="saveQlCmd"
      @add-ql-group="addQlGroup"
      @delete-ql-group="deleteQlGroup"
    />

    <SettingsDialog
      ref="settingsRef"
      :visible="settingsVisible"
      :theme="currentTheme"
      :shell="defaultShell"
      @update:visible="settingsVisible = $event"
      @theme-change="changeTheme"
      @shell-change="changeDefaultShell"
    />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, computed, nextTick } from 'vue'
import { Monitor, Grid, Promotion, SetUp, Close, Plus } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import TerminalTab from './views/terminal/TerminalTab.vue'
import QuickLaunchTab from './views/quicklaunch/QuickLaunchTab.vue'
import SettingsDialog from './views/settings/SettingsDialog.vue'
import ToolsPage from './views/tools/ToolsPage.vue'
import AppSidebar from './views/app/AppSidebar.vue'
import AppDialogs from './views/app/AppDialogs.vue'
import SearchBar from './components/SearchBar.vue'
import { searchInContainer, findNextInContainer, findPrevInContainer, clearHighlights } from './utils/domSearch'
import { useAppTabs } from './composables/useAppTabs'
import { useAppService } from './composables/useAppService'
import { useQuickLaunch } from './composables/useQuickLaunch'
import { useTheme } from './composables/useTheme'
import { useSettings } from './composables/useSettings'
import { useTerminalEvent } from './composables/useTerminalEvent'

const navItems = [
  { key: 'terminal', label: '终端', icon: Monitor },
  { key: 'apps', label: '我的应用', icon: Grid },
  { key: 'shortcuts', label: '快速启动', icon: Promotion },
  { key: 'tools', label: '实用工具', icon: SetUp }
]

const activeNav = ref('terminal')
const terminalSubTab = ref('shortcuts')
const quickLaunchTabRef = ref(null)
const appDialogsRef = ref(null)
const settingsVisible = ref(false)
const settingsRef = ref(null)

const { currentTheme, changeTheme, loadTheme } = useTheme()
const { defaultShell, changeDefaultShell, loadSettings } = useSettings()
const { sendCommand, recordHistory } = useTerminalEvent()

const {
  tabs, activeTabId, terminalTabs, appTabs, quickLaunchTab, toolTabs,
  getTabIcon, addTerminalTab, addAppTab, addQuickLaunchTab, addToolTab,
  switchTab, getNavKeyByTab, closeTab, closeAppTab
} = useAppTabs()

const {
  apps, appsLoading, serverStatus,
  appSettingsVisible, staticDir,
  appImportVisible, appImportTab, importZipPath, importDirPath, importAppName,
  appEditNameVisible, appEditNameValue,
  appRenameDirVisible, appRenameDirValue,
  webAppDialogVisible, isEditingWebApp, webAppForm,
  loadApps, refreshApps, loadServerStatus,
  getAppIconUrl, openApp,
  showAppSettings, selectDirectory, saveStaticDir,
  showAppImport, selectZipFile, selectImportDir,
  doImportZip, doImportDir,
  showAddWebAppDialog, saveWebApp, updateWebAppForm,
  handleAppCmd, saveAppDisplayName, saveAppDirName, handleIconUpload
} = useAppService(closeAppTab, appDialogsRef)

const {
  qlGroups, qlCmds, expandedQlGroups,
  qlCmdDialogVisible, isEditingQlCmd, qlCmdForm,
  qlGroupDialogVisible, newGroupName,
  ungroupedQlCmds, getQlCmdsByGroup, getQlCmdCount,
  toggleQlGroup, loadQlGroups, loadQlCmds,
  executeQlCmd, showQlAddDialog, editQlCmd,
  selectWorkDir, saveQlCmd, deleteQlCmd,
  showQlGroupDialog, addQlGroup, deleteQlGroup
} = useQuickLaunch(addQuickLaunchTab)

/**
 * 切换左侧导航
 */
const switchNav = (key) => {
  activeNav.value = key
  if (key === 'terminal') {
    loadQlPanelData()
  } else if (key === 'apps') {
    loadApps()
    loadServerStatus()
  } else if (key === 'shortcuts') {
    loadQlCmds()
    loadQlGroups()
  }
}

/**
 * 终端快捷命令执行
 */
const handleTerminalQlExec = (command) => {
  if (terminalTabs.value.length === 0) {
    addTerminalTab(defaultShell.value)
  }
  sendCommand(activeTabId.value, command)
}

/**
 * 终端历史命令执行
 */
const handleTerminalHistoryExec = (command) => {
  if (terminalTabs.value.length === 0) {
    addTerminalTab(defaultShell.value)
  }
  sendCommand(activeTabId.value, command)
}

/**
 * 命令执行完成回调
 */
const handleCommandExecuted = (data) => {
  if (data && data.command) {
    recordHistory(data.command)
  }
}

/**
 * 向终端发送命令
 */
const handleSendCommand = (tabId, command) => {
  sendCommand(tabId, command)
}

/**
 * 加载快捷面板数据（预留接口）
 */
const loadQlPanelData = () => {}

/**
 * 执行快速启动命令
 */
const executeQlCmdWithRef = (cmd) => {
  executeQlCmd(cmd, quickLaunchTabRef)
}

/**
 * 更新快速启动命令表单字段
 */
const handleUpdateQlCmdForm = ({ key, value }) => {
  qlCmdForm.value[key] = value
}

/**
 * 打开工具
 */
const openTool = (toolKey, toolName) => {
  addToolTab(toolKey, toolName)
}

/**
 * 打开应用
 */
const openAppHandler = (app) => {
  openApp(app, addAppTab)
}

/**
 * 打开系统设置
 */
const openSettings = () => {
  if (settingsRef.value) settingsRef.value.handleOpen()
  settingsVisible.value = true
}

const searchBarRef = ref(null)
const mainTabsBodyRef = ref(null)
const searchVisibleMap = reactive({})
const lastSearchKeyword = reactive({})

const currentSearchVisible = computed(() => !!searchVisibleMap[activeTabId.value])

/**
 * 获取当前活动 tab 对应的 DOM 可搜索容器
 * 对于 iframe (app) 类型返回 null（不支持搜索）
 */
const getSearchableContainer = () => {
  const activeTab = tabs.value.find(t => t.id === activeTabId.value)
  if (!activeTab || !mainTabsBodyRef.value) return null
  if (activeTab.type === 'app') return null
  if (activeTab.type === 'terminal') return null
  return mainTabsBodyRef.value.querySelector(`[data-tab-id="${activeTabId.value}"]`)
}

/**
 * 全局按键处理
 */
const handleGlobalKeyDown = (e) => {
  // Ctrl+F 快捷键防止默认事件
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

/**
 * 搜索栏可见性变化
 */
const handleSearchBarVisibleChange = (val) => {
  searchVisibleMap[activeTabId.value] = val
  if (!val) {
    handleSearchCleanup()
  }
}

/**
 * 清理搜索状态（关闭搜索栏时调用）
 */
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

/**
 * 搜索输入
 */
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

/**
 * 搜索下一个
 */
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
    // iframe 不支持搜索
  } else {
    const container = getSearchableContainer()
    if (container) {
      const result = findNextInContainer(container)
      if (result && searchBarRef.value) searchBarRef.value.updateMatchInfo(result.current, result.total)
    }
  }
}

/**
 * 搜索上一个
 */
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
    // iframe 不支持搜索
  } else {
    const container = getSearchableContainer()
    if (container) {
      const result = findPrevInContainer(container)
      if (result && searchBarRef.value) searchBarRef.value.updateMatchInfo(result.current, result.total)
    }
  }
}

/**
 * 搜索关闭
 */
const handleSearchClose = () => {
  searchVisibleMap[activeTabId.value] = false
  handleSearchCleanup()
}

/**
 * 监听搜索结果事件，更新SearchBar的匹配信息
 */
const handleSearchResult = (event) => {
  const { tabId, resultIndex, resultCount } = event.detail
  if (tabId === activeTabId.value && searchBarRef.value) {
    searchBarRef.value.updateMatchInfo(resultIndex, resultCount)
  }
}

onMounted(async () => {
  await Promise.all([loadTheme(), loadSettings()])
  addTerminalTab(defaultShell.value)
  loadApps()
  loadQlGroups()
  loadQlCmds()
  window.addEventListener('keydown', handleGlobalKeyDown, true)
  window.addEventListener('tab-search-result', handleSearchResult)
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

.app-iframe {
  width: 100%;
  height: 100%;
  border: none;
  position: absolute;
  top: 0;
  left: 0;
}
</style>
