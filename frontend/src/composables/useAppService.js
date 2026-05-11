import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  GetServerStatus,
  ScanApps, UpdateDisplayName,
  DeleteApp, ExportApp, ImportZip, ImportDir, ImportHtml, OpenApp as OpenAppService,
  BatchExportApps, CreateWebApp, UpdateWebApp
} from '../../wailsjs/go/main/AppService'
import { OpenDirectoryDialog, OpenFileDialog, SaveFileDialog } from '../../wailsjs/go/main/App'

/**
 * 我的应用服务组合式函数
 * 负责应用列表、静态服务器、导入导出等逻辑
 */
export function useAppService(closeAppTab) {
  const apps = ref([])
  const appsLoading = ref(false)
  const serverStatus = ref({ running: false, port: 0, dir: '' })

  const appImportVisible = ref(false)
  const appImportTab = ref('zip')
  const importZipPath = ref('')
  const importDirPath = ref('')
  const importDirName = ref('')
  const importHtmlPath = ref('')
  const importHtmlName = ref('')

  const appEditNameVisible = ref(false)
  const appEditNameValue = ref('')
  const editingAppId = ref(null)

  const webAppDialogVisible = ref(false)
  const isEditingWebApp = ref(false)
  const editingWebAppId = ref(null)
  const webAppForm = ref({ name: '', url: '' })

  const batchExportVisible = ref(false)
  const batchExportSelected = ref([])

  /**
   * 加载应用列表
   */
  const loadApps = async () => {
    appsLoading.value = true
    try {
      apps.value = await ScanApps()
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
   * 显示导入对话框
   */
  const showAppImport = () => {
    importZipPath.value = ''
    importDirPath.value = ''
    importDirName.value = ''
    importHtmlPath.value = ''
    importHtmlName.value = ''
    appImportVisible.value = true
  }

  /**
   * 选择 ZIP 文件
   */
  const selectZipFile = async () => {
    try {
      const path = await OpenFileDialog('选择 ZIP 文件', 'ZIP 文件 (*.zip)|*.zip')
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
   * 选择 HTML 文件
   */
  const selectHtmlFile = async () => {
    try {
      const path = await OpenFileDialog('选择 HTML 文件', 'HTML 文件 (*.html;*.htm)|*.html;*.htm')
      if (path) importHtmlPath.value = path
    } catch (err) {
      console.error('选择文件失败:', err)
      ElMessage.warning('文件选择对话框打开失败，请手动输入路径')
    }
  }

  /**
   * 执行 ZIP 导入
   */
  const doImportZip = async () => {
    if (!importZipPath.value) {
      ElMessage.warning('请选择 ZIP 文件')
      return
    }
    try {
      const skipped = await ImportZip(importZipPath.value)
      appImportVisible.value = false
      await refreshApps()
      if (skipped) {
        ElMessage.warning(`以下应用已存在，已跳过: ${skipped}`)
      } else {
        ElMessage.success('导入成功')
      }
    } catch (err) {
      ElMessage.error('导入失败: ' + err)
    }
  }

  /**
   * 执行目录导入
   */
  const doImportDir = async () => {
    if (!importDirPath.value) {
      ElMessage.warning('请选择应用目录')
      return
    }
    try {
      await ImportDir(importDirPath.value, importDirName.value)
      ElMessage.success('导入成功')
      appImportVisible.value = false
      await refreshApps()
    } catch (err) {
      ElMessage.error('导入失败: ' + err)
    }
  }

  /**
   * 执行 HTML 文件导入
   */
  const doImportHtml = async () => {
    if (!importHtmlPath.value) {
      ElMessage.warning('请选择 HTML 文件')
      return
    }
    if (!importHtmlName.value) {
      ElMessage.warning('请输入应用名称')
      return
    }
    try {
      await ImportHtml(importHtmlPath.value, importHtmlName.value)
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
   * 导出单个应用
   * 弹出保存文件对话框让用户选择保存路径
   * 支持静态应用和网页应用
   */
  const doExportApp = async (app) => {
    try {
      const baseName = app.appType === 'web' ? app.displayName : app.dirName
      const defaultName = `${baseName}_${new Date().toISOString().replace(/[-:T]/g, '').slice(0, 14)}.zip`
      const savePath = await SaveFileDialog('导出应用', defaultName, 'ZIP 文件 (*.zip)|*.zip')
      if (!savePath) return

      await ExportApp(app.id, savePath)
      ElMessage.success(`已导出到: ${savePath}`)
    } catch (err) {
      ElMessage.error('导出失败: ' + err)
    }
  }

  /**
   * 显示批量导出对话框
   */
  const showBatchExport = () => {
    if (apps.value.length === 0) {
      ElMessage.info('没有可导出的应用')
      return
    }
    batchExportSelected.value = []
    batchExportVisible.value = true
  }

  /**
   * 执行批量导出
   */
  const doBatchExport = async () => {
    if (batchExportSelected.value.length === 0) {
      ElMessage.warning('请至少选择一个应用')
      return
    }

    try {
      const defaultName = `apps_export_${new Date().toISOString().replace(/[-:T]/g, '').slice(0, 14)}.zip`
      const savePath = await SaveFileDialog('批量导出应用', defaultName, 'ZIP 文件 (*.zip)|*.zip')
      if (!savePath) return

      await BatchExportApps(batchExportSelected.value, savePath)
      ElMessage.success(`已导出 ${batchExportSelected.value.length} 个应用到: ${savePath}`)
      batchExportVisible.value = false
    } catch (err) {
      ElMessage.error('导出失败: ' + err)
    }
  }

  /**
   * 切换批量导出中的应用选中状态
   */
  const toggleBatchExportItem = (appId) => {
    const idx = batchExportSelected.value.indexOf(appId)
    if (idx >= 0) {
      batchExportSelected.value.splice(idx, 1)
    } else {
      batchExportSelected.value.push(appId)
    }
  }

  /**
   * 全选或取消全选批量导出中的应用
   */
  const toggleBatchExportAll = () => {
    if (batchExportSelected.value.length === apps.value.length) {
      batchExportSelected.value = []
    } else {
      batchExportSelected.value = apps.value.map(app => app.id)
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
    appImportVisible,
    appImportTab,
    importZipPath,
    importDirPath,
    importDirName,
    importHtmlPath,
    importHtmlName,
    appEditNameVisible,
    appEditNameValue,
    webAppDialogVisible,
    isEditingWebApp,
    webAppForm,
    batchExportVisible,
    batchExportSelected,
    loadApps,
    refreshApps,
    loadServerStatus,
    getAppIconUrl,
    openApp,
    showAppImport,
    selectZipFile,
    selectImportDir,
    selectHtmlFile,
    doImportZip,
    doImportDir,
    doImportHtml,
    showAddWebAppDialog,
    showEditWebAppDialog,
    saveWebApp,
    updateWebAppForm,
    handleAppCmd,
    saveAppDisplayName,
    doExportApp,
    showBatchExport,
    doBatchExport,
    toggleBatchExportItem,
    toggleBatchExportAll,
    doDeleteApp
  }
}
