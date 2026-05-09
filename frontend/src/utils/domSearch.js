/**
 * DOM 文本搜索工具
 * 用于在非终端 tab 页中搜索和高亮文本内容
 */

const HIGHLIGHT_TAG = 'search-highlight'
const HIGHLIGHT_ACTIVE_TAG = 'search-highlight-active'

/**
 * 在指定容器中搜索文本并高亮匹配项
 * @param {HTMLElement} container - 搜索的容器元素
 * @param {string} keyword - 搜索关键词
 * @returns {{ total: number, current: number }} 匹配总数和当前激活的索引
 */
export function searchInContainer(container, keyword) {
  clearHighlights(container)
  if (!keyword || !container) return { total: 0, current: 0 }

  const matches = []
  const walker = document.createTreeWalker(container, NodeFilter.SHOW_TEXT, {
    acceptNode: (node) => {
      if (node.parentElement?.closest('[contenteditable]')) return NodeFilter.FILTER_REJECT
      if (node.parentElement?.tagName === 'SCRIPT' || node.parentElement?.tagName === 'STYLE') return NodeFilter.FILTER_REJECT
      if (node.parentElement?.classList?.contains(HIGHLIGHT_TAG) || node.parentElement?.classList?.contains(HIGHLIGHT_ACTIVE_TAG)) return NodeFilter.FILTER_REJECT
      return NodeFilter.FILTER_ACCEPT
    }
  })

  const textNodes = []
  while (walker.nextNode()) {
    textNodes.push(walker.currentNode)
  }

  const lowerKeyword = keyword.toLowerCase()

  for (const node of textNodes) {
    const text = node.textContent
    const lowerText = text.toLowerCase()
    let startIndex = 0
    let offset = lowerText.indexOf(lowerKeyword, startIndex)

    while (offset !== -1) {
      matches.push({ node, offset, length: keyword.length })
      startIndex = offset + 1
      offset = lowerText.indexOf(lowerKeyword, startIndex)
    }
  }

  if (matches.length === 0) return { total: 0, current: 0 }

  const range = document.createRange()
  const highlightElements = []

  for (let i = matches.length - 1; i >= 0; i--) {
    const match = matches[i]
    try {
      range.setStart(match.node, match.offset)
      range.setEnd(match.node, match.offset + match.length)
      const span = document.createElement('span')
      span.className = HIGHLIGHT_TAG
      span.dataset.searchIndex = String(i)
      range.surroundContents(span)
      highlightElements.unshift(span)
    } catch {
      // 跨元素边界的情况，跳过
    }
  }

  if (highlightElements.length > 0) {
    highlightElements[0].className = HIGHLIGHT_ACTIVE_TAG
    highlightElements[0].scrollIntoView({ block: 'center', behavior: 'smooth' })
  }

  return { total: highlightElements.length, current: highlightElements.length > 0 ? 1 : 0 }
}

/**
 * 跳转到下一个匹配项
 * @param {HTMLElement} container - 搜索的容器元素
 * @returns {{ total: number, current: number } | null}
 */
export function findNextInContainer(container) {
  const active = container.querySelector(`.${HIGHLIGHT_ACTIVE_TAG}`)
  const highlights = container.querySelectorAll(`.${HIGHLIGHT_TAG}, .${HIGHLIGHT_ACTIVE_TAG}`)
  if (highlights.length === 0) return null

  let currentIndex = -1
  if (active) {
    currentIndex = parseInt(active.dataset.searchIndex)
    active.className = HIGHLIGHT_TAG
  }

  const nextIndex = (currentIndex + 1) % highlights.length
  highlights[nextIndex].className = HIGHLIGHT_ACTIVE_TAG
  highlights[nextIndex].scrollIntoView({ block: 'center', behavior: 'smooth' })

  return { total: highlights.length, current: nextIndex + 1 }
}

/**
 * 跳转到上一个匹配项
 * @param {HTMLElement} container - 搜索的容器元素
 * @returns {{ total: number, current: number } | null}
 */
export function findPrevInContainer(container) {
  const active = container.querySelector(`.${HIGHLIGHT_ACTIVE_TAG}`)
  const highlights = container.querySelectorAll(`.${HIGHLIGHT_TAG}, .${HIGHLIGHT_ACTIVE_TAG}`)
  if (highlights.length === 0) return null

  let currentIndex = 0
  if (active) {
    currentIndex = parseInt(active.dataset.searchIndex)
    active.className = HIGHLIGHT_TAG
  }

  const prevIndex = (currentIndex - 1 + highlights.length) % highlights.length
  highlights[prevIndex].className = HIGHLIGHT_ACTIVE_TAG
  highlights[prevIndex].scrollIntoView({ block: 'center', behavior: 'smooth' })

  return { total: highlights.length, current: prevIndex + 1 }
}

/**
 * 清除所有搜索高亮
 * @param {HTMLElement} container - 搜索的容器元素
 */
export function clearHighlights(container) {
  if (!container) return
  const highlights = container.querySelectorAll(`.${HIGHLIGHT_TAG}, .${HIGHLIGHT_ACTIVE_TAG}`)
  for (let i = highlights.length - 1; i >= 0; i--) {
    const span = highlights[i]
    const parent = span.parentNode
    while (span.firstChild) {
      parent.insertBefore(span.firstChild, span)
    }
    parent.removeChild(span)
    parent.normalize()
  }
}
