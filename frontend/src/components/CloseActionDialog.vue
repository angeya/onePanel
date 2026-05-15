<template>
  <el-dialog
    v-model="visible"
    title="关闭行为"
    width="400px"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    :show-close="false"
    class="close-action-dialog"
    @opened="onOpened"
  >
    <div class="close-action-body">
      <p class="close-action-message">请选择关闭窗口时的行为：</p>
      <el-checkbox v-model="dontAskAgain" class="dont-ask-checkbox">不再提问，记住我的选择</el-checkbox>
    </div>
    <template #footer>
      <div class="close-action-footer">
        <el-button @click="handleClose('close')">直接退出</el-button>
        <el-button type="primary" @click="handleClose('tray')">最小化到托盘</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref } from 'vue'

const visible = ref(false)
const dontAskAgain = ref(false)

let resolvePromise = null

const open = () => {
  dontAskAgain.value = false
  visible.value = true
  return new Promise((resolve) => {
    resolvePromise = resolve
  })
}

const handleClose = (action) => {
  visible.value = false
  if (resolvePromise) {
    resolvePromise({ action, dontAsk: dontAskAgain.value })
    resolvePromise = null
  }
}

const onOpened = () => {
  dontAskAgain.value = false
}

defineExpose({ open })
</script>

<style scoped>
.close-action-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.close-action-message {
  margin: 0;
  font-size: 14px;
  color: var(--text-primary);
}

.dont-ask-checkbox {
  font-size: 12px;
  color: var(--text-muted);
}

.close-action-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
