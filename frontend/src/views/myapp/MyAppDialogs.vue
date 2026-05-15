<template>
  <div>
    <ImportAppDialog ref="importAppRef" />
    <EditAppNameDialog ref="editNameRef" />
    <WebAppDialog ref="webAppRef" />
    <BatchExportDialog ref="batchExportRef" />
  </div>
</template>

<script setup>
import { ref, inject } from 'vue'
import ImportAppDialog from './ImportAppDialog.vue'
import EditAppNameDialog from './EditAppNameDialog.vue'
import WebAppDialog from './WebAppDialog.vue'
import BatchExportDialog from './BatchExportDialog.vue'

const appService = inject('appService')

const importAppRef = ref(null)
const editNameRef = ref(null)
const webAppRef = ref(null)
const batchExportRef = ref(null)

const showAppImport = () => {
  importAppRef.value.open()
}

const showAddWebAppDialog = () => {
  webAppRef.value.openAdd()
}

const showBatchExport = () => {
  batchExportRef.value.open()
}

const handleAppCmd = (command, app) => {
  if (command === 'edit') {
    if (app.appType === 'web') {
      webAppRef.value.openEdit(app)
    } else {
      editNameRef.value.open(app)
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
