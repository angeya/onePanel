import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { GetSetting, SetSetting } from '../../wailsjs/go/main/SettingService'
import { SetStaticDir } from '../../wailsjs/go/main/AppService'
import { OpenDirectoryDialog } from '../../wailsjs/go/main/App'

/**
 * 系统设置组合式函数
 * 负责默认 Shell、自定义应用目录等系统级配置的加载和持久化
 */
export function useSettings() {
  const defaultShell = ref('cmd.exe')
  const customAppDir = ref('')

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
   * 选择自定义应用目录
   */
  const selectAppDir = async () => {
    try {
      const dir = await OpenDirectoryDialog('选择应用目录')
      if (dir) customAppDir.value = dir
    } catch (err) {
      console.error('选择目录失败:', err)
      ElMessage.warning('目录选择对话框打开失败，请手动输入路径')
    }
  }

  /**
   * 保存自定义应用目录
   * 清空表示使用默认目录（exe 同级 apps 目录）
   */
  const saveAppDir = async () => {
    try {
      await SetStaticDir(customAppDir.value)
      ElMessage.success('保存成功')
    } catch (err) {
      ElMessage.error('保存失败: ' + err)
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

  /**
   * 加载自定义应用目录配置
   * 从数据库读取原始配置值，为空则表示使用默认目录
   */
  const loadAppDir = async () => {
    try {
      const dir = await GetSetting('static_dir')
      customAppDir.value = dir || ''
    } catch (err) {
      customAppDir.value = ''
    }
  }

  return {
    defaultShell,
    customAppDir,
    changeDefaultShell,
    selectAppDir,
    saveAppDir,
    loadSettings,
    loadAppDir
  }
}
