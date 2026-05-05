<template>
  <div class="tools-page">
    <div class="tools-header">
      <span class="page-title">实用工具</span>
    </div>

    <div class="tools-body">
      <div class="tool-section">
        <div class="section-header">
          <div class="section-title-area">
            <el-icon :size="18" color="#409eff"><Connection /></el-icon>
            <span class="section-title">网络端口</span>
            <el-tag size="small" type="info">共 {{ filteredPorts.length }} 条</el-tag>
          </div>
          <div class="section-actions">
            <el-input
              v-model="searchKeyword"
              placeholder="搜索端口 / PID / 进程名"
              size="small"
              clearable
              style="width: 240px"
              :prefix-icon="Search"
            />
            <el-select v-model="protocolFilter" size="small" style="width: 110px" clearable placeholder="协议">
              <el-option label="TCP" value="TCP" />
              <el-option label="UDP" value="UDP" />
            </el-select>
            <el-select v-model="stateFilter" size="small" style="width: 140px" clearable placeholder="状态">
              <el-option label="LISTENING" value="LISTENING" />
              <el-option label="ESTABLISHED" value="ESTABLISHED" />
              <el-option label="TIME_WAIT" value="TIME_WAIT" />
              <el-option label="CLOSE_WAIT" value="CLOSE_WAIT" />
              <el-option label="SYN_SENT" value="SYN_SENT" />
              <el-option label="SYN_RECEIVED" value="SYN_RECEIVED" />
            </el-select>
            <el-button size="small" type="primary" @click="loadPortList" :loading="loading" plain>
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
          </div>
        </div>

        <div class="port-table-wrapper">
          <el-table
            :data="paginatedPorts"
            stripe
            size="small"
            height="100%"
            :default-sort="{ prop: 'localPort', order: 'ascending' }"
            @sort-change="handleSortChange"
            style="width: 100%"
          >
            <el-table-column prop="protocol" label="协议" width="70" />
            <el-table-column prop="localAddr" label="本地地址" width="130" />
            <el-table-column prop="localPort" label="本地端口" width="100" sortable="custom" />
            <el-table-column prop="foreignAddr" label="外部地址" width="130" />
            <el-table-column prop="foreignPort" label="外部端口" width="100" sortable="custom" />
            <el-table-column prop="state" label="状态" width="130">
              <template #default="{ row }">
                <el-tag
                  v-if="row.state"
                  :type="getStateTagType(row.state)"
                  size="small"
                >
                  {{ row.state }}
                </el-tag>
                <span v-else style="color: #666">-</span>
              </template>
            </el-table-column>
            <el-table-column prop="pid" label="PID" width="80" sortable="custom" />
            <el-table-column prop="processName" label="进程名称" min-width="160">
              <template #default="{ row }">
                <span class="process-name">{{ row.processName || '-' }}</span>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="80" fixed="right">
              <template #default="{ row }">
                <el-button
                  size="small"
                  type="danger"
                  text
                  @click="killProcess(row)"
                >
                  终止
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <div class="pagination-area">
          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :page-sizes="[50, 100, 200, 500]"
            :total="filteredPorts.length"
            layout="total, sizes, prev, pager, next, jumper"
            size="small"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Connection, Search, Refresh } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { GetPortList, KillProcess } from '../../wailsjs/go/main/ToolService'

const ports = ref([])
const loading = ref(false)
const searchKeyword = ref('')
const protocolFilter = ref('')
const stateFilter = ref('')
const currentPage = ref(1)
const pageSize = ref(100)
const sortProp = ref('localPort')
const sortOrder = ref('ascending')

/**
 * 过滤后的端口列表
 */
const filteredPorts = computed(() => {
  let list = ports.value

  if (protocolFilter.value) {
    list = list.filter(p => p.protocol === protocolFilter.value)
  }

  if (stateFilter.value) {
    list = list.filter(p => p.state === stateFilter.value)
  }

  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    list = list.filter(p =>
      String(p.localPort).includes(keyword) ||
      String(p.pid).includes(keyword) ||
      (p.processName && p.processName.toLowerCase().includes(keyword)) ||
      (p.localAddr && p.localAddr.includes(keyword)) ||
      (p.state && p.state.toLowerCase().includes(keyword))
    )
  }

  if (sortProp.value) {
    const prop = sortProp.value
    const order = sortOrder.value === 'ascending' ? 1 : -1
    list = [...list].sort((a, b) => {
      const va = a[prop]
      const vb = b[prop]
      if (typeof va === 'number' && typeof vb === 'number') {
        return (va - vb) * order
      }
      return String(va).localeCompare(String(vb)) * order
    })
  }

  return list
})

/**
 * 分页后的端口列表
 */
const paginatedPorts = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return filteredPorts.value.slice(start, start + pageSize.value)
})

/**
 * 获取状态标签类型
 */
const getStateTagType = (state) => {
  switch (state) {
    case 'LISTENING':
      return 'success'
    case 'ESTABLISHED':
      return 'primary'
    case 'TIME_WAIT':
      return 'warning'
    case 'CLOSE_WAIT':
      return 'danger'
    case 'SYN_SENT':
    case 'SYN_RECEIVED':
      return 'info'
    default:
      return 'info'
  }
}

/**
 * 处理排序变化
 */
const handleSortChange = ({ prop, order }) => {
  sortProp.value = prop
  sortOrder.value = order
}

/**
 * 加载端口列表
 */
const loadPortList = async () => {
  loading.value = true
  try {
    ports.value = await GetPortList()
  } catch (err) {
    ElMessage.error('获取端口列表失败: ' + err)
  } finally {
    loading.value = false
  }
}

/**
 * 终止进程
 */
const killProcess = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定要终止进程 "${row.processName || '未知'}" (PID: ${row.pid}) 吗？这可能导致相关程序异常关闭。`,
      '确认终止进程',
      { confirmButtonText: '终止', cancelButtonText: '取消', type: 'warning' }
    )
    await KillProcess(row.pid)
    ElMessage.success(`进程 ${row.pid} 已终止`)
    await loadPortList()
  } catch (err) {
    if (err !== 'cancel') {
      ElMessage.error('终止进程失败: ' + err)
    }
  }
}

onMounted(() => {
  loadPortList()
})
</script>

<style scoped>
.tools-page {
  display: flex;
  flex-direction: column;
  height: 100%;
  background-color: #252526;
}

.tools-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  background-color: #2d2d2d;
  border-bottom: 1px solid #3d3d3d;
}

.page-title {
  font-size: 16px;
  font-weight: 600;
  color: #e5e5e5;
}

.tools-body {
  flex: 1;
  overflow: hidden;
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
}

.tool-section {
  display: flex;
  flex-direction: column;
  flex: 1;
  overflow: hidden;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  flex-shrink: 0;
}

.section-title-area {
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-title {
  font-size: 15px;
  font-weight: 500;
  color: #e5e5e5;
}

.section-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.port-table-wrapper {
  flex: 1;
  overflow: hidden;
}

.pagination-area {
  display: flex;
  justify-content: flex-end;
  padding-top: 12px;
  flex-shrink: 0;
}

.process-name {
  font-family: Consolas, 'Courier New', monospace;
  font-size: 12px;
}

.port-table-wrapper :deep(.el-table) {
  --el-table-bg-color: #2d2d2d;
  --el-table-tr-bg-color: #2d2d2d;
  --el-table-header-bg-color: #363636;
  --el-table-row-hover-bg-color: #363636;
  --el-table-border-color: #3d3d3d;
  --el-table-text-color: #c0c0c0;
  --el-table-header-text-color: #e5e5e5;
}

.port-table-wrapper :deep(.el-table__body tr.el-table__row--striped td.el-table__cell) {
  background-color: #292929;
}

.pagination-area :deep(.el-pagination) {
  --el-pagination-bg-color: transparent;
  --el-pagination-text-color: #a0a0a0;
  --el-pagination-button-bg-color: #2d2d2d;
  --el-pagination-hover-color: #409eff;
}
</style>
