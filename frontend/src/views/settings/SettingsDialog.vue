<template>
  <el-dialog
    v-model="visible"
    title="系统设置"
    width="520px"
    :close-on-click-modal="false"
    top="5vh"
  >
    <div class="settings-content">
      <div class="setting-section">
        <div class="section-title">主题设置</div>
        <div class="theme-grid">
          <div
            v-for="theme in themes"
            :key="theme.key"
            class="theme-card"
            :class="{ active: currentTheme === theme.key }"
            @click="changeTheme(theme.key)"
          >
            <div class="theme-preview" :style="theme.previewStyle">
              <div class="preview-sidebar" :style="{ backgroundColor: theme.sidebarBg }">
                <div class="preview-dot" v-for="i in 3" :key="i"></div>
              </div>
              <div class="preview-main" :style="{ backgroundColor: theme.mainBg }">
                <div class="preview-line" :style="{ backgroundColor: theme.lineColor }" v-for="i in 3" :key="i"></div>
              </div>
            </div>
            <div class="theme-name">{{ theme.label }}</div>
          </div>
        </div>
      </div>

      <div class="setting-section">
        <div class="section-title">默认终端</div>
        <el-radio-group v-model="defaultShell" @change="saveDefaultShell">
          <el-radio value="cmd.exe">CMD</el-radio>
          <el-radio value="powershell.exe">PowerShell</el-radio>
        </el-radio-group>
      </div>

      <div class="setting-section">
        <div class="section-title">关闭行为</div>
        <div class="close-action-desc">点击窗口右上角关闭按钮时的行为</div>
        <el-radio-group v-model="currentCloseAction" @change="saveCloseAction">
          <el-radio value="ask">每次提问</el-radio>
          <el-radio value="tray">最小化到托盘</el-radio>
          <el-radio value="close">直接退出应用</el-radio>
        </el-radio-group>
      </div>

      <div class="setting-section">
        <div class="section-title">调试设置</div>
        <div class="setting-row">
          <div class="setting-row-info">
            <span class="setting-row-label">允许调试</span>
            <span class="setting-row-desc">开启后允许在应用页面中右键弹出浏览器菜单</span>
          </div>
          <el-switch v-model="allowDebug" @change="saveAllowDebug" />
        </div>
      </div>

      <div class="setting-section">
        <div class="section-title">日志管理</div>
        <div class="setting-row">
          <div class="setting-row-info">
            <span class="setting-row-label">应用日志</span>
            <span class="setting-row-desc">查看应用运行日志文件</span>
          </div>
          <el-button size="small" @click="openLogsDir">查看日志</el-button>
        </div>
      </div>

      <div class="setting-section">
        <div class="section-title">版本信息</div>
        <div class="info-row">
          <span class="info-label">当前版本</span>
          <span class="info-value">v0.0.1</span>
        </div>
      </div>

      <div class="setting-section">
        <div class="section-title">联系我们</div>
        <div class="info-row">
          <span class="info-label">邮箱</span>
          <span class="info-value email" @click="copyEmail">1571858518@qq.com</span>
        </div>
      </div>
    </div>
  </el-dialog>
</template>

<script setup>
import { ref, inject } from 'vue'
import { ElMessage } from 'element-plus'
import { GetSetting, SetSetting } from '../../../wailsjs/go/main/SettingService'
import { GetCloseAction, SetCloseAction, OpenLogsDir } from '../../../wailsjs/go/main/App'

const emit = defineEmits(['themeChange', 'shellChange', 'closeActionChange', 'allowDebugChange'])

const visible = ref(false)
const currentTheme = ref('dark')
const defaultShell = ref('cmd.exe')
const currentCloseAction = ref('ask')
const allowDebug = inject('allowDebug')
const changeAllowDebug = inject('changeAllowDebug')

const themes = [
  {
    key: 'dark',
    label: '深色主题',
    sidebarBg: '#1e1e1e',
    mainBg: '#252526',
    lineColor: '#3d3d3d',
    previewStyle: { border: '2px solid #3d3d3d' }
  },
  {
    key: 'light',
    label: '浅色主题',
    sidebarBg: '#f5f5f5',
    mainBg: '#ffffff',
    lineColor: '#e5e5e5',
    previewStyle: { border: '2px solid #dcdfe6' }
  },
  {
    key: 'blue',
    label: '蓝色主题',
    sidebarBg: '#0d1b2a',
    mainBg: '#1b2838',
    lineColor: '#2a3f5f',
    previewStyle: { border: '2px solid #2a3f5f' }
  },
  {
    key: 'green',
    label: '绿色主题',
    sidebarBg: '#1a2e1a',
    mainBg: '#243024',
    lineColor: '#3a5a3a',
    previewStyle: { border: '2px solid #3a5a3a' }
  }
]

const changeTheme = async (key) => {
  currentTheme.value = key
  try {
    await SetSetting('theme', key)
    emit('themeChange', key)
  } catch (err) {
    ElMessage.error('保存主题失败: ' + err)
  }
}

const saveDefaultShell = async (val) => {
  try {
    await SetSetting('default_shell', val)
    emit('shellChange', val)
  } catch (err) {
    ElMessage.error('保存默认终端失败: ' + err)
  }
}

const saveCloseAction = async (val) => {
  try {
    await SetCloseAction(val)
    currentCloseAction.value = val
    emit('closeActionChange', val)
  } catch (err) {
    ElMessage.error('保存关闭行为设置失败: ' + err)
  }
}

const saveAllowDebug = async (val) => {
  try {
    await changeAllowDebug(val)
    emit('allowDebugChange', val)
  } catch (err) {
    allowDebug.value = !val
    ElMessage.error('保存调试开关设置失败: ' + err)
  }
}

const openLogsDir = async () => {
  try {
    await OpenLogsDir()
  } catch (err) {
    ElMessage.error('打开日志目录失败: ' + err)
  }
}

const copyEmail = async () => {
  try {
    await navigator.clipboard.writeText('1571858518@qq.com')
    ElMessage.success('邮箱已复制到剪贴板')
  } catch {
    ElMessage.info('邮箱: 1571858518@qq.com')
  }
}

const loadSettings = async () => {
  try {
    const theme = await GetSetting('theme')
    if (theme) currentTheme.value = theme

    const shell = await GetSetting('default_shell')
    if (shell) defaultShell.value = shell

    const action = await GetCloseAction()
    if (action) currentCloseAction.value = action
  } catch (err) {
    console.error('加载设置失败:', err)
  }
}

const open = async () => {
  await loadSettings()
  visible.value = true
}

defineExpose({ loadSettings, open })
</script>

<style scoped>
.settings-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
  max-height: 70vh;
  overflow-y: auto;
  padding: 4px 12px 4px 4px;
}

.settings-content::-webkit-scrollbar {
  width: 6px;
}

.settings-content::-webkit-scrollbar-track {
  background: transparent;
}

.settings-content::-webkit-scrollbar-thumb {
  background-color: var(--scrollbar-thumb);
  border-radius: 3px;
}

.settings-content::-webkit-scrollbar-thumb:hover {
  background-color: var(--text-dimmed);
}

.setting-section {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  padding-bottom: 4px;
  border-bottom: 1px solid var(--border-light);
}

.theme-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 10px;
}

.theme-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 8px;
  border-radius: 8px;
  cursor: pointer;
  border: 2px solid transparent;
  transition: all 0.2s;
}

.theme-card:hover {
  background-color: rgba(255, 255, 255, 0.05);
}

.theme-card.active {
  border-color: var(--accent);
  background-color: rgba(64, 158, 255, 0.1);
}

.theme-preview {
  width: 100%;
  height: 48px;
  border-radius: 4px;
  display: flex;
  overflow: hidden;
}

.preview-sidebar {
  width: 30%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 3px;
  padding: 4px;
}

.preview-dot {
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background-color: rgba(255, 255, 255, 0.3);
}

.preview-main {
  flex: 1;
  height: 100%;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 3px;
  padding: 4px 6px;
}

.preview-line {
  height: 3px;
  border-radius: 1px;
  width: 80%;
}

.theme-name {
  font-size: 11px;
  color: var(--text-muted);
  text-align: center;
}

.theme-card.active .theme-name {
  color: var(--accent);
}

.info-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.info-label {
  font-size: 13px;
  color: var(--text-muted);
}

.info-value {
  font-size: 13px;
  color: var(--text-primary);
}

.email {
  cursor: pointer;
  color: var(--accent);
}

.email:hover {
  text-decoration: underline;
}

.close-action-desc {
  font-size: 12px;
  color: var(--text-muted);
  margin-bottom: 4px;
}

.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.setting-row-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.setting-row-label {
  font-size: 13px;
  color: var(--text-primary);
}

.setting-row-desc {
  font-size: 12px;
  color: var(--text-muted);
}
</style>
