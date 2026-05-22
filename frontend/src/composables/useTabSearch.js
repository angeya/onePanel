import { computed, reactive, watch } from 'vue'
import { searchInContainer, findNextInContainer, findPrevInContainer, clearHighlights } from '../utils/domSearch'

/**
 * useTabSearch 管理主工作区内非终端页签和终端页签的搜索状态。
 * 对终端类型通过事件分发交给 TerminalTab 处理，对普通页面直接操作 DOM 高亮。
 */
export function useTabSearch({ tabs, activeTabId, mainTabsBodyRef, searchBarRef }) {
  const searchVisibleMap = reactive({})
  const lastSearchKeyword = reactive({})

  /**
   * 监听 tab 集合变化，及时清理已关闭页签遗留的搜索状态。
   */
  watch(tabs, (currentTabs) => {
    const activeIds = new Set(currentTabs.map(tab => tab.id))
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

  /**
   * getActiveTab 返回当前激活页签，减少各搜索方法中的重复查找逻辑。
   */
  const getActiveTab = () => tabs.value.find(tab => tab.id === activeTabId.value)

  /**
   * getSearchableContainer 返回当前可直接进行 DOM 搜索的容器。
   * 应用 iframe 和终端页签不走这里的高亮搜索逻辑。
   */
  const getSearchableContainer = () => {
    const activeTab = getActiveTab()
    if (!activeTab || !mainTabsBodyRef.value) return null
    if (activeTab.type === 'app' || activeTab.type === 'terminal') return null
    return mainTabsBodyRef.value.querySelector(`[data-tab-id="${activeTabId.value}"]`)
  }

  /**
   * dispatchTerminalSearchEvent 统一派发终端搜索事件。
   * TerminalTab 会根据 action 与 keyword 自行执行 xterm 搜索。
   */
  const dispatchTerminalSearchEvent = (action, keyword = '') => {
    window.dispatchEvent(new CustomEvent('tab-search', {
      detail: { tabId: activeTabId.value, action, keyword }
    }))
  }

  /**
   * handleSearchCleanup 清理当前激活页签的搜索态与高亮结果。
   */
  const handleSearchCleanup = () => {
    const activeTab = getActiveTab()
    if (activeTab?.type === 'terminal') {
      window.dispatchEvent(new CustomEvent('tab-search-close', {
        detail: { tabId: activeTabId.value }
      }))
      return
    }

    const container = getSearchableContainer()
    if (container) {
      clearHighlights(container)
    }
  }

  /**
   * handleSearchBarVisibleChange 响应搜索框显隐变化。
   * 当搜索条关闭时，需要同步移除当前页签的高亮状态。
   */
  const handleSearchBarVisibleChange = (visible) => {
    searchVisibleMap[activeTabId.value] = visible
    if (!visible) {
      handleSearchCleanup()
    }
  }

  /**
   * handleSearchInput 处理关键词输入事件。
   * 终端页签通过事件转发，普通页面直接执行全文搜索和高亮。
   */
  const handleSearchInput = (keyword) => {
    lastSearchKeyword[activeTabId.value] = keyword
    const activeTab = getActiveTab()
    if (!activeTab) return

    if (activeTab.type === 'terminal') {
      dispatchTerminalSearchEvent('search', keyword)
      return
    }

    if (activeTab.type === 'app') {
      searchBarRef.value?.setUnsupported()
      return
    }

    const container = getSearchableContainer()
    if (!container) return

    if (keyword) {
      const result = searchInContainer(container, keyword)
      searchBarRef.value?.updateMatchInfo(result.current, result.total)
      return
    }

    clearHighlights(container)
    searchBarRef.value?.clearMatchInfo()
  }

  /**
   * handleSearchFindNext 查找下一项匹配结果。
   */
  const handleSearchFindNext = (keyword) => {
    lastSearchKeyword[activeTabId.value] = keyword
    const activeTab = getActiveTab()
    if (!activeTab) return

    if (activeTab.type === 'terminal') {
      dispatchTerminalSearchEvent('findNext', keyword)
      return
    }

    if (activeTab.type === 'app') {
      return
    }

    const container = getSearchableContainer()
    if (!container) return

    const result = findNextInContainer(container)
    if (result) {
      searchBarRef.value?.updateMatchInfo(result.current, result.total)
    }
  }

  /**
   * handleSearchFindPrev 查找上一项匹配结果。
   */
  const handleSearchFindPrev = (keyword) => {
    lastSearchKeyword[activeTabId.value] = keyword
    const activeTab = getActiveTab()
    if (!activeTab) return

    if (activeTab.type === 'terminal') {
      dispatchTerminalSearchEvent('findPrev', keyword)
      return
    }

    if (activeTab.type === 'app') {
      return
    }

    const container = getSearchableContainer()
    if (!container) return

    const result = findPrevInContainer(container)
    if (result) {
      searchBarRef.value?.updateMatchInfo(result.current, result.total)
    }
  }

  /**
   * handleSearchClose 主动关闭搜索条并清理搜索状态。
   */
  const handleSearchClose = () => {
    searchVisibleMap[activeTabId.value] = false
    handleSearchCleanup()
  }

  /**
   * handleSearchResult 接收终端搜索结果并同步到 SearchBar 显示。
   */
  const handleSearchResult = (event) => {
    const { tabId, resultIndex, resultCount } = event.detail
    if (tabId === activeTabId.value) {
      searchBarRef.value?.updateMatchInfo(resultIndex, resultCount)
    }
  }

  /**
   * handleGlobalKeyDown 处理全局搜索快捷键与 ESC 关闭逻辑。
   */
  const handleGlobalKeyDown = (event) => {
    if ((event.ctrlKey || event.metaKey) && event.key === 'f') {
      event.preventDefault()
      event.stopPropagation()
      const tabId = activeTabId.value
      if (!tabId) return
      if (searchVisibleMap[tabId]) {
        searchBarRef.value?.focus()
      } else {
        searchVisibleMap[tabId] = true
      }
      return
    }

    if (event.key === 'Escape' && searchVisibleMap[activeTabId.value]) {
      event.stopPropagation()
      handleSearchClose()
    }
  }

  return {
    currentSearchVisible,
    handleGlobalKeyDown,
    handleSearchBarVisibleChange,
    handleSearchInput,
    handleSearchFindNext,
    handleSearchFindPrev,
    handleSearchClose,
    handleSearchResult
  }
}
