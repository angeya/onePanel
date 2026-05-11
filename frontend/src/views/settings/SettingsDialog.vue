<template>
  <el-dialog
    :model-value="visible"
    @update:model-value="$emit('update:visible', $event)"
    title="系统设置"
    width="480px"
    :close-on-click-modal="false"
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
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { GetSetting, SetSetting } from '../../../wailsjs/go/main/SettingService'

const props = defineProps({
  visible: { type: Boolean, default: false },
  theme: { type: String, default: 'dark' },
  shell: { type: String, default: 'cmd.exe' }
})

const emit = defineEmits(['update:visible', 'themeChange', 'shellChange'])

const currentTheme = ref(props.theme)
const defaultShell = ref(props.shell)

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

/**
 * 切换主题
 */
const changeTheme = async (key) => {
  currentTheme.value = key
  try {
    await SetSetting('theme', key)
    emit('themeChange', key)
  } catch (err) {
    ElMessage.error('保存主题失败: ' + err)
  }
}

/**
 * 保存默认终端
 */
const saveDefaultShell = async (val) => {
  try {
    await SetSetting('default_shell', val)
    emit('shellChange', val)
  } catch (err) {
    ElMessage.error('保存默认终端失败: ' + err)
  }
}

/**
 * 复制邮箱地址
 */
const copyEmail = async () => {
  try {
    await navigator.clipboard.writeText('1571858518@qq.com')
    ElMessage.success('邮箱已复制到剪贴板')
  } catch {
    ElMessage.info('邮箱: 1571858518@qq.com')
  }
}

/**
 * 加载设置
 */
const loadSettings = async () => {
  try {
    const theme = await GetSetting('theme')
    if (theme) currentTheme.value = theme

    const shell = await GetSetting('default_shell')
    if (shell) defaultShell.value = shell
  } catch (err) {
    console.error('加载设置失败:', err)
  }
}

/**
 * 打开时同步最新设置
 */
const handleOpen = () => {
  currentTheme.value = props.theme
  defaultShell.value = props.shell
}

defineExpose({ loadSettings, handleOpen })
</script>

<style scoped>
.settings-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
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
</style>
