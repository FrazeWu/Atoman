import { marked, type Token } from 'marked'
import { markedHighlight } from 'marked-highlight'
import markedKatex from 'marked-katex-extension'
import hljs from 'highlight.js'

type EmbedData = {
  id: string
  title: string
  summary?: string
  meta?: string
  href?: string
}

type RenderMarkdownOptions = {
  postEmbeds?: Record<string, EmbedData>
  musicEmbeds?: Record<string, EmbedData>
  videoEmbeds?: Record<string, EmbedData>
}

// Configure marked once, reuse the same instance
const renderer = new marked.Renderer()

renderer.heading = function ({ text, depth }) {
  const id = text
    .toLowerCase()
    .replace(/[^\w一-龥]+/g, '-')
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

export function parseBlocks(markdown: string): Token[] {
  return marked.lexer(markdown)
}

export function renderToken(token: Token): string {
  try {
    return marked.parser([token] as Token[]) as string
  } catch {
    return `<pre>${escapeHtml((token as any).raw || '')}</pre>`
  }
}

export function renderInline(markdown: string): string {
  try {
    return marked.parseInline(markdown) as string
  } catch {
    return escapeHtml(markdown)
  }
}

export function lexInline(text: string): Token[] {
  try {
    return (marked.Lexer as any).lexInline(text) as Token[]
  } catch {
    return [{ type: 'text', raw: text, text } as any]
  }
}

export function escapeHtml(str: string): string {
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

function renderEmbedCard(kind: 'post' | 'music' | 'video', embed: EmbedData, missing = false): string {
  const labelMap = {
    post: '文章引用',
    music: '音乐引用',
    video: '视频引用',
  } as const

  const title = escapeHtml(embed.title || labelMap[kind])
  const summary = escapeHtml(
    embed.summary || (kind === 'video' ? '该视频暂时不可见，点击尝试打开。' : '该引用内容暂无摘要'),
  )
  const meta = embed.meta ? escapeHtml(embed.meta) : ''
  const href = escapeHtml(embed.href || '#')

  return [
    `<div class="atoman-post-embed atoman-post-embed--${kind}${missing ? ' atoman-post-embed--missing' : ''}">`,
    `  <a class="atoman-post-embed__link" href="${href}">`,
    `    <div class="atoman-post-embed__label">${labelMap[kind]}</div>`,
    `    <div class="atoman-post-embed__title">${title}</div>`,
    `    <div class="atoman-post-embed__summary">${summary}</div>`,
    meta ? `    <div class="atoman-post-embed__meta">${meta}</div>` : '',
    '  </a>',
    '</div>',
  ].filter(Boolean).join('\n')
}

function replaceDirective(
  content: string,
  kind: 'post' | 'music' | 'video',
  pattern: RegExp,
  embeds: Record<string, EmbedData> | undefined,
  fallbackHref: (id: string) => string,
): string {
  return content.replace(pattern, (_match, id: string) => {
    const embed = embeds?.[id]
    if (!embed) {
      return renderEmbedCard(kind, { id, title: labelText(kind), href: fallbackHref(id) }, true)
    }

    return renderEmbedCard(kind, embed)
  })
}

function labelText(kind: 'post' | 'music' | 'video') {
  return {
    post: '引用文章',
    music: '引用音乐',
    video: '引用视频',
  }[kind]
}

function preprocessDirectives(content: string, options?: RenderMarkdownOptions): string {
  let next = replaceDirective(
    content,
    'post',
    /:::post\{id="([0-9a-fA-F-]{36})"\}\s*:::/g,
    options?.postEmbeds,
    (id) => `/post/${id}`,
  )

  next = replaceDirective(
    next,
    'music',
    /:::music\{id="([0-9a-fA-F-]{36})"\}\s*:::/g,
    options?.musicEmbeds,
    (id) => `/music/albums/${id}`,
  )

  next = replaceDirective(
    next,
    'video',
    /:::video\{id="([0-9a-fA-F-]{36})"\}\s*:::/g,
    options?.videoEmbeds,
    (id) => `#video-${id}`,
  )

  return next
}

export function useMarkdownRenderer() {
  function renderMarkdown(content: string, options?: RenderMarkdownOptions): string {
    if (!content) return ''
    try {
      return marked(preprocessDirectives(content, options)) as string
    } catch {
      return `<pre>${escapeHtml(content)}</pre>`
    }
  }

  return { renderMarkdown }
}
