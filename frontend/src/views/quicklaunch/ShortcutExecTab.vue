<template>
  <div class="shortcut-exec-tab">
    <div class="exec-header">
      <span class="exec-title">{{ commandName }}</span>
      <el-tag size="small" :type="shell === 'powershell.exe' ? 'info' : 'primary'">
        {{ shell === 'powershell.exe' ? 'PowerShell' : 'CMD' }}
      </el-tag>
    </div>
    <div class="exec-info">
      <div v-if="workDir" class="info-row">
        <span class="info-label">工作目录:</span>
        <span class="info-value">{{ workDir }}</span>
      </div>
      <div class="info-row">
        <span class="info-label">命令内容:</span>
        <pre class="command-text">{{ commands }}</pre>
      </div>
    </div>
    <div class="exec-result">
      <div class="result-header">
        <span>执行结果</span>
        <el-tag v-if="execState === 'running'" type="warning" size="small">执行中</el-tag>
        <el-tag v-else-if="execState === 'success'" type="success" size="small">执行成功</el-tag>
        <el-tag v-else-if="execState === 'error'" type="danger" size="small">执行失败</el-tag>
      </div>
      <div class="result-body">
        <div v-if="execState === 'idle'" class="result-idle">
          <el-button type="primary" @click="doExecute">执行命令</el-button>
        </div>
        <div v-else-if="execState === 'running'" class="result-running">
          <el-icon class="is-loading"><Loading /></el-icon>
          <span>正在执行命令...</span>
        </div>
        <div v-else class="result-message">{{ resultMessage }}</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { Loading } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { ExecuteCommand } from '../../../wailsjs/go/main/ShortcutCmdService'

const props = defineProps({
  commandId: { type: Number, required: true },
  commandName: { type: String, required: true },
  shell: { type: String, default: 'cmd.exe' },
  workDir: { type: String, default: '' },
  commands: { type: String, required: true }
})

const execState = ref('idle')
const resultMessage = ref('')

/**
 * 执行快捷命令
 */
const doExecute = async () => {
  execState.value = 'running'
  try {
    await ExecuteCommand(props.commandId)
    execState.value = 'success'
    resultMessage.value = '命令已成功执行'
    ElMessage.success(`已执行: ${props.commandName}`)
  } catch (err) {
    execState.value = 'error'
    resultMessage.value = `执行失败: ${err}`
    ElMessage.error('执行失败: ' + err)
  }
}
</script>

<style scoped>
.shortcut-exec-tab {
  display: flex;
  flex-direction: column;
  height: 100%;
  background-color: #252526;
  padding: 16px 20px;
}

.exec-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.exec-title {
  font-size: 16px;
  font-weight: 600;
  color: #e5e5e5;
}

.exec-info {
  background-color: #1e1e1e;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
}

.info-row {
  margin-bottom: 12px;
}

.info-row:last-child {
  margin-bottom: 0;
}

.info-label {
  font-size: 12px;
  color: #888;
  display: block;
  margin-bottom: 4px;
}

.info-value {
  font-size: 13px;
  color: #c0c0c0;
  font-family: Consolas, 'Courier New', monospace;
}

.command-text {
  font-size: 13px;
  color: #e5e5e5;
  font-family: Consolas, 'Courier New', monospace;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
}

.exec-result {
  flex: 1;
  background-color: #1e1e1e;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.result-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid #2d2d2d;
  font-size: 13px;
  color: #c0c0c0;
}

.result-body {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
}

.result-idle {
  text-align: center;
}

.result-running {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #e6a23c;
  font-size: 14px;
}

.result-message {
  font-size: 13px;
  color: #c0c0c0;
  font-family: Consolas, 'Courier New', monospace;
  text-align: center;
}
</style>
