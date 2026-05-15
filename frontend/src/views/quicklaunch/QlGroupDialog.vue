<template>
  <el-dialog v-model="visible" title="管理分组" width="420px" :close-on-click-modal="false">
    <div class="group-manage">
      <div class="group-add-row">
        <el-input v-model="newGroupName" placeholder="输入分组名称" @keyup.enter="handleAddGroup" />
        <el-button type="primary" @click="handleAddGroup">添加</el-button>
      </div>
      <div class="group-list">
        <div v-for="group in qlGroups" :key="group.id" class="group-manage-item">
          <span>{{ group.name }}</span>
          <el-icon class="action-icon" @click="handleDeleteGroup(group)"><Delete /></el-icon>
        </div>
        <el-empty v-if="qlGroups.length === 0" description="暂无分组" :image-size="40" />
      </div>
    </div>
  </el-dialog>
</template>

<script setup>
import { ref, computed, inject } from 'vue'
import { Delete } from '@element-plus/icons-vue'

const qlService = inject('qlService')

const visible = ref(false)
const newGroupName = ref('')

const qlGroups = computed(() => qlService.qlGroups.value)

const open = () => {
  newGroupName.value = ''
  visible.value = true
}

const handleAddGroup = async () => {
  const ok = await qlService.addQlGroup(newGroupName.value)
  if (ok) newGroupName.value = ''
}

const handleDeleteGroup = (group) => {
  qlService.deleteQlGroup(group)
}

defineExpose({ open })
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
