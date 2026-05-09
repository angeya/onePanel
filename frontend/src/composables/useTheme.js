import { ref } from 'vue'
import { GetSetting, SetSetting } from '../../wailsjs/go/main/SettingService'

/**
 * 主题管理组合式函数
 * 负责主题的加载、切换和应用，持久化到后端设置
 */
export function useTheme() {
  const currentTheme = ref('dark')

  /**
   * 应用主题到 DOM
   * 通过修改 html 元素的 className 实现 CSS 变量切换
   */
  const applyTheme = (theme) => {
    currentTheme.value = theme
    const html = document.documentElement
    html.className = ''
    if (theme && theme !== 'dark') {
      html.classList.add(`theme-${theme}`)
    }
  }

  /**
   * 切换主题并持久化
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
   * 从后端加载已保存的主题设置
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
