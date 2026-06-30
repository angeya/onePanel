<template>
  <el-dialog v-model="visible" title="编辑应用名称" width="400px" :close-on-click-modal="false">
    <el-form label-width="80px">
      <el-form-item label="应用名称">
        <el-input v-model="nameValue" placeholder="请输入应用名称" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="handleSave">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, inject } from 'vue'

const appService = inject('appService')

const visible = ref(false)
const nameValue = ref('')
const appId = ref(null)

const open = (app) => {
  appId.value = app.id
  nameValue.value = app.name
  visible.value = true
}

const handleSave = async () => {
  const ok = await appService.saveAppName(appId.value, nameValue.value)
  if (ok) visible.value = false
}

defineExpose({ open })
</script>
