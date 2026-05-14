<template>
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
</template>

<script setup>
defineProps({
  qlCmdDialogVisible: { type: Boolean, default: false },
  isEditingQlCmd: { type: Boolean, default: false },
  qlCmdForm: { type: Object, default: () => ({}) },
  qlGroups: { type: Array, default: () => [] }
})

defineEmits([
  'update:qlCmdDialogVisible',
  'updateQlCmdForm',
  'selectWorkDir',
  'saveQlCmd'
])
</script>
