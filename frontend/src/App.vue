<template>
  <div class="app-container">
    <el-container class="main-container">
      <el-aside width="200px" class="app-aside">
        <div class="aside-logo">
          <span class="logo-text">onePanel</span>
        </div>
        <el-menu
          :default-active="activeMenu"
          class="aside-menu"
          @select="handleMenuSelect"
          background-color="#1e1e1e"
          text-color="#a0a0a0"
          active-text-color="#409eff"
        >
          <el-menu-item index="terminal">
            <el-icon><Monitor /></el-icon>
            <span>终端</span>
          </el-menu-item>
          <el-menu-item index="apps">
            <el-icon><Grid /></el-icon>
            <span>我的应用</span>
          </el-menu-item>
          <el-menu-item index="shortcuts">
            <el-icon><Promotion /></el-icon>
            <span>快捷命令</span>
          </el-menu-item>
          <el-menu-item index="tools">
            <el-icon><SetUp /></el-icon>
            <span>实用工具</span>
          </el-menu-item>
        </el-menu>
      </el-aside>
      <el-main class="app-main">
        <TerminalPage v-if="activeMenu === 'terminal'" />
        <AppsPage v-else-if="activeMenu === 'apps'" />
        <ShortcutCmdPage v-else-if="activeMenu === 'shortcuts'" />
        <ToolsPage v-else-if="activeMenu === 'tools'" />
      </el-main>
    </el-container>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { Monitor, Grid, Promotion, SetUp } from '@element-plus/icons-vue'
import TerminalPage from './views/TerminalPage.vue'
import AppsPage from './views/AppsPage.vue'
import ShortcutCmdPage from './views/ShortcutCmdPage.vue'
import ToolsPage from './views/ToolsPage.vue'

const activeMenu = ref('terminal')

/**
 * 处理侧边栏菜单选择
 */
const handleMenuSelect = (index) => {
  activeMenu.value = index
}
</script>

<style scoped>
.app-container {
  height: 100vh;
  width: 100%;
  overflow: hidden;
}

.main-container {
  height: 100%;
}

.app-aside {
  background-color: #1e1e1e;
  border-right: 1px solid #2d2d2d;
  overflow: hidden;
}

.aside-logo {
  height: 50px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid #2d2d2d;
}

.logo-text {
  font-size: 18px;
  font-weight: 700;
  color: #409eff;
  letter-spacing: 1px;
}

.aside-menu {
  border-right: none;
}

.aside-menu .el-menu-item {
  height: 48px;
  line-height: 48px;
}

.aside-menu .el-menu-item.is-disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.app-main {
  padding: 0;
  background-color: #252526;
  overflow: hidden;
}
</style>
