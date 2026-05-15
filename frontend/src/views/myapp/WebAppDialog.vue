<template>
  <el-dialog
    v-model="visible"
    :title="isEditing ? '编辑网页应用' : '新增网页应用'"
    width="480px"
    :close-on-click-modal="false"
  >
    <el-form label-width="90px" size="default">
      <el-form-item label="应用名称" required>
        <el-input v-model="form.name" placeholder="请输入应用名称" />
      </el-form-item>
      <el-form-item label="应用地址" required>
        <el-input v-model="form.url" placeholder="请输入网页地址，如 https://example.com" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="handleSave">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, inject } from 'vue'

const appService = inject('appService')

const visible = ref(false)
const isEditing = ref(false)
const editingId = ref(null)
const form = reactive({ name: '', url: '' })

const openAdd = () => {
  isEditing.value = false
  editingId.value = null
  form.name = ''
  form.url = ''
  visible.value = true
}

const openEdit = (app) => {
  isEditing.value = true
  editingId.value = app.id
  form.name = app.displayName
  form.url = app.entryUrl
  visible.value = true
}

const handleSave = async () => {
  const ok = await appService.saveWebApp(isEditing.value, editingId.value, { name: form.name, url: form.url })
  if (ok) visible.value = false
}

defineExpose({ openAdd, openEdit })
</script>
