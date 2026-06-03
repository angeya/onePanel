import { provide } from 'vue'

/**
 * useAppProviders 统一注册 App 根组件向下游注入的能力。
 * 将对话框、服务实例、导航状态和终端操作集中暴露，减少 App.vue 内的 provide 样板代码。
 */
export function useAppProviders({
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
}) {
  /**
   * registerProviders 一次性完成 App 根级别的 provide 注入。
   * 这里暴露的都是各子组件实际依赖的能力，便于后续继续收敛边界。
   */
  const registerProviders = () => {
    provide('appService', appService)
    provide('qlService', qlService)

    provide('openMyAppDialog', (action, data) => {
      if (!myAppDialogsRef.value) return
      switch (action) {
        case 'import':
          myAppDialogsRef.value.showAppImport()
          break
        case 'addWebApp':
          myAppDialogsRef.value.showAddWebAppDialog()
          break
        case 'batchExport':
          myAppDialogsRef.value.showBatchExport()
          break
        case 'handleCmd':
          myAppDialogsRef.value.handleAppCmd(data.cmd, data.app)
          break
      }
    })

    provide('openQlDialog', (action, data) => {
      if (!quickLaunchDialogsRef.value) return
      switch (action) {
        case 'add':
          quickLaunchDialogsRef.value.showQlAddDialog()
          break
        case 'category':
          quickLaunchDialogsRef.value.showQlCategoryDialog()
          break
        case 'edit':
          quickLaunchDialogsRef.value.editQlCmd(data)
          break
      }
    })

    provide('openSettings', () => {
      settingsRef.value?.open()
    })

    provide('activeNav', activeNav)
    provide('switchNav', switchNav)
    provide('handleTerminalCommand', handleTerminalCommand)
    provide('addTerminalTab', addTerminalTab)
    provide('defaultShell', defaultShell)
    provide('executeShortcutCommand', executeShortcutCommand)
    provide('addAppTab', addAppTab)
    provide('addToolTab', addToolTab)
    provide('addQuickLaunchTab', addQuickLaunchTab)
    provide('quickLaunchTabRef', quickLaunchTabRef)
    provide('sendCommand', sendCommand)
  }

  return {
    registerProviders
  }
}
