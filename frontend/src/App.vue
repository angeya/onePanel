<template>
  <div class="app-container">
    <el-container>
      <el-header class="app-header">
        <div class="header-content">
          <h1 class="app-title">onePanel 终端</h1>
        </div>
      </el-header>
      <el-main class="app-main">
        <Terminal />
      </el-main>
    </el-container>
  </div>
</template>

<script setup>
import Terminal from './components/Terminal.vue'
</script>

<style scoped>
.app-container {
  height: 100vh;
  width: 100%;
}

.app-header {
  background-color: #2b2b2b;
  border-bottom: 1px solid #3d3d3d;
  display: flex;
  align-items: center;
  padding: 0 20px;
  height: 50px !important;
}

.header-content {
  display: flex;
  align-items: center;
}

.app-title {
  margin: 0;
  font-size: 18px;
  font-weight: 500;
  color: #e5e5e5;
}

.app-main {
  padding: 15px;
  background-color: #1e1e1e;
  height: calc(100vh - 50px);
}
</style>
