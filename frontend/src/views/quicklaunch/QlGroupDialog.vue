<template>
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
</template>

<script setup>
import { Delete } from '@element-plus/icons-vue'

defineProps({
  qlGroupDialogVisible: { type: Boolean, default: false },
  qlGroups: { type: Array, default: () => [] },
  newGroupName: { type: String, default: '' }
})

defineEmits([
  'update:qlGroupDialogVisible',
  'update:newGroupName',
  'addQlGroup',
  'deleteQlGroup'
])
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
</style>
