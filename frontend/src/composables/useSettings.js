import { ref } from 'vue'
import { GetSetting, SetSetting } from '../../wailsjs/go/main/SettingService'
import { GetCloseAction, SetCloseAction, SetAllowDebug } from '../../wailsjs/go/main/App'

/**
 * useSettings 管理系统设置相关状态。
 * 统一处理默认终端、关闭行为和调试开关的读取与持久化。
 */
export function useSettings() {
  const defaultShell = ref('cmd.exe')
  const closeAction = ref('ask')
  const allowDebug = ref(false)

  /**
   * applyBootstrapSettings 应用启动阶段批量读取到的设置值。
   * 该方法只负责更新本地状态，不发起额外请求。
   */
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

  /**
   * changeDefaultShell 切换默认终端并持久化到后端配置。
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
   * changeCloseAction 切换窗口关闭行为并同步到后端。
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
   * changeAllowDebug 切换调试开关。
   * 后端除了保存配置，还会同步控制 WebView2 右键菜单权限。
   */
  const changeAllowDebug = async (value) => {
    await SetAllowDebug(value)
    allowDebug.value = value
  }

  /**
   * loadSettings 兼容非启动时机的设置刷新。
   * 例如设置窗口重新打开时，需要再次从后端读取最新值。
   */
  const loadSettings = async () => {
    try {
      const [shell, action, debug] = await Promise.all([
        GetSetting('default_shell'),
        GetCloseAction(),
        GetSetting('allow_debug')
      ])

      if (shell) {
        defaultShell.value = shell
      }
      if (action) {
        closeAction.value = action
      }
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
