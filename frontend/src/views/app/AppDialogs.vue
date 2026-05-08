<template>
  <div>
    <el-dialog :model-value="appSettingsVisible" @update:model-value="$emit('update:appSettingsVisible', $event)" title="应用设置" width="500px" :close-on-click-modal="false">
      <el-form label-width="100px">
        <el-form-item label="静态目录">
          <div style="display: flex; gap: 8px; width: 100%">
            <el-input :model-value="staticDir" @update:model-value="$emit('update:staticDir', $event)" placeholder="选择静态文件目录或手动输入路径" />
            <el-button @click="$emit('selectDirectory')">选择</el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="$emit('update:appSettingsVisible', false)">关闭</el-button>
        <el-button type="primary" @click="$emit('saveStaticDir')">保存</el-button>
      </template>
    </el-dialog>

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
              <el-input :model-value="importAppName" @update:model-value="$emit('update:importAppName', $event)" placeholder="留空则使用目录名称" />
            </el-form-item>
          </el-form>
          <div style="text-align: right">
            <el-button type="primary" @click="$emit('doImportDir')" :disabled="!importDirPath">导入</el-button>
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

    <el-dialog :model-value="appRenameDirVisible" @update:model-value="$emit('update:appRenameDirVisible', $event)" title="修改目录名称" width="400px" :close-on-click-modal="false">
      <el-form label-width="80px">
        <el-form-item label="目录名称">
          <el-input :model-value="appRenameDirValue" @update:model-value="$emit('update:appRenameDirValue', $event)" placeholder="请输入新的目录名称" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="$emit('update:appRenameDirVisible', false)">取消</el-button>
        <el-button type="primary" @click="$emit('saveAppDirName')">保存</el-button>
      </template>
    </el-dialog>

    <input ref="iconInputRef" type="file" accept="image/png" style="display: none" @change="$emit('handleIconUpload', $event)" />

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
import { ref } from 'vue'
import { Delete } from '@element-plus/icons-vue'

defineProps({
  appSettingsVisible: { type: Boolean, default: false },
  staticDir: { type: String, default: '' },
  appImportVisible: { type: Boolean, default: false },
  appImportTab: { type: String, default: 'zip' },
  importZipPath: { type: String, default: '' },
  importDirPath: { type: String, default: '' },
  importAppName: { type: String, default: '' },
  appEditNameVisible: { type: Boolean, default: false },
  appEditNameValue: { type: String, default: '' },
  appRenameDirVisible: { type: Boolean, default: false },
  appRenameDirValue: { type: String, default: '' },
  qlCmdDialogVisible: { type: Boolean, default: false },
  isEditingQlCmd: { type: Boolean, default: false },
  qlCmdForm: { type: Object, default: () => ({}) },
  qlGroups: { type: Array, default: () => [] },
  qlGroupDialogVisible: { type: Boolean, default: false },
  newGroupName: { type: String, default: '' }
})

defineEmits([
  'update:appSettingsVisible',
  'update:staticDir',
  'update:appImportVisible',
  'update:appImportTab',
  'update:importZipPath',
  'update:importDirPath',
  'update:importAppName',
  'update:appEditNameVisible',
  'update:appEditNameValue',
  'update:appRenameDirVisible',
  'update:appRenameDirValue',
  'update:qlCmdDialogVisible',
  'update:qlGroupDialogVisible',
  'update:newGroupName',
  'updateQlCmdForm',
  'selectDirectory',
  'saveStaticDir',
  'selectZipFile',
  'selectImportDir',
  'doImportZip',
  'doImportDir',
  'saveAppDisplayName',
  'saveAppDirName',
  'handleIconUpload',
  'selectWorkDir',
  'saveQlCmd',
  'addQlGroup',
  'deleteQlGroup'
])

const iconInputRef = ref(null)
defineExpose({ iconInputRef })
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
  color: #e5e5e5;
  font-size: 13px;
}

.group-manage-item:hover {
  background-color: #2d2d2d;
}

.action-icon {
  cursor: pointer;
  color: #a0a0a0;
  padding: 2px;
  border-radius: 4px;
  font-size: 14px;
}

.action-icon:hover {
  color: #e5e5e5;
  background-color: #3d3d3d;
}
</style>
