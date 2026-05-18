<template>
  <el-dialog
    v-model="visible"
    :title="isEditing ? '编辑会话' : '新增会话'"
    width="420px"
    :close-on-click-modal="false"
  >
    <el-form :model="form" label-width="100px" size="default">
      <el-form-item label="会话名称">
        <el-input v-model="form.sessionName" placeholder="可选，默认使用 用户名@主机名" />
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
      <el-form-item label="主机地址">
        <el-input v-model="form.host" placeholder="如 192.168.1.100" />
      </el-form-item>
      <el-form-item label="端口">
        <el-input-number v-model="form.port" :min="1" :max="65535" style="width: 100%" />
      </el-form-item>
      <el-form-item label="用户名">
        <el-input v-model="form.user" placeholder="如 root" />
      </el-form-item>
      <el-form-item label="密钥登录">
        <el-checkbox v-model="form.useKeyLogin">
          以后使用密钥免密登录
        </el-checkbox>
        <div v-if="form.useKeyLogin" class="key-hint">
          首次登录需输入密码，登录成功后自动部署公钥
        </div>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="handleSave" :loading="saving">
        {{ isEditing ? '保存' : '确认并登录' }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { AddServer, UpdateServer } from '../../../wailsjs/go/main/ServerListService'

const props = defineProps({
  categories: { type: Array, required: true }
})

const emit = defineEmits(['saved', 'login'])

const visible = ref(false)
const isEditing = ref(false)
const editingId = ref(null)
const saving = ref(false)
const form = ref(getDefaultForm())

function getDefaultForm() {
  return {
    sessionName: '',
    categoryId: null,
    host: '',
    port: 22,
    user: '',
    useKeyLogin: true
  }
}

const show = (server) => {
  if (server) {
    isEditing.value = true
    editingId.value = server.id
    form.value = {
      sessionName: server.sessionName,
      categoryId: server.categoryId,
      host: server.host,
      port: server.port,
      user: server.user,
      useKeyLogin: server.useKeyLogin
    }
  } else {
    isEditing.value = false
    editingId.value = null
    form.value = getDefaultForm()
  }
  visible.value = true
}

const handleSave = async () => {
  if (!form.value.host) {
    ElMessage.warning('请输入主机地址')
    return
  }
  if (!form.value.user) {
    ElMessage.warning('请输入用户名')
    return
  }

  saving.value = true
  try {
    if (isEditing.value) {
      await UpdateServer(
        editingId.value,
        form.value.categoryId,
        form.value.sessionName,
        form.value.host,
        form.value.port,
        form.value.user,
        form.value.useKeyLogin
      )
      ElMessage.success('更新成功')
      visible.value = false
      emit('saved')
    } else {
      const server = await AddServer(
        form.value.categoryId,
        form.value.sessionName,
        form.value.host,
        form.value.port,
        form.value.user,
        form.value.useKeyLogin
      )
      visible.value = false
      emit('saved')
      emit('login', server)
    }
  } catch (err) {
    ElMessage.error('保存失败: ' + err)
  } finally {
    saving.value = false
  }
}

defineExpose({ show })
</script>

<style scoped>
.key-hint {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 4px;
  line-height: 1.4;
}
</style>
