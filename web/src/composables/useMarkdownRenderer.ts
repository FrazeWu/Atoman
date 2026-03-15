import { marked } from 'marked'
import { markedHighlight } from 'marked-highlight'
import markedKatex from 'marked-katex-extension'
import hljs from 'highlight.js'

// Configure marked once, reuse the same instance
const renderer = new marked.Renderer()

// Override heading to add id anchors for TOC navigation
renderer.heading = function ({ text, depth }) {
  const id = text
    .toLowerCase()
    .replace(/[^\w\u4e00-\u9fa5]+/g, '-')
    .replace(/^-|-$/g, '')
  return `<h${depth} id="${id}">${text}</h${depth}>\n`
}

marked.use(
  markedHighlight({
    langPrefix: 'hljs language-',
    highlight(code, lang) {
      const language = hljs.getLanguage(lang) ? lang : 'plaintext'
      return hljs.highlight(code, { language }).value
    },
  }),
)

marked.use(
  markedKatex({
    throwOnError: false,
    displayMode: false,
  }),
)

marked.use({ renderer })

export function useMarkdownRenderer() {
  function renderMarkdown(content: string): string {
    if (!content) return ''
    try {
      return marked(content) as string
    } catch {
      return `<pre>${content}</pre>`
    }
  }

  return { renderMarkdown }
}
