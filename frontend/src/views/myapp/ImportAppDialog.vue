<template>
  <el-dialog v-model="visible" title="导入应用" width="500px" :close-on-click-modal="false">
    <el-tabs v-model="activeTab">
      <el-tab-pane label="导入 ZIP" name="zip">
        <el-form label-width="80px">
          <el-form-item label="ZIP 文件">
            <div style="display: flex; gap: 8px; width: 100%">
              <el-input v-model="zipPath" placeholder="选择 ZIP 压缩包或手动输入路径" />
              <el-button @click="handleSelectZipFile">选择</el-button>
            </div>
          </el-form-item>
        </el-form>
        <div style="text-align: right">
          <el-button type="primary" @click="handleDoImportZip" :disabled="!zipPath">导入</el-button>
        </div>
      </el-tab-pane>
      <el-tab-pane label="导入目录" name="dir">
        <el-form label-width="80px">
          <el-form-item label="应用目录">
            <div style="display: flex; gap: 8px; width: 100%">
              <el-input v-model="dirPath" placeholder="选择包含 index.html 的目录或手动输入路径" />
              <el-button @click="handleSelectImportDir">选择</el-button>
            </div>
          </el-form-item>
          <el-form-item label="应用名称">
            <el-input v-model="dirName" placeholder="留空则使用目录名称" />
          </el-form-item>
        </el-form>
        <div style="text-align: right">
          <el-button type="primary" @click="handleDoImportDir" :disabled="!dirPath">导入</el-button>
        </div>
      </el-tab-pane>
      <el-tab-pane label="导入 HTML" name="html">
        <el-form label-width="80px">
          <el-form-item label="HTML 文件">
            <div style="display: flex; gap: 8px; width: 100%">
              <el-input v-model="htmlPath" placeholder="选择 HTML 文件" />
              <el-button @click="handleSelectHtmlFile">选择</el-button>
            </div>
          </el-form-item>
          <el-form-item label="应用名称" required>
            <el-input v-model="htmlName" placeholder="必填，将作为应用目录名" />
          </el-form-item>
        </el-form>
        <div style="text-align: right">
          <el-button type="primary" @click="handleDoImportHtml" :disabled="!htmlPath || !htmlName">导入</el-button>
        </div>
      </el-tab-pane>
    </el-tabs>
  </el-dialog>
</template>

<script setup>
import { ref, inject } from 'vue'

const appService = inject('appService')

const visible = ref(false)
const activeTab = ref('zip')
const zipPath = ref('')
const dirPath = ref('')
const dirName = ref('')
const htmlPath = ref('')
const htmlName = ref('')

const resetForm = () => {
  zipPath.value = ''
  dirPath.value = ''
  dirName.value = ''
  htmlPath.value = ''
  htmlName.value = ''
}

const open = () => {
  resetForm()
  activeTab.value = 'zip'
  visible.value = true
}

const handleSelectZipFile = async () => {
  const path = await appService.selectZipFile()
  if (path) zipPath.value = path
}

const handleSelectImportDir = async () => {
  const dir = await appService.selectImportDir()
  if (dir) dirPath.value = dir
}

const handleSelectHtmlFile = async () => {
  const path = await appService.selectHtmlFile()
  if (path) htmlPath.value = path
}

const handleDoImportZip = async () => {
  const ok = await appService.doImportZip(zipPath.value)
  if (ok) visible.value = false
}

const handleDoImportDir = async () => {
  const ok = await appService.doImportDir(dirPath.value, dirName.value)
  if (ok) visible.value = false
}

const handleDoImportHtml = async () => {
  const ok = await appService.doImportHtml(htmlPath.value, htmlName.value)
  if (ok) visible.value = false
}

defineExpose({ open })
</script>
