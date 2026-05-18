<template>
  <el-dialog
    v-model="visible"
    :title="isEditing ? '编辑快捷命令' : '新增快捷命令'"
    width="450px"
    :close-on-click-modal="false"
  >
    <el-form :model="form" label-width="80px" size="default">
      <el-form-item label="命令名称">
        <el-input v-model="form.name" placeholder="请输入命令名称" />
      </el-form-item>
      <el-form-item label="所属分类">
        <el-select v-model="form.categoryId" placeholder="请选择分类" clearable style="width: 100%">
          <el-option
            v-for="cat in categories"
            :key="cat.id"
            :label="cat.name"
            :value="cat.id"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="Shell">
        <el-select v-model="form.shell" style="width: 100%">
          <el-option label="cmd.exe" value="cmd.exe" />
          <el-option label="powershell.exe" value="powershell.exe" />
        </el-select>
      </el-form-item>
      <el-form-item label="工作目录">
        <el-input v-model="form.workDir" placeholder="留空则使用默认目录" />
      </el-form-item>
      <el-form-item label="命令内容">
        <el-input
          v-model="form.commands"
          type="textarea"
          :rows="4"
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
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { CreateCommand, UpdateCommand } from '../../../wailsjs/go/main/ShortcutService'

const props = defineProps({
  categories: { type: Array, required: true }
})

const emit = defineEmits(['saved'])

const visible = ref(false)
const isEditing = ref(false)
const editingId = ref(null)
const form = ref(getDefaultForm())

function getDefaultForm() {
  return {
    name: '',
    categoryId: null,
    shell: 'cmd.exe',
    workDir: '',
    commands: ''
  }
}

const show = (cmd) => {
  if (cmd) {
    isEditing.value = true
    editingId.value = cmd.id
    form.value = {
      name: cmd.name,
      categoryId: cmd.categoryId,
      shell: cmd.shell,
      workDir: cmd.workDir,
      commands: cmd.commands
    }
  } else {
    isEditing.value = false
    editingId.value = null
    form.value = getDefaultForm()
  }
  visible.value = true
}

const handleSave = async () => {
  if (!form.value.name) {
    ElMessage.warning('请输入命令名称')
    return
  }
  if (!form.value.commands) {
    ElMessage.warning('请输入命令内容')
    return
  }

  try {
    if (isEditing.value) {
      await UpdateCommand(
        editingId.value,
        form.value.categoryId,
        form.value.name,
        form.value.shell,
        form.value.workDir,
        form.value.commands,
        0
      )
      ElMessage.success('更新成功')
    } else {
      await CreateCommand(
        form.value.categoryId,
        form.value.name,
        form.value.shell,
        form.value.workDir,
        form.value.commands,
        0
      )
      ElMessage.success('创建成功')
    }
    visible.value = false
    emit('saved')
  } catch (err) {
    ElMessage.error('保存失败: ' + err)
  }
}

defineExpose({ show })
</script>
