<template>
  <el-dialog :model-value="appImportVisible" @update:model-value="$emit('update:appImportVisible', $event)" title="导入应用" width="500px" :close-on-click-modal="false">
    <el-tabs :model-value="appImportTab" @update:model-value="$emit('update:appImportTab', $event)">
      <el-tab-pane label="导入 ZIP" name="zip">
        <el-form label-width="80px">
          <el-form-item label="ZIP 文件">
            <div style="display: flex; gap: 8px; width: 100%">
              <el-input :model-value="importZipPath" @update:model-value="$emit('update:importZipPath', $event)" placeholder="选择 ZIP 压缩包或手动输入路径" />
              <el-button @click="$emit('selectZipFile')">选择</el-button>
            </div>
          </el-form-item>
        </el-form>
        <div style="text-align: right">
          <el-button type="primary" @click="$emit('doImportZip')" :disabled="!importZipPath">导入</el-button>
        </div>
      </el-tab-pane>
      <el-tab-pane label="导入目录" name="dir">
        <el-form label-width="80px">
          <el-form-item label="应用目录">
            <div style="display: flex; gap: 8px; width: 100%">
              <el-input :model-value="importDirPath" @update:model-value="$emit('update:importDirPath', $event)" placeholder="选择包含 index.html 的目录或手动输入路径" />
              <el-button @click="$emit('selectImportDir')">选择</el-button>
            </div>
          </el-form-item>
          <el-form-item label="应用名称">
            <el-input :model-value="importDirName" @update:model-value="$emit('update:importDirName', $event)" placeholder="留空则使用目录名称" />
          </el-form-item>
        </el-form>
        <div style="text-align: right">
          <el-button type="primary" @click="$emit('doImportDir')" :disabled="!importDirPath">导入</el-button>
        </div>
      </el-tab-pane>
      <el-tab-pane label="导入 HTML" name="html">
        <el-form label-width="80px">
          <el-form-item label="HTML 文件">
            <div style="display: flex; gap: 8px; width: 100%">
              <el-input :model-value="importHtmlPath" @update:model-value="$emit('update:importHtmlPath', $event)" placeholder="选择 HTML 文件" />
              <el-button @click="$emit('selectHtmlFile')">选择</el-button>
            </div>
          </el-form-item>
          <el-form-item label="应用名称" required>
            <el-input :model-value="importHtmlName" @update:model-value="$emit('update:importHtmlName', $event)" placeholder="必填，将作为应用目录名" />
          </el-form-item>
        </el-form>
        <div style="text-align: right">
          <el-button type="primary" @click="$emit('doImportHtml')" :disabled="!importHtmlPath || !importHtmlName">导入</el-button>
        </div>
      </el-tab-pane>
    </el-tabs>
  </el-dialog>
</template>

<script setup>
defineProps({
  appImportVisible: { type: Boolean, default: false },
  appImportTab: { type: String, default: 'zip' },
  importZipPath: { type: String, default: '' },
  importDirPath: { type: String, default: '' },
  importDirName: { type: String, default: '' },
  importHtmlPath: { type: String, default: '' },
  importHtmlName: { type: String, default: '' }
})

defineEmits([
  'update:appImportVisible',
  'update:appImportTab',
  'update:importZipPath',
  'update:importDirPath',
  'update:importDirName',
  'update:importHtmlPath',
  'update:importHtmlName',
  'selectZipFile',
  'selectImportDir',
  'selectHtmlFile',
  'doImportZip',
  'doImportDir',
  'doImportHtml'
])
</script>
