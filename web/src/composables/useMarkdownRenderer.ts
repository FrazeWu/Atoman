import { marked, type Token } from 'marked'
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

// ─── Standalone exports for IR engine ───────────────────────────────────────

/** Parse full markdown document into top-level block tokens */
export function parseBlocks(markdown: string): Token[] {
  return marked.lexer(markdown)
}

/** Render a single block token to HTML */
export function renderToken(token: Token): string {
  try {
    return marked.parser([token] as Token[]) as string
  } catch {
    return `<pre>${escapeHtml((token as any).raw || '')}</pre>`
  }
}

/** Render a single line of inline markdown to HTML */
export function renderInline(markdown: string): string {
  try {
    return marked.parseInline(markdown) as string
  } catch {
    return escapeHtml(markdown)
  }
}

/** Lex inline markdown into tokens for IR mixed rendering */
export function lexInline(text: string): Token[] {
  try {
    return (marked.Lexer as any).lexInline(text) as Token[]
  } catch {
    return [{ type: 'text', raw: text, text } as any]
  }
}

/** HTML-escape a string */
export function escapeHtml(str: string): string {
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

// ─── Composable (for components that need renderMarkdown) ────────────────────

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
