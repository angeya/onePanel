<template>
  <el-dialog
    v-model="visible"
    title="管理分类"
    width="380px"
  >
    <div class="category-manage">
      <div class="category-add-row">
        <el-input v-model="newName" placeholder="分类名称" size="default" />
        <el-button type="primary" @click="addCategory" size="default">添加</el-button>
      </div>
      <div class="category-list">
        <div v-for="cat in categories" :key="cat.id" class="category-manage-item">
          <span>{{ cat.name }}</span>
          <el-icon class="action-icon" @click="deleteCategory(cat.id)"><Delete /></el-icon>
        </div>
        <el-empty v-if="categories.length === 0" description="暂无分类" :image-size="40" />
      </div>
    </div>
  </el-dialog>
</template>

<script setup>
import { ref } from 'vue'
import { Delete } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { CreateSessionCategory, DeleteSessionCategory } from '../../../wailsjs/go/main/ServerListService'

const props = defineProps({
  categories: { type: Array, required: true }
})

const emit = defineEmits(['saved'])

const visible = ref(false)
const newName = ref('')

const show = () => {
  newName.value = ''
  visible.value = true
}

const addCategory = async () => {
  if (!newName.value.trim()) {
    ElMessage.warning('请输入分类名称')
    return
  }
  try {
    await CreateSessionCategory(newName.value.trim(), 0)
    newName.value = ''
    ElMessage.success('分类创建成功')
    emit('saved')
  } catch (err) {
    ElMessage.error('创建分类失败: ' + err)
  }
}

const deleteCategory = async (id) => {
  try {
    await ElMessageBox.confirm(
      '删除分类后，该分类下的会话将变为未分类，确定删除？',
      '提示',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
    await DeleteSessionCategory(id)
    ElMessage.success('分类删除成功')
    emit('saved')
  } catch {
    // 用户取消
  }
}

defineExpose({ show })
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
  padding: 8px;
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
}

.action-icon:hover {
  color: var(--text-primary);
  background-color: var(--bg-active);
}
</style>
