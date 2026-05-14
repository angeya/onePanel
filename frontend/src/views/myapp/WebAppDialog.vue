<template>
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
</template>

<script setup>
defineProps({
  webAppDialogVisible: { type: Boolean, default: false },
  isEditingWebApp: { type: Boolean, default: false },
  webAppForm: { type: Object, default: () => ({ name: '', url: '' }) }
})

defineEmits([
  'update:webAppDialogVisible',
  'updateWebAppForm',
  'saveWebApp'
])
</script>
