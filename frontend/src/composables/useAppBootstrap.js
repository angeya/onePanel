import { EventsOn } from '../../wailsjs/runtime/runtime'
import { GetBootstrapSettings } from '../../wailsjs/go/main/SettingService'

/**
 * useAppBootstrap 负责编排 App 根组件启动阶段的初始化流程。
 * 统一处理基础配置预加载、常用数据预热和全局事件注册。
 */
export function useAppBootstrap({
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
  handleContextMenu,
  handleCloseRequested
}) {
  /**
   * bootstrapApp 执行应用启动初始化。
   * 先通过批量配置快速恢复关键状态，再并发加载常用数据。
   */
  const bootstrapApp = async () => {
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
  }

  /**
   * cleanupBootstrapEffects 卸载应用根组件注册的全局事件。
   */
  const cleanupBootstrapEffects = () => {
    window.removeEventListener('keydown', handleGlobalKeyDown, true)
    window.removeEventListener('tab-search-result', handleSearchResult)
    window.removeEventListener('tab-title-change', handleTabTitleChange)
    window.removeEventListener('contextmenu', handleContextMenu)
  }

  return {
    bootstrapApp,
    cleanupBootstrapEffects
  }
}
