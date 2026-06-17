import MarkdownIt from 'markdown-it'
import hljs from 'highlight.js/lib/core'
import go from 'highlight.js/lib/languages/go'
import javascript from 'highlight.js/lib/languages/javascript'
import typescript from 'highlight.js/lib/languages/typescript'
import python from 'highlight.js/lib/languages/python'
import java from 'highlight.js/lib/languages/java'
import bash from 'highlight.js/lib/languages/bash'
import json from 'highlight.js/lib/languages/json'
import xml from 'highlight.js/lib/languages/xml'
import css from 'highlight.js/lib/languages/css'
import sql from 'highlight.js/lib/languages/sql'
import yaml from 'highlight.js/lib/languages/yaml'
import 'highlight.js/styles/github.css'

// 按需注册常用语言，避免打包全部语言导致体积过大
hljs.registerLanguage('go', go)
hljs.registerLanguage('javascript', javascript)
hljs.registerLanguage('typescript', typescript)
hljs.registerLanguage('python', python)
hljs.registerLanguage('java', java)
hljs.registerLanguage('bash', bash)
hljs.registerLanguage('json', json)
hljs.registerLanguage('xml', xml)
hljs.registerLanguage('html', xml)
hljs.registerLanguage('css', css)
hljs.registerLanguage('sql', sql)
hljs.registerLanguage('yaml', yaml)

/**
 * 将标题文本转换为 URL 友好的锚点 ID
 * 支持中文、英文、数字，其他字符替换为 -
 * @param {string} text
 * @returns {string}
 */
const slugify = (text) => {
  return text
    .toLowerCase()
    .replace(/[^\w\u4e00-\u9fa5]+/g, '-')
    .replace(/^-+|-+$/g, '')
}

// 配置 markdown-it：启用自动链接、排版增强，但禁用原始 HTML 防止 XSS
// 代码块使用 highlight.js 进行语法高亮
const md = new MarkdownIt({
  html: false,
  linkify: true,
  typographer: true,
  breaks: true,
  highlight: (str, lang) => {
    if (lang && hljs.getLanguage(lang)) {
      try {
        return `<pre class="hljs"><code>${hljs.highlight(str, { language: lang, ignoreIllegals: true }).value}</code></pre>`
      } catch (__) {
        // 忽略高亮失败，使用默认转义
      }
    }
    return `<pre class="hljs"><code>${md.utils.escapeHtml(str)}</code></pre>`
  }
})

// 自定义标题渲染：为每个 h1-h6 添加 id 锚点，便于目录跳转
md.renderer.rules.heading_open = (tokens, idx, options, env, self) => {
  const token = tokens[idx]
  const inlineToken = tokens[idx + 1]
  if (inlineToken && inlineToken.type === 'inline') {
    const id = slugify(inlineToken.content)
    token.attrSet('id', id)
  }
  return self.renderToken(tokens, idx, options)
}

/**
 * 渲染 Markdown 为 HTML
 * @param {string} text
 * @returns {string}
 */
export const renderMarkdown = (text) => {
  if (!text) return ''
  return md.render(text)
}

/**
 * 根据 Markdown 内容生成目录（Table of Contents）
 * @param {string} text
 * @returns {Array<{level: number, title: string, id: string}>}
 */
export const generateTOC = (text) => {
  if (!text) return []
  const tokens = md.parse(text, {})
  const toc = []
  for (let i = 0; i < tokens.length; i++) {
    const token = tokens[i]
    if (token.type === 'heading_open') {
      const level = parseInt(token.tag.slice(1))
      const inlineToken = tokens[i + 1]
      if (inlineToken && inlineToken.type === 'inline') {
        const title = inlineToken.content
        toc.push({ level, title, id: slugify(title) })
      }
    }
  }
  return toc
}

/**
 * 去除 Markdown 标记，提取纯文本摘要
 * @param {string} text
 * @param {number} maxLength
 * @returns {string}
 */
export const stripMarkdown = (text, maxLength = 120) => {
  if (!text) return ''
  const html = md.render(text)
  const tmp = document.createElement('div')
  tmp.innerHTML = html
  let plain = tmp.textContent || tmp.innerText || ''
  plain = plain.replace(/\s+/g, ' ').trim()
  if (plain.length > maxLength) {
    plain = plain.slice(0, maxLength) + '...'
  }
  return plain
}
