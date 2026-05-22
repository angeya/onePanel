<template>
  <div class="command-item-actions" @click.stop>
    <el-icon class="action-icon" title="执行" @click="$emit('execute')"><VideoPlay /></el-icon>
    <el-icon class="action-icon" title="编辑" @click="$emit('edit')"><Edit /></el-icon>
    <el-icon class="action-icon" title="删除" @click="$emit('delete')"><Delete /></el-icon>
  </div>
</template>

<script setup>
import { Edit, Delete, VideoPlay } from '@element-plus/icons-vue'

defineEmits(['execute', 'edit', 'delete'])
</script>

<style scoped>
.command-item-actions {
  display: flex;
  gap: 2px;
  opacity: 0;
  transition: opacity 0.15s;
  flex-shrink: 0;
}

.command-item:hover .command-item-actions {
  opacity: 1;
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
</style>
