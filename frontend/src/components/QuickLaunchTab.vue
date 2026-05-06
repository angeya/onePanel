<template>
  <div class="quick-launch-tab">
    <div v-if="execLogs.length === 0" class="empty-state">
      <el-empty description="暂无执行记录，双击左侧快速启动命令即可执行" :image-size="60" />
    </div>
    <div v-else class="exec-logs">
      <div
        v-for="(log, index) in execLogs"
        :key="index"
        class="log-item"
        :class="{ 'log-success': log.state === 'success', 'log-error': log.state === 'error', 'log-running': log.state === 'running' }"
      >
        <div class="log-header">
          <span class="log-name">{{ log.name }}</span>
          <el-tag v-if="log.state === 'running'" type="warning" size="small">执行中</el-tag>
          <el-tag v-else-if="log.state === 'success'" type="success" size="small">成功</el-tag>
          <el-tag v-else-if="log.state === 'error'" type="danger" size="small">失败</el-tag>
          <span class="log-time">{{ log.time }}</span>
        </div>
        <div class="log-detail">
          <div class="log-meta">
            <el-tag size="small" :type="log.shell === 'powershell' ? 'info' : 'primary'" class="shell-tag">
              {{ log.shell === 'powershell' ? 'PS' : 'CMD' }}
            </el-tag>
            <span v-if="log.workDir" class="log-workdir" :title="log.workDir">
              <el-icon size="12"><Folder /></el-icon>
              {{ log.workDir }}
            </span>
          </div>
          <pre class="log-commands">{{ log.commands }}</pre>
          <div v-if="log.state === 'error' && log.errorMsg" class="log-error-msg">
            <el-icon size="14" color="#f56c6c"><CircleCloseFilled /></el-icon>
            <span>{{ log.errorMsg }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { Folder, CircleCloseFilled } from '@element-plus/icons-vue'
import { ExecuteCommand } from '../../wailsjs/go/main/ShortcutCmdService'

const execLogs = ref([])

/**
 * 执行快速启动命令并记录日志
 */
const execute = async (cmd) => {
  const now = new Date()
  const timeStr = `${now.getHours().toString().padStart(2, '0')}:${now.getMinutes().toString().padStart(2, '0')}:${now.getSeconds().toString().padStart(2, '0')}`

  const log = {
    name: cmd.name,
    shell: cmd.shell || 'cmd.exe',
    workDir: cmd.workDir || '',
    commands: cmd.commands,
    state: 'running',
    time: timeStr,
    errorMsg: ''
  }

  execLogs.value.unshift(log)

  try {
    await ExecuteCommand(cmd.id)
    log.state = 'success'
  } catch (err) {
    log.state = 'error'
    log.errorMsg = String(err)
  }
}

defineExpose({ execute })
</script>

<style scoped>
.quick-launch-tab {
  display: flex;
  flex-direction: column;
  height: 100%;
  background-color: #252526;
}

.empty-state {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.exec-logs {
  flex: 1;
  overflow-y: auto;
  padding: 12px 16px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.exec-logs::-webkit-scrollbar {
  width: 4px;
}

.exec-logs::-webkit-scrollbar-thumb {
  background-color: #555;
  border-radius: 2px;
}

.log-item {
  background-color: #1e1e1e;
  border-radius: 8px;
  border-left: 3px solid #555;
  overflow: hidden;
}

.log-item.log-success {
  border-left-color: #67c23a;
}

.log-item.log-error {
  border-left-color: #f56c6c;
}

.log-item.log-running {
  border-left-color: #e6a23c;
}

.log-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  border-bottom: 1px solid #2d2d2d;
}

.log-name {
  font-size: 14px;
  font-weight: 500;
  color: #e5e5e5;
  flex: 1;
}

.log-time {
  font-size: 11px;
  color: #666;
  flex-shrink: 0;
}

.log-detail {
  padding: 10px 14px;
}

.log-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.shell-tag {
  flex-shrink: 0;
}

.log-workdir {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: #888;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}

.log-commands {
  font-size: 12px;
  color: #c0c0c0;
  font-family: Consolas, 'Courier New', monospace;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
  background-color: #252526;
  border-radius: 4px;
  padding: 8px 10px;
}

.log-error-msg {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  margin-top: 8px;
  padding: 8px 10px;
  background-color: rgba(245, 108, 108, 0.1);
  border-radius: 4px;
  font-size: 12px;
  color: #f56c6c;
  font-family: Consolas, 'Courier New', monospace;
  word-break: break-all;
}
</style>
