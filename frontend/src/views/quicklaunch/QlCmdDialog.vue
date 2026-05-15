<template>
  <el-dialog
    v-model="visible"
    :title="isEditing ? '编辑快速启动' : '新增快速启动'"
    width="520px"
    :close-on-click-modal="false"
  >
    <el-form :model="form" label-width="90px" size="default">
      <el-form-item label="命令名称" required>
        <el-input v-model="form.name" placeholder="请输入命令名称" />
      </el-form-item>
      <el-form-item label="所属分类">
        <el-select v-model="form.categoryId" placeholder="请选择分类（可选）" clearable style="width: 100%">
          <el-option v-for="category in qlCategories" :key="category.id" :label="category.name" :value="category.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="Shell 类型" required>
        <el-radio-group v-model="form.shell">
          <el-radio value="cmd.exe">CMD</el-radio>
          <el-radio value="powershell.exe">PowerShell</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="工作目录">
        <div style="display: flex; gap: 8px; width: 100%">
          <el-input v-model="form.workDir" placeholder="留空则使用默认目录" />
          <el-button @click="handleSelectWorkDir">选择</el-button>
        </div>
      </el-form-item>
      <el-form-item label="命令内容" required>
        <el-input
          v-model="form.commands"
          type="textarea"
          :rows="5"
          placeholder="每行一条命令"
        />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="handleSave">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, inject } from 'vue'

const qlService = inject('qlService')

const visible = ref(false)
const isEditing = ref(false)
const editingId = ref(null)
const form = reactive({ name: '', categoryId: null, shell: 'cmd.exe', workDir: '', commands: '' })

const qlCategories = computed(() => qlService.qlCategories.value)

const openAdd = () => {
  isEditing.value = false
  editingId.value = null
  form.name = ''
  form.categoryId = null
  form.shell = 'cmd.exe'
  form.workDir = ''
  form.commands = ''
  visible.value = true
}

const openEdit = (cmd) => {
  isEditing.value = true
  editingId.value = cmd.id
  form.name = cmd.name
  form.categoryId = cmd.categoryId || null
  form.shell = cmd.shell || 'cmd.exe'
  form.workDir = cmd.workDir || ''
  form.commands = cmd.commands
  visible.value = true
}

const handleSelectWorkDir = async () => {
  const dir = await qlService.selectWorkDir()
  if (dir) form.workDir = dir
}

const handleSave = async () => {
  const ok = await qlService.saveQlCmd(isEditing.value, editingId.value, { name: form.name, categoryId: form.categoryId, shell: form.shell, workDir: form.workDir, commands: form.commands })
  if (ok) visible.value = false
}

defineExpose({ openAdd, openEdit })
</script>
