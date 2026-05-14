import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  GetServerStatus,
  ScanApps, UpdateDisplayName,
  DeleteApp, ExportApp, ImportZip, ImportDir, ImportHtml, OpenApp as OpenAppService,
  BatchExportApps, CreateWebApp, UpdateWebApp
} from '../../wailsjs/go/main/AppService'
import { OpenFileDialog, OpenDirectoryDialog, SaveFileDialog } from '../../wailsjs/go/main/App'

export function useAppService(closeAppTab) {
  const apps = ref([])
  const appsLoading = ref(false)
  const serverStatus = ref({ running: false, port: 0, dir: '' })

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

  const loadServerStatus = async () => {
    try {
      serverStatus.value = await GetServerStatus()
    } catch (err) {
      console.error('获取服务器状态失败:', err)
    }
  }

  const getAppIconUrl = (app) => {
    if (!app.iconPath || !serverStatus.value.running) return ''
    const dir = app.entryUrl.replace('/index.html', '')
    return `http://127.0.0.1:${serverStatus.value.port}${dir}/icon.png`
  }

  const openApp = async (app, addAppTab) => {
    try {
      const result = await OpenAppService(app.id)
      await loadServerStatus()
      addAppTab(app.id, result.name, result.url)
    } catch (err) {
      ElMessage.error('打开应用失败: ' + err)
    }
  }

  const selectZipFile = async () => {
    try {
      const path = await OpenFileDialog('选择 ZIP 文件', 'ZIP 文件 (*.zip)|*.zip')
      if (path) return path
    } catch (err) {
      console.error('选择文件失败:', err)
      ElMessage.warning('文件选择对话框打开失败，请手动输入路径')
    }
    return null
  }

  const selectImportDir = async () => {
    try {
      const dir = await OpenDirectoryDialog('选择应用目录')
      if (dir) return dir
    } catch (err) {
      console.error('选择目录失败:', err)
      ElMessage.warning('目录选择对话框打开失败，请手动输入路径')
    }
    return null
  }

  const selectHtmlFile = async () => {
    try {
      const path = await OpenFileDialog('选择 HTML 文件', 'HTML 文件 (*.html;*.htm)|*.html;*.htm')
      if (path) return path
    } catch (err) {
      console.error('选择文件失败:', err)
      ElMessage.warning('文件选择对话框打开失败，请手动输入路径')
    }
    return null
  }

  const doImportZip = async (zipPath) => {
    if (!zipPath) { ElMessage.warning('请选择 ZIP 文件'); return }
    try {
      const skipped = await ImportZip(zipPath)
      await refreshApps()
      if (skipped) {
        ElMessage.warning(`以下应用已存在，已跳过: ${skipped}`)
      } else {
        ElMessage.success('导入成功')
      }
      return true
    } catch (err) {
      ElMessage.error('导入失败: ' + err)
      return false
    }
  }

  const doImportDir = async (dirPath, dirName) => {
    if (!dirPath) { ElMessage.warning('请选择应用目录'); return }
    try {
      await ImportDir(dirPath, dirName)
      ElMessage.success('导入成功')
      await refreshApps()
      return true
    } catch (err) {
      ElMessage.error('导入失败: ' + err)
      return false
    }
  }

  const doImportHtml = async (htmlPath, htmlName) => {
    if (!htmlPath) { ElMessage.warning('请选择 HTML 文件'); return }
    if (!htmlName) { ElMessage.warning('请输入应用名称'); return }
    try {
      await ImportHtml(htmlPath, htmlName)
      ElMessage.success('导入成功')
      await refreshApps()
      return true
    } catch (err) {
      ElMessage.error('导入失败: ' + err)
      return false
    }
  }

  const saveWebApp = async (isEditing, editingId, form) => {
    if (!form.name) { ElMessage.warning('请输入应用名称'); return }
    if (!form.url) { ElMessage.warning('请输入应用地址'); return }

    try {
      if (isEditing) {
        await UpdateWebApp(editingId, form.name, form.url)
        ElMessage.success('修改成功')
      } else {
        await CreateWebApp(form.name, form.url)
        ElMessage.success('添加成功')
      }
      await loadApps()
      return true
    } catch (err) {
      ElMessage.error('保存失败: ' + err)
      return false
    }
  }

  const showEditWebAppDialog = (app) => {
    return { isEditing: true, editingId: app.id, form: { name: app.displayName, url: app.entryUrl } }
  }

  const saveAppDisplayName = async (appId, name) => {
    try {
      await UpdateDisplayName(appId, name)
      ElMessage.success('修改成功')
      await loadApps()
      return true
    } catch (err) {
      ElMessage.error('修改失败: ' + err)
      return false
    }
  }

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

  const doBatchExport = async (selectedIds) => {
    if (selectedIds.length === 0) { ElMessage.warning('请至少选择一个应用'); return }
    try {
      const defaultName = `apps_export_${new Date().toISOString().replace(/[-:T]/g, '').slice(0, 14)}.zip`
      const savePath = await SaveFileDialog('批量导出应用', defaultName, 'ZIP 文件 (*.zip)|*.zip')
      if (!savePath) return

      await BatchExportApps(selectedIds, savePath)
      ElMessage.success(`已导出 ${selectedIds.length} 个应用到: ${savePath}`)
      return true
    } catch (err) {
      ElMessage.error('导出失败: ' + err)
      return false
    }
  }

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

  const handleAppCmd = (command, app) => {
    switch (command) {
      case 'edit':
        return { cmd: 'edit', app }
      case 'export':
        doExportApp(app)
        return null
      case 'delete':
        doDeleteApp(app)
        return null
    }
  }

  return {
    apps,
    appsLoading,
    serverStatus,
    loadApps,
    refreshApps,
    loadServerStatus,
    getAppIconUrl,
    openApp,
    selectZipFile,
    selectImportDir,
    selectHtmlFile,
    doImportZip,
    doImportDir,
    doImportHtml,
    saveWebApp,
    showEditWebAppDialog,
    saveAppDisplayName,
    doExportApp,
    doBatchExport,
    handleAppCmd,
    doDeleteApp
  }
}
