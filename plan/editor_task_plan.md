# 任务计划：编辑器重构 v2 — 统一 SV 模式

## 目标
去掉 wysiwyg 模式，所有场景统一使用 SV（Split View）模式。
左侧 CodeMirror 6 编辑 Markdown 源码，右侧实时预览。
博客保留 Yjs 实时协作，改用 y-codemirror.next 实现。

## 当前阶段
**v2 重构全部完成 ✅** — 2026/05/10

---

## 最终决策

| 问题 | 决策 |
|------|------|
| 模式 | 统一 sv（CodeMirror 6 + 预览），plain 保留 |
| wysiwyg | 移除 |
| 左侧编辑器 | CodeMirror 6（@codemirror/lang-markdown） |
| 协作 | Yjs + y-codemirror.next（替换 Tiptap Collaboration） |
| Tiptap | 全部移除 |
| 工具栏 | sticky top-0，浮在编辑区最上方，两侧内容在下方滚动 |
| 同步滚动 | CodeMirror scroll 事件监听，行号对齐，左右双向同步 |
| 图片：工具栏按钮 | 点击 → file input → 上传 S3 |
| 图片：粘贴 Ctrl+V | paste 事件检测 image file → 上传 S3 |
| 图片：拖拽 | drop 事件 → 上传 S3（保留现有逻辑） |
| 图片上传中 | 立即插入 `![上传中…]()` 占位，完成后替换为真实 URL |
| @提及 | 保留现有逻辑，适配 CodeMirror |
| plain 模式 | 保留原样（纯 textarea，辩论用） |

---

## 依赖变更

### 新增
```
@codemirror/view
@codemirror/state
@codemirror/lang-markdown
@codemirror/language
@codemirror/commands
y-codemirror.next
```

### 移除
```
@tiptap/core
@tiptap/vue-3
@tiptap/starter-kit
@tiptap/extension-*（所有）
@tiptap/pm
@tiptap/suggestion
y-websocket（确认 y-codemirror.next 是否自带）
```

---

## 实现阶段

### 阶段 1：调研 + 依赖准备
- [ ] 确认 y-codemirror.next 依赖链（是否含 y-websocket）
- [ ] 确认 CodeMirror scroll 同步 API（domEventHandlers vs updateListener）
- [ ] bun add @codemirror/* y-codemirror.next
- [ ] bun remove @tiptap/* yjs y-websocket（或保留 yjs）
- **状态：** pending

### 阶段 2：AEditor sv 模式重写
- [ ] 左侧替换 textarea → CodeMirror 6 EditorView
- [ ] 工具栏改为 sticky top-0，编辑区独立滚动
- [ ] 实现左→右同步滚动（行号对齐法）
- [ ] 实现右→左同步滚动
- [ ] 粘贴图片：paste 事件 → 检测 image → 上传 → 占位符替换
- [ ] 拖拽图片：drop 事件（保留现有逻辑，适配 CodeMirror）
- [ ] @提及：适配 CodeMirror（监听 update transaction 替代 textarea input）
- [ ] 博客 enableCollab：Yjs + y-codemirror.next
- **状态：** pending

### 阶段 3：博客迁移
- [ ] PostEditorView：去掉 mode="wysiwyg"，改用 mode="sv" enableCollab enableEmbeds
- [ ] embed 插入（POST/MUSIC/VIDEO）：改为插入特殊 MD 语法 `{{embed:post:uuid}}`
- [ ] 预览侧解析 embed 语法，渲染卡片
- **状态：** pending

### 阶段 4：Props 接口更新
- [ ] 移除 Props 中 postId（改为 collabRoomId 或保持）
- [ ] 移除 enableEmbeds wysiwyg 专属逻辑
- [ ] 更新所有调用方（论坛、辩论、音乐）确保无 wysiwyg prop
- **状态：** pending

### 阶段 5：清理 + 验证
- [ ] 删除 blog/editor/tiptap/ 目录（PostEmbed、MusicEmbed、VideoEmbed nodes）
- [ ] 删除 parseMarkdownToHtml、serializeTiptapToMarkdown（如不再需要）
- [ ] 删除 wysiwyg 相关 CSS
- [ ] vue-tsc 0 errors
- [ ] bun run build 成功
- [ ] bundle 分析：确认 tiptap chunk 消失
- **状态：** pending

---

## 组件 Props 接口（目标）

```ts
interface AEditorProps {
  modelValue: string
  mode: 'sv' | 'plain'
  placeholder?: string
  noBorder?: boolean
  // sv only
  enableImageUpload?: boolean   // 默认 true
  enableMentions?: boolean      // 默认 false
  enableEmbeds?: boolean        // 博客用，默认 false
  enableCollab?: boolean        // 博客用，需配合 collabRoomId
  collabRoomId?: string
}
```

## 各模块使用映射（目标）

| 模块 | mode | enableMentions | enableEmbeds | enableCollab |
|------|------|---------------|-------------|-------------|
| 博客 PostEditorView | sv | false | true | true |
| 论坛新话题/回复 | sv | true | false | false |
| 辩论论点/回复 | plain | false | false | false |
| 音乐讨论 | sv | true | false | false |

---

## SV 工具栏按钮（目标）

```
[ H2 ] [ H3 ] | [ B ] [ I ] [ S ] [ code ] [ Link ] [ Img ] | [ Quote ] [ ·List ] [ 1.List ] [ Table ] [ HR ]
博客额外：| [ POST ] [ MUSIC ] [ VIDEO ]
```

工具栏 sticky top-0，z-index 高于滚动内容。

---

## 滚动同步方案

```
左侧 CodeMirror scroll → 取 scrollTop / scrollHeight 比例
→ 右侧 preview.scrollTop = 比例 × preview.scrollHeight

右侧 preview scroll → 同理反向同步
（设 syncing flag 防止循环触发）
```

---

## 图片占位符替换方案

```
上传开始：
  在光标位置插入 `![上传中…]()`
  记录该字符串的起始 index

上传完成：
  在 CodeMirror state 中找到 `![上传中…]()` 的位置
  dispatch transaction 替换为 `![图片](url)`
```

多张并发上传：用唯一 ID 区分，如 `![上传中-abc123]()`

---

## 遇到的错误
（待记录）
