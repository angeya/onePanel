<template>
  <div>
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

    <el-dialog :model-value="appEditNameVisible" @update:model-value="$emit('update:appEditNameVisible', $event)" title="编辑应用名称" width="400px" :close-on-click-modal="false">
      <el-form label-width="80px">
        <el-form-item label="应用名称">
          <el-input :model-value="appEditNameValue" @update:model-value="$emit('update:appEditNameValue', $event)" placeholder="请输入应用名称" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="$emit('update:appEditNameVisible', false)">取消</el-button>
        <el-button type="primary" @click="$emit('saveAppDisplayName')">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog
      :model-value="webAppDialogVisible"
      @update:model-value="$emit('update:webAppDialogVisible', $event)"
      :title="isEditingWebApp ? '编辑网页应用' : '新增网页应用'"
      width="480px"
      :close-on-click-modal="false"
    >
      <el-form label-width="90px" size="default">
        <el-form-item label="应用名称" required>
          <el-input :model-value="webAppForm.name" @update:model-value="$emit('updateWebAppForm', { key: 'name', value: $event })" placeholder="请输入应用名称" />
        </el-form-item>
        <el-form-item label="应用地址" required>
          <el-input :model-value="webAppForm.url" @update:model-value="$emit('updateWebAppForm', { key: 'url', value: $event })" placeholder="请输入网页地址，如 https://example.com" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="$emit('update:webAppDialogVisible', false)">取消</el-button>
        <el-button type="primary" @click="$emit('saveWebApp')">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog
      :model-value="batchExportVisible"
      @update:model-value="$emit('update:batchExportVisible', $event)"
      title="批量导出应用"
      width="460px"
      :close-on-click-modal="false"
    >
      <div class="batch-export-header">
        <el-checkbox
          :model-value="isAllSelected"
          :indeterminate="isPartialSelected"
          @change="$emit('toggleBatchExportAll')"
        >
          全选
        </el-checkbox>
        <span class="batch-export-count">已选 {{ batchExportSelected.length }} 个</span>
      </div>
      <div class="batch-export-list">
        <div
          v-for="app in apps"
          :key="app.id"
          class="batch-export-item"
          :class="{ selected: batchExportSelected.includes(app.id) }"
          @click="$emit('toggleBatchExportItem', app.id)"
        >
          <el-checkbox
            :model-value="batchExportSelected.includes(app.id)"
            @click.stop
            @change="$emit('toggleBatchExportItem', app.id)"
          />
          <el-icon v-if="app.appType === 'web'" :size="16" color="#67c23a"><Link /></el-icon>
          <el-icon v-else :size="16" color="#409eff"><Document /></el-icon>
          <span class="batch-export-name">{{ app.displayName }}</span>
          <span class="batch-export-dir">{{ app.appType === 'web' ? '网页应用' : app.dirName }}</span>
        </div>
        <el-empty v-if="apps.length === 0" description="没有可导出的应用" :image-size="40" />
      </div>
      <template #footer>
        <el-button @click="$emit('update:batchExportVisible', false)">取消</el-button>
        <el-button type="primary" @click="$emit('doBatchExport')" :disabled="batchExportSelected.length === 0">导出</el-button>
      </template>
    </el-dialog>

    <el-dialog
      :model-value="qlCmdDialogVisible"
      @update:model-value="$emit('update:qlCmdDialogVisible', $event)"
      :title="isEditingQlCmd ? '编辑快速启动' : '新增快速启动'"
      width="520px"
      :close-on-click-modal="false"
    >
      <el-form :model="qlCmdForm" label-width="90px" size="default">
        <el-form-item label="命令名称" required>
          <el-input :model-value="qlCmdForm.name" @update:model-value="$emit('updateQlCmdForm', { key: 'name', value: $event })" placeholder="请输入命令名称" />
        </el-form-item>
        <el-form-item label="所属分组">
          <el-select :model-value="qlCmdForm.groupId" @update:model-value="$emit('updateQlCmdForm', { key: 'groupId', value: $event })" placeholder="请选择分组（可选）" clearable style="width: 100%">
            <el-option v-for="group in qlGroups" :key="group.id" :label="group.name" :value="group.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="Shell 类型" required>
          <el-radio-group :model-value="qlCmdForm.shell" @update:model-value="$emit('updateQlCmdForm', { key: 'shell', value: $event })">
            <el-radio value="cmd.exe">CMD</el-radio>
            <el-radio value="powershell.exe">PowerShell</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="工作目录">
          <div style="display: flex; gap: 8px; width: 100%">
            <el-input :model-value="qlCmdForm.workDir" @update:model-value="$emit('updateQlCmdForm', { key: 'workDir', value: $event })" placeholder="留空则使用默认目录" />
            <el-button @click="$emit('selectWorkDir')">选择</el-button>
          </div>
        </el-form-item>
        <el-form-item label="命令内容" required>
          <el-input
            :model-value="qlCmdForm.commands"
            @update:model-value="$emit('updateQlCmdForm', { key: 'commands', value: $event })"
            type="textarea"
            :rows="5"
            placeholder="每行一条命令"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="$emit('update:qlCmdDialogVisible', false)">取消</el-button>
        <el-button type="primary" @click="$emit('saveQlCmd')">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog :model-value="qlGroupDialogVisible" @update:model-value="$emit('update:qlGroupDialogVisible', $event)" title="管理分组" width="420px" :close-on-click-modal="false">
      <div class="group-manage">
        <div class="group-add-row">
          <el-input :model-value="newGroupName" @update:model-value="$emit('update:newGroupName', $event)" placeholder="输入分组名称" @keyup.enter="$emit('addQlGroup')" />
          <el-button type="primary" @click="$emit('addQlGroup')">添加</el-button>
        </div>
        <div class="group-list">
          <div v-for="group in qlGroups" :key="group.id" class="group-manage-item">
            <span>{{ group.name }}</span>
            <el-icon class="action-icon" @click="$emit('deleteQlGroup', group)"><Delete /></el-icon>
          </div>
          <el-empty v-if="qlGroups.length === 0" description="暂无分组" :image-size="40" />
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Delete, Document, Link } from '@element-plus/icons-vue'

const props = defineProps({
  apps: { type: Array, default: () => [] },
  appImportVisible: { type: Boolean, default: false },
  appImportTab: { type: String, default: 'zip' },
  importZipPath: { type: String, default: '' },
  importDirPath: { type: String, default: '' },
  importDirName: { type: String, default: '' },
  importHtmlPath: { type: String, default: '' },
  importHtmlName: { type: String, default: '' },
  appEditNameVisible: { type: Boolean, default: false },
  appEditNameValue: { type: String, default: '' },
  webAppDialogVisible: { type: Boolean, default: false },
  isEditingWebApp: { type: Boolean, default: false },
  webAppForm: { type: Object, default: () => ({ name: '', url: '' }) },
  batchExportVisible: { type: Boolean, default: false },
  batchExportSelected: { type: Array, default: () => [] },
  qlCmdDialogVisible: { type: Boolean, default: false },
  isEditingQlCmd: { type: Boolean, default: false },
  qlCmdForm: { type: Object, default: () => ({}) },
  qlGroups: { type: Array, default: () => [] },
  qlGroupDialogVisible: { type: Boolean, default: false },
  newGroupName: { type: String, default: '' }
})

defineEmits([
  'update:appImportVisible',
  'update:appImportTab',
  'update:importZipPath',
  'update:importDirPath',
  'update:importDirName',
  'update:importHtmlPath',
  'update:importHtmlName',
  'update:appEditNameVisible',
  'update:appEditNameValue',
  'update:webAppDialogVisible',
  'updateWebAppForm',
  'saveWebApp',
  'update:batchExportVisible',
  'doBatchExport',
  'toggleBatchExportItem',
  'toggleBatchExportAll',
  'update:qlCmdDialogVisible',
  'update:qlGroupDialogVisible',
  'update:newGroupName',
  'updateQlCmdForm',
  'selectZipFile',
  'selectImportDir',
  'selectHtmlFile',
  'doImportZip',
  'doImportDir',
  'doImportHtml',
  'saveAppDisplayName',
  'selectWorkDir',
  'saveQlCmd',
  'addQlGroup',
  'deleteQlGroup'
])

const isAllSelected = computed(() => {
  return props.apps.length > 0 && props.batchExportSelected.length === props.apps.length
})

const isPartialSelected = computed(() => {
  return props.batchExportSelected.length > 0 && props.batchExportSelected.length < props.apps.length
})
</script>

<style scoped>
.group-manage {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.group-add-row {
  display: flex;
  gap: 8px;
}

.group-add-row .el-input {
  flex: 1;
}

.group-list {
  max-height: 300px;
  overflow-y: auto;
}

.group-manage-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  border-radius: 4px;
  color: var(--text-primary);
  font-size: 13px;
}

.group-manage-item:hover {
  background-color: var(--bg-hover);
}

.action-icon {
  cursor: pointer;
  color: var(--text-muted);
  padding: 2px;
  border-radius: 4px;
  font-size: 14px;
}

.action-icon:hover {
  color: var(--text-primary);
  background-color: var(--bg-active);
}

.batch-export-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.batch-export-count {
  font-size: 12px;
  color: var(--text-muted);
}

.batch-export-list {
  max-height: 300px;
  overflow-y: auto;
}

.batch-export-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.15s;
}

.batch-export-item:hover {
  background-color: var(--bg-hover);
}

.batch-export-item.selected {
  background-color: var(--bg-active);
}

.batch-export-name {
  flex: 1;
  font-size: 13px;
  color: var(--text-primary);
}

.batch-export-dir {
  font-size: 11px;
  color: var(--text-faint);
}
</style>
