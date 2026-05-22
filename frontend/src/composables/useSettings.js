import { ref } from 'vue'
import { GetSetting, SetSetting } from '../../wailsjs/go/main/SettingService'
import { GetCloseAction, SetCloseAction, SetAllowDebug } from '../../wailsjs/go/main/App'

export function useSettings() {
  const defaultShell = ref('cmd.exe')
  const closeAction = ref('ask')
  const allowDebug = ref(false)

  const applyBootstrapSettings = (settings = {}) => {
    if (settings.default_shell) {
      defaultShell.value = settings.default_shell
    }
    if (settings.close_action) {
      closeAction.value = settings.close_action
    }
    if (Object.prototype.hasOwnProperty.call(settings, 'allow_debug')) {
      allowDebug.value = String(settings.allow_debug) === 'true'
    }
  }

  const changeDefaultShell = async (shell) => {
    try {
      await SetSetting('default_shell', shell)
      defaultShell.value = shell
    } catch (err) {
      console.error('保存默认终端失败:', err)
    }
  }

  const changeCloseAction = async (action) => {
    try {
      await SetCloseAction(action)
      closeAction.value = action
    } catch (err) {
      console.error('保存关闭行为设置失败:', err)
    }
  }

  const changeAllowDebug = async (value) => {
    await SetAllowDebug(value)
    allowDebug.value = value
  }

  const loadSettings = async () => {
    try {
      const shell = await GetSetting('default_shell')
      if (shell) {
        defaultShell.value = shell
      }

      const action = await GetCloseAction()
      if (action) {
        closeAction.value = action
      }

      const debug = await GetSetting('allow_debug')
      allowDebug.value = String(debug) === 'true'
    } catch (err) {
      console.error('加载设置失败:', err)
    }
  }

  return {
    defaultShell,
    closeAction,
    allowDebug,
    applyBootstrapSettings,
    changeDefaultShell,
    changeCloseAction,
    changeAllowDebug,
    loadSettings
  }
}
