# Editor Spec

## Purpose
编辑器系统的目标是统一各模块的文本编辑体验，并以 Split View 为核心模型：左侧编辑 Markdown 源码，右侧实时预览。博客场景保留实时协作能力，论坛与音乐等场景复用统一编辑器能力，辩论仍可保留轻量 plain 模式。

## Core Editing Model

### Modes
- 主模式统一为 `sv`（Split View）。
- 保留 `plain` 作为轻量纯文本模式。
- `wysiwyg` 应移除，不再作为长期模式。

### Editing stack
- 左侧源码编辑器使用 CodeMirror 6。
- 右侧为实时预览。
- 博客协作使用 Yjs + `y-codemirror.next`。
- Tiptap 不再是统一编辑器内核。

## Functional Rules
- 工具栏应以 Markdown/SV 语义工作，而不是富文本节点语义。
- 图片上传支持工具栏上传、粘贴上传、拖拽上传。
- 上传过程中立即插入占位符，成功后替换为真实链接。
- `@` 提及能力保留，并适配 CodeMirror 输入事件模型。
- 博客需要支持 embed 特殊语法，如 `{{embed:post:uuid}}`。
- 预览侧必须能解析并渲染 embed 语法。

## Props Contract
```ts
interface AEditorProps {
  modelValue: string
  mode: 'sv' | 'plain'
  placeholder?: string
  noBorder?: boolean
  enableImageUpload?: boolean
  enableMentions?: boolean
  enableEmbeds?: boolean
  enableCollab?: boolean
  collabRoomId?: string
}
```

## Module Usage Rules
- 博客：`sv` + embeds + collab
- 论坛：`sv` + mentions
- 辩论：`plain`
- 音乐讨论：`sv` + mentions

## Durable Decisions
- 编辑器统一方向是 SV，不再并行维护 WYSIWYG。
- CodeMirror 6 是统一源码编辑基础。
- 博客的协作实现应迁移到 Yjs + `y-codemirror.next`。
- Tiptap 相关编辑器目录和依赖最终应退出统一编辑器系统。
- 工具栏应 sticky 固定在编辑区顶部。
- 左右滚动同步是核心体验要求，而不是可选增强。

## Boundaries
- `plain` 模式可继续存在，但只用于真正不需要复杂编辑能力的场景。
- embed、协作、图片上传等能力按模块选择启用，不要求所有调用方都打开。
