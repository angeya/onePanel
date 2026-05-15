<template>
  <el-dialog v-model="visible" title="管理分类" width="420px" :close-on-click-modal="false">
    <div class="category-manage">
      <div class="category-add-row">
        <el-input v-model="newCategoryName" placeholder="输入分类名称" @keyup.enter="handleAddCategory" />
        <el-button type="primary" @click="handleAddCategory">添加</el-button>
      </div>
      <div class="category-list">
        <div v-for="category in qlCategories" :key="category.id" class="category-manage-item">
          <span>{{ category.name }}</span>
          <el-icon class="action-icon" @click="handleDeleteCategory(category)"><Delete /></el-icon>
        </div>
        <el-empty v-if="qlCategories.length === 0" description="暂无分类" :image-size="40" />
      </div>
    </div>
  </el-dialog>
</template>

<script setup>
import { ref, computed, inject } from 'vue'
import { Delete } from '@element-plus/icons-vue'

const qlService = inject('qlService')

const visible = ref(false)
const newCategoryName = ref('')

const qlCategories = computed(() => qlService.qlCategories.value)

const open = () => {
  newCategoryName.value = ''
  visible.value = true
}

const handleAddCategory = async () => {
  const ok = await qlService.addQlCategory(newCategoryName.value)
  if (ok) newCategoryName.value = ''
}

const handleDeleteCategory = (category) => {
  qlService.deleteQlCategory(category)
}

defineExpose({ open })
</script>

<style scoped>
.category-manage {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.category-add-row {
  display: flex;
  gap: 8px;
}

.category-add-row .el-input {
  flex: 1;
}

.category-list {
  max-height: 300px;
  overflow-y: auto;
}

.category-manage-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  border-radius: 4px;
  color: var(--text-primary);
  font-size: 13px;
}

.category-manage-item:hover {
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
