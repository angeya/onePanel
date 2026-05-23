import { KEYWORD_HIGHLIGHT_RULES } from './terminalConstants'

/**
 * 已编译的高亮规则缓存，避免每次调用时重复编译正则。
 * 在模块加载时一次性初始化，后续直接使用。
 */
const COMPILED_RULES = KEYWORD_HIGHLIGHT_RULES.map((rule) => ({
  regex: new RegExp(rule.pattern, 'gi'),
  fg: rule.fg,
  bold: rule.bold
}))

/**
 * ANSI 转义序列正则，用于检测文本中已有的颜色控制码。
 * 匹配 CSI 序列（如 \x1b[31m、\x1b[0m、\x1b[1;32m 等）。
 */
const ANSI_ESCAPE_REGEX = /\x1b\[[0-9;]*[a-zA-Z]/g

/**
 * wrapMatchWithAnsi 将匹配到的文本用 ANSI 转义序列包裹。
 * 根据规则设置前景色和加粗样式，并在包裹结束后恢复默认。
 *
 * @param {string} text - 需要着色的原始文本
 * @param {number} fg - ANSI 前景色编号（31-36）
 * @param {boolean} bold - 是否加粗
 * @returns {string} 包裹了 ANSI 转义序列的文本
 */
function wrapMatchWithAnsi(text, fg, bold) {
  const beginCodes = []
  const endCodes = []

  if (bold) {
    beginCodes.push('1')
    endCodes.push('22')
  }
  beginCodes.push(String(fg))
  endCodes.push('39')

  const beginSeq = `\x1b[${beginCodes.join(';')}m`
  const endSeq = `\x1b[${endCodes.join(';')}m`

  return `${beginSeq}${text}${endSeq}`
}

/**
 * applyKeywordHighlight 对终端输出文本应用关键字高亮。
 * 核心思路参考 Tabby Highlight 插件的 ANSI 转义序列注入方式：
 * 1. 将文本按已有的 ANSI 转义序列分段，保留原有颜色码不被破坏
 * 2. 在不含 ANSI 控制码的纯文本段中，逐规则匹配关键字并注入颜色
 * 3. 含有 ANSI 控制码的段保持原样，不做任何修改
 *
 * @param {string} data - 终端原始输出文本
 * @returns {string} 注入了关键字高亮 ANSI 码的文本
 */
export function applyKeywordHighlight(data) {
  if (!data || typeof data !== 'string') {
    return data
  }

  const segments = data.split(ANSI_ESCAPE_REGEX)
  const escapes = []
  let match

  ANSI_ESCAPE_REGEX.lastIndex = 0
  while ((match = ANSI_ESCAPE_REGEX.exec(data)) !== null) {
    escapes.push(match[0])
  }

  if (segments.length === 1 && escapes.length === 0) {
    return highlightPlainText(data)
  }

  let result = ''
  for (let i = 0; i < segments.length; i++) {
    if (segments[i]) {
      result += highlightPlainText(segments[i])
    }
    if (i < escapes.length) {
      result += escapes[i]
    }
  }

  return result
}

/**
 * highlightPlainText 对不含 ANSI 转义序列的纯文本应用关键字高亮。
 * 逐规则匹配并注入 ANSI 颜色码，已着色的字符不会被后续规则覆盖。
 *
 * @param {string} text - 纯文本内容
 * @returns {string} 高亮处理后的文本
 */
function highlightPlainText(text) {
  if (!text) {
    return text
  }

  const colored = new Array(text.length).fill(false)
  const replacements = []

  for (const rule of COMPILED_RULES) {
    rule.regex.lastIndex = 0
    let match

    while ((match = rule.regex.exec(text)) !== null) {
      const start = match.index
      const end = start + match[0].length

      let hasOverlap = false
      for (let i = start; i < end; i++) {
        if (colored[i]) {
          hasOverlap = true
          break
        }
      }

      if (!hasOverlap) {
        for (let i = start; i < end; i++) {
          colored[i] = true
        }
        replacements.push({
          start,
          end,
          replacement: wrapMatchWithAnsi(match[0], rule.fg, rule.bold)
        })
      }
    }
  }

  if (replacements.length === 0) {
    return text
  }

  replacements.sort((a, b) => a.start - b.start)

  let result = ''
  let lastEnd = 0
  for (const rep of replacements) {
    result += text.slice(lastEnd, rep.start)
    result += rep.replacement
    lastEnd = rep.end
  }
  result += text.slice(lastEnd)

  return result
}
