import { ref } from 'vue'
import { GetSetting, SetSetting } from '../../wailsjs/go/main/SettingService'

/**
 * 系统设置组合式函数
 * 负责默认 Shell 等系统级配置的加载和持久化
 */
export function useSettings() {
  const defaultShell = ref('cmd.exe')

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
   * 从后端加载已保存的设置项
   */
  const loadSettings = async () => {
    try {
      const shell = await GetSetting('default_shell')
      if (shell) {
        defaultShell.value = shell
      }
    } catch (err) {
      console.error('加载设置失败:', err)
    }
  }

  return {
    defaultShell,
    changeDefaultShell,
    loadSettings
  }
}
