<template>
  <el-dialog
    v-model="visible"
    title="重命名会话"
    width="360px"
    :close-on-click-modal="false"
  >
    <el-input v-model="name" placeholder="请输入新的会话名称" @keyup.enter="handleRename" />
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="handleRename">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { RenameServer } from '../../../wailsjs/go/main/ServerListService'

const emit = defineEmits(['saved'])

const visible = ref(false)
const name = ref('')
const targetId = ref(null)

const show = (server) => {
  name.value = server.sessionName
  targetId.value = server.id
  visible.value = true
}

const handleRename = async () => {
  if (!name.value.trim()) {
    ElMessage.warning('会话名称不能为空')
    return
  }

  try {
    await RenameServer(targetId.value, name.value.trim())
    ElMessage.success('重命名成功')
    visible.value = false
    emit('saved')
  } catch (err) {
    ElMessage.error('重命名失败: ' + err)
  }
}

defineExpose({ show })
</script>
