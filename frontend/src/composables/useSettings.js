import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { GetSetting, SetSetting, GetGlobalHotkey, SetGlobalHotkey } from '../../wailsjs/go/main/SettingService'
import { GetCloseAction, SetCloseAction } from '../../wailsjs/go/main/App'

/**
 * useSettings 管理系统设置相关状态。
 * 统一处理默认终端、关闭行为和全局快捷键的读取与持久化。
 */
export function useSettings() {
  const defaultShell = ref('cmd.exe')
  const closeAction = ref('ask')
  const hotkeyConfig = ref({ modifiers: ['ctrl', 'alt'], key: 'O' })

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
    if (settings.global_hotkey) {
      try {
        hotkeyConfig.value = JSON.parse(settings.global_hotkey)
      } catch {
        hotkeyConfig.value = { modifiers: ['ctrl', 'alt'], key: 'O' }
      }
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
   * changeGlobalHotkey 保存全局快捷键配置。
   */
  const changeGlobalHotkey = async (config) => {
    try {
      await SetGlobalHotkey(config)
      hotkeyConfig.value = config
      ElMessage.success('快捷键已保存，重启后生效')
    } catch (err) {
      console.error('保存全局快捷键失败:', err)
      ElMessage.error('保存快捷键失败: ' + err)
    }
  }

  /**
   * loadSettings 兼容非启动时机的设置刷新。
   * 例如设置窗口重新打开时，需要再次从后端读取最新值。
   */
  const loadSettings = async () => {
    try {
      const [shell, action, hotkey] = await Promise.all([
        GetSetting('default_shell'),
        GetCloseAction(),
        GetGlobalHotkey()
      ])

      if (shell) {
        defaultShell.value = shell
      }
      if (action) {
        closeAction.value = action
      }
      if (hotkey) {
        hotkeyConfig.value = hotkey
      }
    } catch (err) {
      console.error('加载设置失败:', err)
    }
  }

  return {
    defaultShell,
    closeAction,
    hotkeyConfig,
    applyBootstrapSettings,
    changeDefaultShell,
    changeCloseAction,
    changeGlobalHotkey,
    loadSettings
  }
}
