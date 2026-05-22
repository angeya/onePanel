<template>
  <el-dialog
    v-model="visible"
    title="输入登录密码"
    width="380px"
    :close-on-click-modal="false"
  >
    <div class="deploy-info">
      <span class="deploy-server">{{ displayName }}</span>
    </div>
    <el-form :model="form" label-width="80px" size="default" @submit.prevent>
      <el-form-item label="登录密码">
        <el-input
          ref="passwordInputRef"
          v-model="form.password"
          type="password"
          show-password
          placeholder="输入服务器登录密码"
          @keyup.enter="handleSetup"
        />
      </el-form-item>
    </el-form>
    <div v-if="errorMsg" class="setup-error">{{ errorMsg }}</div>
    <template #footer>
      <el-button @click="handleSkip">跳过</el-button>
      <el-button type="primary" @click="handleSetup" :loading="deploying">
        确认
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { DeployKey } from '../../../wailsjs/go/main/ServerListService'

const emit = defineEmits(['deployed', 'skip'])

const visible = ref(false)
const deploying = ref(false)
const errorMsg = ref('')
const server = ref(null)
const passwordInputRef = ref(null)
const form = ref({ password: '' })

const displayName = ref('')

const show = (srv) => {
  server.value = srv
  form.value.password = ''
  errorMsg.value = ''
  const name = srv.sessionName || `${srv.host} (${srv.user})`
  displayName.value = name
  visible.value = true
  nextTick(() => {
    passwordInputRef.value?.focus()
  })
}

const handleSetup = async () => {
  if (!form.value.password) {
    errorMsg.value = '请输入密码'
    return
  }

  deploying.value = true
  errorMsg.value = ''
  try {
    await DeployKey(server.value.id, form.value.password)
    ElMessage.success('设置成功，后续可免密登录')
    visible.value = false
    emit('deployed')
  } catch (err) {
    errorMsg.value = '密码验证失败: ' + err
  } finally {
    deploying.value = false
  }
}

const handleSkip = () => {
  visible.value = false
  emit('skip', form.value.password || '')
}

defineExpose({ show })
</script>

<style scoped>
.deploy-info {
  margin-bottom: 16px;
}

.deploy-server {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
}

.setup-error {
  font-size: 12px;
  color: var(--el-color-danger);
  margin-top: -8px;
  margin-bottom: 4px;
  line-height: 1.4;
}
</style>
