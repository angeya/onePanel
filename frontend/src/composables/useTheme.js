import { ref } from 'vue'
import { GetSetting, SetSetting } from '../../wailsjs/go/main/SettingService'

/**
 * useTheme 负责维护当前主题及其在 DOM 上的应用。
 */
export function useTheme() {
  const currentTheme = ref('dark')

  /**
   * applyTheme 将主题同步到 html 根节点。
   * 通过 class 切换 CSS 变量，保证全局样式即时生效。
   */
  const applyTheme = (theme) => {
    currentTheme.value = theme || 'dark'
    const html = document.documentElement
    html.className = ''
    if (currentTheme.value !== 'dark') {
      html.classList.add(`theme-${currentTheme.value}`)
    }
  }

  /**
   * changeTheme 保存并应用新的主题配置。
   */
  const changeTheme = async (theme) => {
    try {
      await SetSetting('theme', theme)
      applyTheme(theme)
    } catch (err) {
      console.error('保存主题失败:', err)
    }
  }

  /**
   * loadTheme 在需要单独刷新主题时从后端重新读取配置。
   */
  const loadTheme = async () => {
    try {
      const theme = await GetSetting('theme')
      if (theme) {
        applyTheme(theme)
      }
    } catch (err) {
      console.error('加载主题失败:', err)
    }
  }

  return {
    currentTheme,
    applyTheme,
    changeTheme,
    loadTheme
  }
}
