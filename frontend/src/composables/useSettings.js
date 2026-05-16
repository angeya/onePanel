import { ref } from 'vue'
import { GetSetting, SetSetting } from '../../wailsjs/go/main/SettingService'
import { GetCloseAction, SetCloseAction, SetAllowDebug } from '../../wailsjs/go/main/App'

/**
 * 系统设置组合式函数
 * 负责默认 Shell、关闭行为、调试开关等系统级配置的加载和持久化
 */
export function useSettings() {
  const defaultShell = ref('cmd.exe')
  const closeAction = ref('ask')
  const allowDebug = ref(false)

  /**
   * 切换默认终端并持久化
   */
  const changeDefaultShell = async (shell) => {
    try {
      await SetSetting('default_shell', shell)
      defaultShell.value = shell
    } catch (err) {
      console.error('保存默认终端失败:', err)
    }
  }

  /**
   * 切换关闭行为并持久化
   * action: "tray"（最小化到托盘）或 "close"（直接退出）
   */
  const changeCloseAction = async (action) => {
    try {
      await SetCloseAction(action)
      closeAction.value = action
    } catch (err) {
      console.error('保存关闭行为设置失败:', err)
    }
  }

  /**
   * 切换调试开关并持久化
   * 同时通知后端动态控制 WebView2 右键菜单的启用状态
   */
  const changeAllowDebug = async (value) => {
    try {
      await SetSetting('allow_debug', value ? 'true' : 'false')
      allowDebug.value = value
      await SetAllowDebug(value)
    } catch (err) {
      console.error('保存调试开关设置失败:', err)
    }
  }

  /**
   * 从后端加载已保存的设置项
   */
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
      if (debug === 'true') {
        allowDebug.value = true
      }
    } catch (err) {
      console.error('加载设置失败:', err)
    }
  }

  return {
    defaultShell,
    closeAction,
    allowDebug,
    changeDefaultShell,
    changeCloseAction,
    changeAllowDebug,
    loadSettings
  }
}
