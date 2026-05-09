import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  GetStaticDir, SetStaticDir, GetServerStatus,
  GetApps, ScanApps, UpdateDisplayName, UpdateDirName, UploadIcon,
  DeleteApp, ExportApp, ImportZip, ImportDir, OpenApp as OpenAppService,
  CreateWebApp, UpdateWebApp
} from '../../wailsjs/go/main/AppService'
import { OpenDirectoryDialog, OpenFileDialog } from '../../wailsjs/go/main/App'

/**
 * 我的应用服务组合式函数
 * 负责应用列表、静态服务器、导入导出等逻辑
 */
export function useAppService(closeAppTab, appDialogsRef) {
  const apps = ref([])
  const appsLoading = ref(false)
  const serverStatus = ref({ running: false, port: 0, dir: '' })

  const appSettingsVisible = ref(false)
  const staticDir = ref('')

  const appImportVisible = ref(false)
  const appImportTab = ref('zip')
  const importZipPath = ref('')
  const importDirPath = ref('')
  const importAppName = ref('')

  const appEditNameVisible = ref(false)
  const appEditNameValue = ref('')
  const editingAppId = ref(null)

  const appRenameDirVisible = ref(false)
  const appRenameDirValue = ref('')
  const renamingAppId = ref(null)

  const iconUploadingAppId = ref(null)

  const webAppDialogVisible = ref(false)
  const isEditingWebApp = ref(false)
  const editingWebAppId = ref(null)
  const webAppForm = ref({ name: '', url: '' })

  /**
   * 加载应用列表
   */
  const loadApps = async () => {
    appsLoading.value = true
    try {
      apps.value = await GetApps()
    } catch (err) {
      ElMessage.error('加载应用列表失败: ' + err)
    } finally {
      appsLoading.value = false
    }
  }

  /**
   * 刷新应用列表
   */
  const refreshApps = async () => {
    appsLoading.value = true
    try {
      apps.value = await ScanApps()
      ElMessage.success('刷新成功')
    } catch (err) {
      ElMessage.error('刷新失败: ' + err)
    } finally {
      appsLoading.value = false
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
   * 获取应用图标 URL
   */
  const getAppIconUrl = (app) => {
    if (!app.iconPath || !serverStatus.value.running) return ''
    const dir = app.entryUrl.replace('/index.html', '')
    return `http://127.0.0.1:${serverStatus.value.port}${dir}/icon.png`
  }

  /**
   * 打开应用 - 调用后端自动启动静态服务并获取 URL
   */
  const openApp = async (app, addAppTab) => {
    try {
      const result = await OpenAppService(app.id)
      await loadServerStatus()
      addAppTab(app.id, result.name, result.url)
    } catch (err) {
      ElMessage.error('打开应用失败: ' + err)
    }
  }

  /**
   * 显示应用设置
   */
  const showAppSettings = async () => {
    try {
      staticDir.value = await GetStaticDir() || ''
    } catch (err) {
      staticDir.value = ''
    }
    await loadServerStatus()
    appSettingsVisible.value = true
  }

  /**
   * 选择目录
   */
  const selectDirectory = async () => {
    try {
      const dir = await OpenDirectoryDialog('选择静态文件目录')
      if (dir) staticDir.value = dir
    } catch (err) {
      console.error('选择目录失败:', err)
      ElMessage.warning('目录选择对话框打开失败，请手动输入路径')
    }
  }

  /**
   * 保存静态目录
   */
  const saveStaticDir = async () => {
    try {
      await SetStaticDir(staticDir.value)
      ElMessage.success('保存成功')
      if (staticDir.value) await refreshApps()
      appSettingsVisible.value = false
    } catch (err) {
      ElMessage.error('保存失败: ' + err)
    }
  }

  /**
   * 显示导入对话框
   */
  const showAppImport = () => {
    importZipPath.value = ''
    importDirPath.value = ''
    importAppName.value = ''
    appImportVisible.value = true
  }

  /**
   * 选择 ZIP 文件
   */
  const selectZipFile = async () => {
    try {
      const path = await OpenFileDialog('选择 ZIP 文件', 'ZIP 文件 (*.zip)')
      if (path) importZipPath.value = path
    } catch (err) {
      console.error('选择文件失败:', err)
      ElMessage.warning('文件选择对话框打开失败，请手动输入路径')
    }
  }

  /**
   * 选择导入目录
   */
  const selectImportDir = async () => {
    try {
      const dir = await OpenDirectoryDialog('选择应用目录')
      if (dir) importDirPath.value = dir
    } catch (err) {
      console.error('选择目录失败:', err)
      ElMessage.warning('目录选择对话框打开失败，请手动输入路径')
    }
  }

  /**
   * 执行 ZIP 导入
   */
  const doImportZip = async () => {
    try {
      await ImportZip(importZipPath.value)
      ElMessage.success('导入成功')
      appImportVisible.value = false
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
      appImportVisible.value = false
      await refreshApps()
    } catch (err) {
      ElMessage.error('导入失败: ' + err)
    }
  }

  /**
   * 显示新增网页应用对话框
   */
  const showAddWebAppDialog = () => {
    isEditingWebApp.value = false
    editingWebAppId.value = null
    webAppForm.value = { name: '', url: '' }
    webAppDialogVisible.value = true
  }

  /**
   * 显示编辑网页应用对话框
   */
  const showEditWebAppDialog = (app) => {
    isEditingWebApp.value = true
    editingWebAppId.value = app.id
    webAppForm.value = { name: app.displayName, url: app.entryUrl }
    webAppDialogVisible.value = true
  }

  /**
   * 保存网页应用（新增或编辑）
   */
  const saveWebApp = async () => {
    if (!webAppForm.value.name) {
      ElMessage.warning('请输入应用名称')
      return
    }
    if (!webAppForm.value.url) {
      ElMessage.warning('请输入应用地址')
      return
    }

    try {
      if (isEditingWebApp.value) {
        await UpdateWebApp(editingWebAppId.value, webAppForm.value.name, webAppForm.value.url)
        ElMessage.success('修改成功')
      } else {
        await CreateWebApp(webAppForm.value.name, webAppForm.value.url)
        ElMessage.success('添加成功')
      }
      webAppDialogVisible.value = false
      await loadApps()
    } catch (err) {
      ElMessage.error('保存失败: ' + err)
    }
  }

  /**
   * 更新网页应用表单字段
   */
  const updateWebAppForm = ({ key, value }) => {
    webAppForm.value[key] = value
  }

  /**
   * 处理应用操作命令
   */
  const handleAppCmd = (command, app) => {
    switch (command) {
      case 'edit':
        if (app.appType === 'web') {
          showEditWebAppDialog(app)
        } else {
          editingAppId.value = app.id
          appEditNameValue.value = app.displayName
          appEditNameVisible.value = true
        }
        break
      case 'rename':
        renamingAppId.value = app.id
        appRenameDirValue.value = app.dirName
        appRenameDirVisible.value = true
        break
      case 'icon':
        iconUploadingAppId.value = app.id
        if (appDialogsRef && appDialogsRef.value) {
          appDialogsRef.value.triggerIconInput()
        }
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
   * 保存应用显示名称
   */
  const saveAppDisplayName = async () => {
    try {
      await UpdateDisplayName(editingAppId.value, appEditNameValue.value)
      ElMessage.success('修改成功')
      appEditNameVisible.value = false
      await loadApps()
    } catch (err) {
      ElMessage.error('修改失败: ' + err)
    }
  }

  /**
   * 保存应用目录名称
   */
  const saveAppDirName = async () => {
    try {
      await UpdateDirName(renamingAppId.value, appRenameDirValue.value)
      ElMessage.success('修改成功')
      appRenameDirVisible.value = false
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
    const isWebApp = app.appType === 'web'
    const message = isWebApp
      ? `确定要删除网页应用 "${app.displayName}" 吗？`
      : `确定要删除应用 "${app.displayName}" 吗？此操作将同时删除应用文件，不可恢复。`

    try {
      await ElMessageBox.confirm(message, '确认删除', {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning'
      })
      await DeleteApp(app.id)
      ElMessage.success('删除成功')
      if (closeAppTab) closeAppTab(app.id)
      await loadApps()
    } catch (err) {
      if (err !== 'cancel') {
        ElMessage.error('删除失败: ' + err)
      }
    }
  }

  return {
    apps,
    appsLoading,
    serverStatus,
    appSettingsVisible,
    staticDir,
    appImportVisible,
    appImportTab,
    importZipPath,
    importDirPath,
    importAppName,
    appEditNameVisible,
    appEditNameValue,
    appRenameDirVisible,
    appRenameDirValue,
    webAppDialogVisible,
    isEditingWebApp,
    webAppForm,
    loadApps,
    refreshApps,
    loadServerStatus,
    getAppIconUrl,
    openApp,
    showAppSettings,
    selectDirectory,
    saveStaticDir,
    showAppImport,
    selectZipFile,
    selectImportDir,
    doImportZip,
    doImportDir,
    showAddWebAppDialog,
    showEditWebAppDialog,
    saveWebApp,
    updateWebAppForm,
    handleAppCmd,
    saveAppDisplayName,
    saveAppDirName,
    handleIconUpload
  }
}
