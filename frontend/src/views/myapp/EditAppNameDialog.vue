<template>
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
</template>

<script setup>
defineProps({
  appEditNameVisible: { type: Boolean, default: false },
  appEditNameValue: { type: String, default: '' }
})

defineEmits([
  'update:appEditNameVisible',
  'update:appEditNameValue',
  'saveAppDisplayName'
])
</script>
