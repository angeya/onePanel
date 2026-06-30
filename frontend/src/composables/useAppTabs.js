import { ref, computed } from 'vue'
import { Monitor, Grid, Promotion, SetUp } from '@element-plus/icons-vue'

/**
 * Tab 管理组合式函数
 * 负责右侧 tab 页的增删切换逻辑
 */
export function useAppTabs() {
  const tabs = ref([])
  const activeTabId = ref('')
  let tabCounter = 0

  const terminalTabs = computed(() => tabs.value.filter(t => t.type === 'terminal'))
  const appTabs = computed(() => tabs.value.filter(t => t.type === 'app'))
  const quickLaunchTab = computed(() => tabs.value.find(t => t.type === 'quick-launch'))
  const toolTabs = computed(() => tabs.value.filter(t => t.type === 'tool'))

  /**
   * 获取 tab 类型对应的图标
   */
  const getTabIcon = (tab) => {
    switch (tab.type) {
      case 'terminal': return Monitor
      case 'app': return Grid
      case 'quick-launch': return Promotion
      case 'tool': return SetUp
      default: return Monitor
    }
  }

  /**
   * 添加终端 tab
   */
  const addTerminalTab = (shell = 'cmd.exe', title = '') => {
    tabCounter++
    const id = `terminal-${Date.now()}-${tabCounter}`
    tabs.value.push({
      id,
      type: 'terminal',
      title: title || `终端 ${tabCounter}`,
      shell,
      closable: true
    })
    activeTabId.value = id
    return id
  }

  /**
   * 添加应用 tab（如已存在则切换）
   */
  const addAppTab = (appId, name, url) => {
    console.log(`添加应用 tab: ${appId}, ${name}, ${Object.assign({}, tabs.value)}`)
    const existing = tabs.value.find(t => t.type === 'app' && t.appId === appId)
    if (existing) {
      activeTabId.value = existing.id
      return
    }
    tabCounter++
    const id = `app-${Date.now()}-${tabCounter}`
    tabs.value.push({
      id,
      type: 'app',
      title: name,
      appId,
      url,
      closable: true
    })
    activeTabId.value = id
  }

  /**
   * 添加或切换到快速启动 tab
   */
  const addQuickLaunchTab = () => {
    let tab = quickLaunchTab.value
    if (!tab) {
      tabCounter++
      tab = {
        id: `quick-launch-${Date.now()}-${tabCounter}`,
        type: 'quick-launch',
        title: '快速启动',
        closable: true
      }
      tabs.value.push(tab)
    }
    activeTabId.value = tab.id
  }

  /**
   * 添加工具 tab（如已存在则切换）
   */
  const addToolTab = (toolKey, toolName) => {
    const existing = tabs.value.find(t => t.type === 'tool' && t.toolKey === toolKey)
    if (existing) {
      activeTabId.value = existing.id
      return
    }
    tabCounter++
    const id = `tool-${Date.now()}-${tabCounter}`
    tabs.value.push({
      id,
      type: 'tool',
      title: toolName,
      toolKey,
      closable: true
    })
    activeTabId.value = id
  }

  /**
   * 切换 tab
   */
  const switchTab = (id) => {
    activeTabId.value = id
  }

  /**
   * 根据 tab 类型获取对应导航 key
   */
  const getNavKeyByTab = (tab) => {
    switch (tab.type) {
      case 'terminal': return 'terminal'
      case 'app': return 'apps'
      case 'quick-launch': return 'shortcuts'
      case 'tool': return 'tools'
      default: return 'terminal'
    }
  }

  /**
   * 关闭 tab
   */
  const closeTab = (id) => {
    const index = tabs.value.findIndex(t => t.id === id)
    if (index === -1) return
    tabs.value.splice(index, 1)
    if (tabs.value.length > 0 && activeTabId.value === id) {
      const newIndex = Math.min(index, tabs.value.length - 1)
      activeTabId.value = tabs.value[newIndex].id
    } else if (tabs.value.length === 0) {
      activeTabId.value = ''
    }
  }

  /**
   * 关闭指定应用 ID 的 tab
   */
  const closeAppTab = (appId) => {
    const appTab = tabs.value.find(t => t.type === 'app' && t.appId === appId)
    if (appTab) closeTab(appTab.id)
  }

  return {
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
    getNavKeyByTab,
    closeTab,
    closeAppTab
  }
}
