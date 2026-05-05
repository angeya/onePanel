<template>
  <div class="apps-page">
    <div class="apps-header">
      <div class="header-left">
        <span class="page-title">我的应用</span>
        <el-tag v-if="serverStatus.running" type="success" size="small">
          服务运行中 :{{ serverStatus.port }}
        </el-tag>
        <el-tag v-else type="info" size="small">服务未启动</el-tag>
      </div>
      <div class="header-right">
        <el-button size="small" @click="showSettingsDialog" plain>
          <el-icon><Setting /></el-icon>
          设置
        </el-button>
        <el-button size="small" type="primary" @click="showImportDialog" plain>
          <el-icon><Upload /></el-icon>
          导入
        </el-button>
        <el-button size="small" @click="refreshApps" plain>
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <div class="apps-body" v-loading="loading">
      <div v-if="apps.length === 0 && !loading" class="apps-empty">
        <el-empty description="暂无应用，请先设置静态目录并导入应用">
          <el-button type="primary" @click="showSettingsDialog">设置静态目录</el-button>
        </el-empty>
      </div>

      <div v-else class="apps-grid">
        <div
          v-for="app in apps"
          :key="app.id"
          class="app-card"
          @click="openApp(app)"
        >
          <div class="app-icon">
            <img v-if="app.iconPath" :src="getIconUrl(app.iconPath)" alt="" />
            <el-icon v-else :size="40" color="#409eff"><Document /></el-icon>
          </div>
          <div class="app-info">
            <div class="app-name" :title="app.displayName">{{ app.displayName }}</div>
            <div class="app-dir" :title="app.dirName">{{ app.dirName }}</div>
          </div>
          <div class="app-actions" @click.stop>
            <el-dropdown trigger="click" @command="(cmd) => handleAppCommand(cmd, app)">
              <el-icon class="action-trigger"><MoreFilled /></el-icon>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="edit">编辑名称</el-dropdown-item>
                  <el-dropdown-item command="rename">修改目录名</el-dropdown-item>
                  <el-dropdown-item command="icon">上传图标</el-dropdown-item>
                  <el-dropdown-item command="export">导出</el-dropdown-item>
                  <el-dropdown-item command="delete" divided>
                    <span style="color: #f56c6c">删除</span>
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>
      </div>
    </div>

    <el-dialog v-model="settingsVisible" title="应用设置" width="500px" :close-on-click-modal="false">
      <el-form label-width="100px">
        <el-form-item label="静态目录">
          <div style="display: flex; gap: 8px; width: 100%">
            <el-input v-model="staticDir" placeholder="选择静态文件目录" readonly />
            <el-button @click="selectDirectory">选择</el-button>
          </div>
        </el-form-item>
        <el-form-item label="服务状态">
          <div style="display: flex; align-items: center; gap: 12px">
            <el-tag v-if="serverStatus.running" type="success" size="small">
              运行中，端口: {{ serverStatus.port }}
            </el-tag>
            <el-tag v-else type="info" size="small">未启动</el-tag>
            <el-button
              v-if="!serverStatus.running"
              size="small"
              type="primary"
              @click="startServer"
              :disabled="!staticDir"
            >
              启动服务
            </el-button>
            <el-button
              v-else
              size="small"
              type="danger"
              @click="stopServer"
            >
              停止服务
            </el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="settingsVisible = false">关闭</el-button>
        <el-button type="primary" @click="saveStaticDir">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="importVisible" title="导入应用" width="500px" :close-on-click-modal="false">
      <el-tabs v-model="importTab">
        <el-tab-pane label="导入 ZIP" name="zip">
          <el-form label-width="80px">
            <el-form-item label="ZIP 文件">
              <div style="display: flex; gap: 8px; width: 100%">
                <el-input v-model="importZipPath" placeholder="选择 ZIP 压缩包" readonly />
                <el-button @click="selectZipFile">选择</el-button>
              </div>
            </el-form-item>
          </el-form>
          <div style="text-align: right">
            <el-button type="primary" @click="doImportZip" :disabled="!importZipPath">导入</el-button>
          </div>
        </el-tab-pane>
        <el-tab-pane label="导入目录" name="dir">
          <el-form label-width="80px">
            <el-form-item label="应用目录">
              <div style="display: flex; gap: 8px; width: 100%">
                <el-input v-model="importDirPath" placeholder="选择包含 index.html 的目录" readonly />
                <el-button @click="selectImportDir">选择</el-button>
              </div>
            </el-form-item>
            <el-form-item label="应用名称">
              <el-input v-model="importAppName" placeholder="留空则使用目录名称" />
            </el-form-item>
          </el-form>
          <div style="text-align: right">
            <el-button type="primary" @click="doImportDir" :disabled="!importDirPath">导入</el-button>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-dialog>

    <el-dialog v-model="editNameVisible" title="编辑应用名称" width="400px" :close-on-click-modal="false">
      <el-form label-width="80px">
        <el-form-item label="应用名称">
          <el-input v-model="editNameValue" placeholder="请输入应用名称" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editNameVisible = false">取消</el-button>
        <el-button type="primary" @click="saveDisplayName">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="renameDirVisible" title="修改目录名称" width="400px" :close-on-click-modal="false">
      <el-form label-width="80px">
        <el-form-item label="目录名称">
          <el-input v-model="renameDirValue" placeholder="请输入新的目录名称" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="renameDirVisible = false">取消</el-button>
        <el-button type="primary" @click="saveDirName">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="appViewerVisible" :title="currentApp?.displayName || '应用'" width="90%" top="3vh" :close-on-click-modal="false" destroy-on-close>
      <div class="app-viewer-container">
        <iframe
          v-if="appViewerVisible && currentApp"
          :src="currentApp.fullUrl"
          class="app-iframe"
          sandbox="allow-scripts allow-same-origin allow-forms allow-popups"
        ></iframe>
      </div>
    </el-dialog>

    <input ref="iconInputRef" type="file" accept="image/png" style="display: none" @change="handleIconUpload" />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Setting, Upload, Refresh, Document, MoreFilled } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  GetStaticDir, SetStaticDir, GetServerStatus, StartServer, StopServer,
  GetApps, ScanApps, UpdateDisplayName, UpdateDirName, UploadIcon,
  DeleteApp, ExportApp, ImportZip, ImportDir
} from '../../wailsjs/go/main/AppService'
import { OpenDirectoryDialog, OpenFileDialog } from '../../wailsjs/go/main/App'

const apps = ref([])
const loading = ref(false)
const serverStatus = ref({ running: false, port: 0, dir: '' })

const settingsVisible = ref(false)
const staticDir = ref('')

const importVisible = ref(false)
const importTab = ref('zip')
const importZipPath = ref('')
const importDirPath = ref('')
const importAppName = ref('')

const editNameVisible = ref(false)
const editNameValue = ref('')
const editingAppId = ref(null)

const renameDirVisible = ref(false)
const renameDirValue = ref('')
const renamingAppId = ref(null)

const appViewerVisible = ref(false)
const currentApp = ref(null)

const iconInputRef = ref(null)
const iconUploadingAppId = ref(null)

/**
 * 加载应用列表
 */
const loadApps = async () => {
  loading.value = true
  try {
    apps.value = await GetApps()
  } catch (err) {
    ElMessage.error('加载应用列表失败: ' + err)
  } finally {
    loading.value = false
  }
}

/**
 * 刷新应用列表（重新扫描）
 */
const refreshApps = async () => {
  loading.value = true
  try {
    apps.value = await ScanApps()
    ElMessage.success('刷新成功')
  } catch (err) {
    ElMessage.error('刷新失败: ' + err)
  } finally {
    loading.value = false
  }
}

/**
 * 加载服务器状态
 */
const loadServerStatus = async () => {
  try {
    serverStatus.value = await GetServerStatus()
  } catch (err) {
    console.error('获取服务器状态失败:', err)
  }
}

/**
 * 显示设置对话框
 */
const showSettingsDialog = async () => {
  try {
    staticDir.value = await GetStaticDir() || ''
  } catch (err) {
    staticDir.value = ''
  }
  await loadServerStatus()
  settingsVisible.value = true
}

/**
 * 选择目录
 */
const selectDirectory = async () => {
  try {
    const dir = await OpenDirectoryDialog('选择静态文件目录')
    if (dir) {
      staticDir.value = dir
    }
  } catch (err) {
    console.error('选择目录失败:', err)
  }
}

/**
 * 保存静态目录设置
 */
const saveStaticDir = async () => {
  try {
    await SetStaticDir(staticDir.value)
    ElMessage.success('保存成功')
    await loadServerStatus()
    if (staticDir.value) {
      await refreshApps()
    }
    settingsVisible.value = false
  } catch (err) {
    ElMessage.error('保存失败: ' + err)
  }
}

/**
 * 启动服务器
 */
const startServer = async () => {
  try {
    await StartServer()
    await loadServerStatus()
    ElMessage.success('服务已启动')
  } catch (err) {
    ElMessage.error('启动失败: ' + err)
  }
}

/**
 * 停止服务器
 */
const stopServer = async () => {
  try {
    await StopServer()
    await loadServerStatus()
    ElMessage.success('服务已停止')
  } catch (err) {
    ElMessage.error('停止失败: ' + err)
  }
}

/**
 * 显示导入对话框
 */
const showImportDialog = () => {
  importZipPath.value = ''
  importDirPath.value = ''
  importAppName.value = ''
  importVisible.value = true
}

/**
 * 选择 ZIP 文件
 */
const selectZipFile = async () => {
  try {
    const path = await OpenFileDialog('选择 ZIP 文件', 'ZIP 文件 (*.zip)')
    if (path) {
      importZipPath.value = path
    }
  } catch (err) {
    console.error('选择文件失败:', err)
  }
}

/**
 * 选择导入目录
 */
const selectImportDir = async () => {
  try {
    const dir = await OpenDirectoryDialog('选择应用目录')
    if (dir) {
      importDirPath.value = dir
    }
  } catch (err) {
    console.error('选择目录失败:', err)
  }
}

/**
 * 执行 ZIP 导入
 */
const doImportZip = async () => {
  try {
    await ImportZip(importZipPath.value)
    ElMessage.success('导入成功')
    importVisible.value = false
    await refreshApps()
  } catch (err) {
    ElMessage.error('导入失败: ' + err)
  }
}

/**
 * 执行目录导入
 */
const doImportDir = async () => {
  try {
    await ImportDir(importDirPath.value, importAppName.value)
    ElMessage.success('导入成功')
    importVisible.value = false
    await refreshApps()
  } catch (err) {
    ElMessage.error('导入失败: ' + err)
  }
}

/**
 * 打开应用
 */
const openApp = (app) => {
  if (!serverStatus.value.running) {
    ElMessage.warning('请先启动静态服务器')
    return
  }
  currentApp.value = {
    ...app,
    fullUrl: `http://127.0.0.1:${serverStatus.value.port}${app.entryUrl}`
  }
  appViewerVisible.value = true
}

/**
 * 获取图标 URL
 */
const getIconUrl = (iconPath) => {
  if (!iconPath || !serverStatus.value.running) return ''
  const app = apps.value.find(a => a.iconPath === iconPath)
  if (!app) return ''
  const dir = app.entryUrl.replace('/index.html', '')
  return `http://127.0.0.1:${serverStatus.value.port}${dir}/icon.png`
}

/**
 * 处理应用操作命令
 */
const handleAppCommand = (command, app) => {
  switch (command) {
    case 'edit':
      editingAppId.value = app.id
      editNameValue.value = app.displayName
      editNameVisible.value = true
      break
    case 'rename':
      renamingAppId.value = app.id
      renameDirValue.value = app.dirName
      renameDirVisible.value = true
      break
    case 'icon':
      iconUploadingAppId.value = app.id
      iconInputRef.value?.click()
      break
    case 'export':
      doExportApp(app)
      break
    case 'delete':
      doDeleteApp(app)
      break
  }
}

/**
 * 保存显示名称
 */
const saveDisplayName = async () => {
  try {
    await UpdateDisplayName(editingAppId.value, editNameValue.value)
    ElMessage.success('修改成功')
    editNameVisible.value = false
    await loadApps()
  } catch (err) {
    ElMessage.error('修改失败: ' + err)
  }
}

/**
 * 保存目录名称
 */
const saveDirName = async () => {
  try {
    await UpdateDirName(renamingAppId.value, renameDirValue.value)
    ElMessage.success('修改成功')
    renameDirVisible.value = false
    await loadApps()
  } catch (err) {
    ElMessage.error('修改失败: ' + err)
  }
}

/**
 * 处理图标上传
 */
const handleIconUpload = async (event) => {
  const file = event.target.files[0]
  if (!file) return

  const reader = new FileReader()
  reader.onload = async (e) => {
    const data = new Uint8Array(e.target.result)
    try {
      await UploadIcon(iconUploadingAppId.value, Array.from(data))
      ElMessage.success('图标上传成功')
      await loadApps()
    } catch (err) {
      ElMessage.error('上传失败: ' + err)
    }
  }
  reader.readAsArrayBuffer(file)
  event.target.value = ''
}

/**
 * 导出应用
 */
const doExportApp = async (app) => {
  try {
    const zipPath = await ExportApp(app.id)
    ElMessage.success(`已导出到: ${zipPath}`)
  } catch (err) {
    ElMessage.error('导出失败: ' + err)
  }
}

/**
 * 删除应用
 */
const doDeleteApp = async (app) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除应用 "${app.displayName}" 吗？此操作将同时删除应用文件，不可恢复。`,
      '确认删除',
      { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' }
    )
    await DeleteApp(app.id)
    ElMessage.success('删除成功')
    await loadApps()
  } catch (err) {
    if (err !== 'cancel') {
      ElMessage.error('删除失败: ' + err)
    }
  }
}

onMounted(async () => {
  await loadServerStatus()
  await loadApps()
})
</script>

<style scoped>
.apps-page {
  display: flex;
  flex-direction: column;
  height: 100%;
  background-color: #252526;
}

.apps-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  background-color: #2d2d2d;
  border-bottom: 1px solid #3d3d3d;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.page-title {
  font-size: 16px;
  font-weight: 600;
  color: #e5e5e5;
}

.header-right {
  display: flex;
  gap: 8px;
}

.apps-body {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}

.apps-empty {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
}

.apps-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 16px;
}

.app-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background-color: #2d2d2d;
  border: 1px solid #3d3d3d;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  position: relative;
}

.app-card:hover {
  background-color: #363636;
  border-color: #409eff;
  transform: translateY(-1px);
}

.app-icon {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  background-color: #1e1e1e;
  flex-shrink: 0;
  overflow: hidden;
}

.app-icon img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.app-info {
  flex: 1;
  min-width: 0;
}

.app-name {
  font-size: 14px;
  font-weight: 500;
  color: #e5e5e5;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.app-dir {
  font-size: 12px;
  color: #888;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-top: 2px;
}

.app-actions {
  position: absolute;
  top: 8px;
  right: 8px;
  opacity: 0;
  transition: opacity 0.2s;
}

.app-card:hover .app-actions {
  opacity: 1;
}

.action-trigger {
  cursor: pointer;
  color: #a0a0a0;
  padding: 4px;
  border-radius: 4px;
}

.action-trigger:hover {
  color: #e5e5e5;
  background-color: #4d4d4d;
}

.app-viewer-container {
  width: 100%;
  height: 70vh;
  background-color: #fff;
  border-radius: 4px;
  overflow: hidden;
}

.app-iframe {
  width: 100%;
  height: 100%;
  border: none;
}
</style>
