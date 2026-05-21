# [OPEN] Debug Session: ssh-auto-login

## 问题描述
- 症状：点击服务器列表中的会话后，会新开终端 Tab，但不会自动执行 SSH 登录命令。
- 期望：新开的终端 Tab 在可写入后自动发送登录命令，进入 SSH 登录流程。

## 当前假设
1. `addTerminalTab` 没有返回 tabId，导致后续发送命令时目标 tabId 为 `undefined`。
2. `ServerListPanel.vue` 使用固定 `setTimeout(300)` 等待终端初始化，存在时序竞争，命令发送时终端尚未 ready。
3. `TerminalTab.vue` 监听 `terminal-send-command` 时，PTY 尚未启动或 `ptyId` 为空，事件被静默丢弃。
4. `tab-ssh-connect` 与 `terminal-send-command` 依赖 window 事件，但事件触发早于新 tab 组件挂载完成。

## 已观察到的静态证据
- `useAppTabs.js` 中 `addTerminalTab()` 只设置 `activeTabId`，未返回新建的 tab id。
- `ServerListPanel.vue` 中 `const tabId = addTerminalTab(defaultShell.value)` 后立即使用 `tabId` 派发事件和发送命令。
- `ServerListPanel.vue` 依赖 `setTimeout(300)` 处理登录时序，没有显式 ready 信号。

## 下一步
- 优先修复 tabId 丢失问题。
- 再把“延时等待”改为“终端 ready 后再发送命令”的稳定机制。
- 修复后执行前端构建验证。
