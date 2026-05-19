<template>
  <el-dialog
    v-model="visible"
    :title="isEditing ? '编辑会话' : '新增会话'"
    width="420px"
    :close-on-click-modal="false"
  >
    <el-form :model="form" label-width="100px" size="default">
      <el-form-item label="会话名称">
        <el-input v-model="form.sessionName" placeholder="可选，默认使用 主机(用户名)" />
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
      <el-form-item label="免密登录">
        <el-checkbox v-model="form.useKeyLogin">
          记住登录方式，以后免输入密码
        </el-checkbox>
      </el-form-item>
      <el-form-item v-if="form.useKeyLogin && !isEditing" label="登录密码">
        <el-input v-model="form.password" type="password" show-password placeholder="输入服务器登录密码" />
        <div class="key-hint">
          首次登录需输入密码，之后可免密登录
        </div>
      </el-form-item>
      <div v-if="setupError" class="setup-error">{{ setupError }}</div>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button
        v-if="setupError"
        @click="handleSkipSetup"
      >
        跳过，直接登录
      </el-button>
      <el-button type="primary" @click="handleSave" :loading="saving">
        {{ isEditing ? '保存' : '确认并登录' }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { AddServer, UpdateServer, DeployKey } from '../../../wailsjs/go/main/ServerListService'

const props = defineProps({
  categories: { type: Array, required: true }
})

const emit = defineEmits(['saved', 'login'])

const visible = ref(false)
const isEditing = ref(false)
const editingId = ref(null)
const saving = ref(false)
const setupError = ref('')
const pendingServer = ref(null)
const form = ref(getDefaultForm())

function getDefaultForm() {
  return {
    sessionName: '',
    categoryId: null,
    host: '',
    port: 22,
    user: '',
    useKeyLogin: true,
    password: ''
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
      useKeyLogin: server.useKeyLogin,
      password: ''
    }
  } else {
    isEditing.value = false
    editingId.value = null
    form.value = getDefaultForm()
  }
  setupError.value = ''
  pendingServer.value = null
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
  setupError.value = ''
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
      emit('saved')
      pendingServer.value = server

      if (form.value.useKeyLogin && form.value.password) {
        try {
          await DeployKey(server.id, form.value.password)
          ElMessage.success('设置成功，后续可免密登录')
          emit('saved')
        } catch (err) {
          setupError.value = '免密登录设置失败: ' + err + '，请检查密码是否正确后重试'
          saving.value = false
          return
        }
      }

      visible.value = false
      emit('login', server)
    }
  } catch (err) {
    ElMessage.error('保存失败: ' + err)
  } finally {
    saving.value = false
  }
}

const handleSkipSetup = () => {
  if (pendingServer.value) {
    visible.value = false
    emit('login', pendingServer.value)
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

.setup-error {
  font-size: 12px;
  color: var(--el-color-danger);
  margin: -4px 0 0 100px;
  line-height: 1.4;
}
</style>
