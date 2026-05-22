import { computed } from 'vue'

/**
 * useTerminalCommandExecution 统一管理根组件层面的终端命令投递策略。
 * 负责决定复用当前终端还是新建终端，并等待终端就绪后批量发送命令。
 */
export function useTerminalCommandExecution({
  tabs,
  activeTabId,
  defaultShell,
  addTerminalTab,
  sendCommand
}) {
  const activeTab = computed(() => tabs.value.find(tab => tab.id === activeTabId.value))

  /**
   * sendLinesToTerminal 按顺序向指定终端页签发送多条命令。
   */
  const sendLinesToTerminal = (tabId, commandLines) => {
    commandLines.forEach((line) => {
      sendCommand(tabId, line)
    })
  }

  /**
   * createTerminalAndRunCommands 创建终端并等待 terminal-ready 后发送命令。
   * 这样可以避免终端实例尚未初始化时命令丢失。
   */
  const createTerminalAndRunCommands = (commandLines, title = '') => {
    const tabId = addTerminalTab(defaultShell.value, title)
    if (!tabId) {
      return
    }

    const handleReady = (event) => {
      if (event.detail.tabId !== tabId) {
        return
      }

      window.removeEventListener('terminal-ready', handleReady)
      sendLinesToTerminal(tabId, commandLines)
    }

    window.addEventListener('terminal-ready', handleReady)
  }

  /**
   * normalizeCommandLines 清洗命令数组并根据工作目录补全前置 cd 命令。
   */
  const normalizeCommandLines = (commandLines, workDir = '') => {
    const lines = (commandLines || [])
      .filter(line => line && line.trim())
      .map(line => line.trim())

    if (lines.length === 0) {
      return []
    }

    return workDir ? [`cd ${workDir}`, ...lines] : lines
  }

  /**
   * executeShortcutCommand 按当前上下文执行快捷命令。
   * 如果当前已经是终端页签且允许复用，则直接复用当前终端。
   */
  const executeShortcutCommand = ({
    commandLines,
    commandName,
    workDir = '',
    forceNewTerminal = false
  }) => {
    const finalLines = normalizeCommandLines(commandLines, workDir)
    if (finalLines.length === 0) {
      return
    }

    const shouldUseCurrentTerminal = !forceNewTerminal && activeTab.value?.type === 'terminal'
    if (shouldUseCurrentTerminal) {
      sendLinesToTerminal(activeTab.value.id, finalLines)
      return
    }

    createTerminalAndRunCommands(finalLines, forceNewTerminal ? commandName : '')
  }

  /**
   * handleTerminalCommand 执行单条即时命令。
   */
  const handleTerminalCommand = (command) => {
    const line = command?.trim()
    if (!line) {
      return
    }
    executeShortcutCommand({ commandLines: [line] })
  }

  return {
    executeShortcutCommand,
    handleTerminalCommand
  }
}
