<template>
  <div>
    <ImportAppDialog
      :app-import-visible="importState.visible"
      :app-import-tab="importState.tab"
      :import-zip-path="importState.zipPath"
      :import-dir-path="importState.dirPath"
      :import-dir-name="importState.dirName"
      :import-html-path="importState.htmlPath"
      :import-html-name="importState.htmlName"
      @update:app-import-visible="importState.visible = $event"
      @update:app-import-tab="importState.tab = $event"
      @update:import-zip-path="importState.zipPath = $event"
      @update:import-dir-path="importState.dirPath = $event"
      @update:import-dir-name="importState.dirName = $event"
      @update:import-html-path="importState.htmlPath = $event"
      @update:import-html-name="importState.htmlName = $event"
      @select-zip-file="handleSelectZipFile"
      @select-import-dir="handleSelectImportDir"
      @select-html-file="handleSelectHtmlFile"
      @do-import-zip="handleDoImportZip"
      @do-import-dir="handleDoImportDir"
      @do-import-html="handleDoImportHtml"
    />

    <EditAppNameDialog
      :app-edit-name-visible="editNameState.visible"
      :app-edit-name-value="editNameState.value"
      @update:app-edit-name-visible="editNameState.visible = $event"
      @update:app-edit-name-value="editNameState.value = $event"
      @save-app-display-name="handleSaveAppDisplayName"
    />

    <WebAppDialog
      :web-app-dialog-visible="webAppState.visible"
      :is-editing-web-app="webAppState.isEditing"
      :web-app-form="webAppState.form"
      @update:web-app-dialog-visible="webAppState.visible = $event"
      @update-web-app-form="handleUpdateWebAppForm"
      @save-web-app="handleSaveWebApp"
    />

    <BatchExportDialog
      :apps="appService.apps.value"
      :batch-export-visible="batchExportState.visible"
      :batch-export-selected="batchExportState.selected"
      @update:batch-export-visible="batchExportState.visible = $event"
      @do-batch-export="handleDoBatchExport"
      @toggle-batch-export-item="toggleBatchExportItem"
      @toggle-batch-export-all="toggleBatchExportAll"
    />
  </div>
</template>

<script setup>
import { reactive, inject } from 'vue'
import ImportAppDialog from './ImportAppDialog.vue'
import EditAppNameDialog from './EditAppNameDialog.vue'
import WebAppDialog from './WebAppDialog.vue'
import BatchExportDialog from './BatchExportDialog.vue'

const appService = inject('appService')

const importState = reactive({
  visible: false,
  tab: 'zip',
  zipPath: '',
  dirPath: '',
  dirName: '',
  htmlPath: '',
  htmlName: ''
})

const editNameState = reactive({
  visible: false,
  value: '',
  appId: null
})

const webAppState = reactive({
  visible: false,
  isEditing: false,
  editingId: null,
  form: { name: '', url: '' }
})

const batchExportState = reactive({
  visible: false,
  selected: []
})

const showAppImport = () => {
  importState.zipPath = ''
  importState.dirPath = ''
  importState.dirName = ''
  importState.htmlPath = ''
  importState.htmlName = ''
  importState.visible = true
}

const showAddWebAppDialog = () => {
  webAppState.isEditing = false
  webAppState.editingId = null
  webAppState.form = { name: '', url: '' }
  webAppState.visible = true
}

const showEditWebAppDialog = (app) => {
  const result = appService.showEditWebAppDialog(app)
  webAppState.isEditing = result.isEditing
  webAppState.editingId = result.editingId
  webAppState.form = result.form
  webAppState.visible = true
}

const showEditAppNameDialog = (app) => {
  editNameState.appId = app.id
  editNameState.value = app.displayName
  editNameState.visible = true
}

const showBatchExport = () => {
  if (appService.apps.value.length === 0) return
  batchExportState.selected = []
  batchExportState.visible = true
}

const handleSelectZipFile = async () => {
  const path = await appService.selectZipFile()
  if (path) importState.zipPath = path
}

const handleSelectImportDir = async () => {
  const dir = await appService.selectImportDir()
  if (dir) importState.dirPath = dir
}

const handleSelectHtmlFile = async () => {
  const path = await appService.selectHtmlFile()
  if (path) importState.htmlPath = path
}

const handleDoImportZip = async () => {
  const ok = await appService.doImportZip(importState.zipPath)
  if (ok) importState.visible = false
}

const handleDoImportDir = async () => {
  const ok = await appService.doImportDir(importState.dirPath, importState.dirName)
  if (ok) importState.visible = false
}

const handleDoImportHtml = async () => {
  const ok = await appService.doImportHtml(importState.htmlPath, importState.htmlName)
  if (ok) importState.visible = false
}

const handleUpdateWebAppForm = ({ key, value }) => {
  webAppState.form[key] = value
}

const handleSaveWebApp = async () => {
  const ok = await appService.saveWebApp(webAppState.isEditing, webAppState.editingId, webAppState.form)
  if (ok) webAppState.visible = false
}

const handleSaveAppDisplayName = async () => {
  const ok = await appService.saveAppDisplayName(editNameState.appId, editNameState.value)
  if (ok) editNameState.visible = false
}

const toggleBatchExportItem = (appId) => {
  const idx = batchExportState.selected.indexOf(appId)
  if (idx >= 0) {
    batchExportState.selected.splice(idx, 1)
  } else {
    batchExportState.selected.push(appId)
  }
}

const toggleBatchExportAll = () => {
  if (batchExportState.selected.length === appService.apps.value.length) {
    batchExportState.selected = []
  } else {
    batchExportState.selected = appService.apps.value.map(app => app.id)
  }
}

const handleDoBatchExport = async () => {
  const ok = await appService.doBatchExport(batchExportState.selected)
  if (ok) batchExportState.visible = false
}

const handleAppCmd = (command, app) => {
  if (command === 'edit') {
    if (app.appType === 'web') {
      showEditWebAppDialog(app)
    } else {
      showEditAppNameDialog(app)
    }
  } else {
    appService.handleAppCmd(command, app)
  }
}

defineExpose({
  showAppImport,
  showAddWebAppDialog,
  showBatchExport,
  handleAppCmd
})
</script>
