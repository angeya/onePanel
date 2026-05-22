import { ref } from 'vue'
import { GetSetting, SetSetting } from '../../wailsjs/go/main/SettingService'

export function useTheme() {
  const currentTheme = ref('dark')

  const applyTheme = (theme) => {
    currentTheme.value = theme || 'dark'
    const html = document.documentElement
    html.className = ''
    if (currentTheme.value !== 'dark') {
      html.classList.add(`theme-${currentTheme.value}`)
    }
  }

  const changeTheme = async (theme) => {
    try {
      await SetSetting('theme', theme)
      applyTheme(theme)
    } catch (err) {
      console.error('保存主题失败:', err)
    }
  }

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
