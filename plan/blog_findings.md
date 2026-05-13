# 发现与决策

## 需求
重新设计博客模块交互逻辑，基于「频道 = 独立创作身份」的新模型。

## 数据模型（最终确认版）

### 三层结构
```
用户账号（消费身份）
  ├── 评论 / 收藏 / 订阅  →  以用户身份
  └── 持有多个频道

频道（创作身份）
  ├── 独立主页 /channel/:slug
  ├── 独立 RSS
  ├── 默认合集（新内容自动落入）
  └── 其他合集（可引用任意内容，跨频道、跨类型）

内容（文章 / 视频 / 音乐）
  ├── 归属频道（不归属用户）
  ├── 首发即进入默认合集
  └── 可被多个频道协作发布（多频道共署名）
```

### 协作发布规则
- 邀请方式：生成邀请链接（含文章ID + token）
- 对方接受后以其频道身份加入协作
- 编辑：实时协同（Yjs + WebSocket），宽权限
- 发布权：仅创建者频道
- 版本历史：每次保存自动快照，可查 diff，可恢复

## 改动范围

### 🔴 必须改
- 文章归属从 user_id → channel_id
- Channel 模型新增 slug 字段
- Topbar 频道切换组件
- 用户主页改为频道列表

### 🟡 新增功能
- 协作邀请链接 + 接受流程
- 版本历史 UI
- 合集「引用他人内容」入口
- 频道主页重新设计

### 🟢 小修
- PostDetail 书签状态初始化 bug（bookmarked ref 未在 onMounted 初始化）
- ChannelView「合集主页」文案错误（应为「频道主页」）

## 技术评估
| 功能 | 复杂度 | 预估工时 |
|------|--------|---------|
| 邀请链接 + 接受流程 | 低 | 1天 |
| 协作者列表展示 | 低 | 0.5天 |
| 版本历史 + 恢复 | 中 | 2天 |
| 实时协同（Yjs + WS） | 高 | 3-5天 |
| 发布权限控制 | 低 | 0.5天 |

## Markdown 编辑器需求（2026-05-09）
- 需要两种编辑模式：SV（分屏源码/预览）和 WYSIWYG。
- 需要支持在内容中嵌入站内视频、音乐，以及外部媒体扩展能力。
- 需要写作辅助：大纲、字数统计、阅读时长、专注模式；不需要 AI 辅助。
- 协作要求高：后续要做邀请协作、版本历史、Yjs 实时协同。
- 当前实现是 `web/src/components/blog/MarkdownEditor.vue`，基于 Vditor IR 模式；`web/src/views/blog/PostEditorView.vue` 已有三栏编辑器壳和右侧设置面板。

## Markdown 编辑器候选结论
- **Tiptap**：最适合未来方向。WYSIWYG 强、Vue 集成成熟、扩展站内内容节点容易、Yjs 协作能力最好；但 Markdown 是 first-party beta，需要自己补齐 SV 模式与 Markdown round-trip 测试。
- **Toast UI Editor**：最符合“现在就要 SV + WYSIWYG 双模式”的现成需求，切换和分屏能力成熟；但协作能力弱，后续做 Yjs 实时协同不如 Tiptap 顺手。
- **Milkdown**：Markdown-first，理念契合，支持 Yjs；但更偏框架型，现成产品感和资料生态通常不如 Tiptap/Toast UI 直接。
- **Vditor**：当前已接入，三模式齐全，Markdown 能力强；但样式定制成本高，实时协作生态弱，不适合作为长期协作编辑基础。
- **Editor.md**：偏旧，不建议新项目继续投入。

## Tiptap 节点设计（站内引用）

### 统一原则
- 三种引用都做成 **块级原子节点**：`musicEmbed`、`videoEmbed`、`postEmbed`。
- **持久化 attrs 只保留 `id`**；标题、封面、作者、时长等展示数据运行时拉取，不写回 Markdown。
- 节点应设置为：`group: 'block'`、`atom: true`、`selectable: true`、`draggable: true`、`isolating: true`。

### 1) musicEmbed
- `type`: `musicEmbed`
- `attrs`: `{ id: string }`
- `parseHTML`: 识别 `<atoman-music data-id="...">`
- `renderHTML`: 输出 `<atoman-music data-id="..."></atoman-music>`
- `NodeView`: 拉取音乐元数据并渲染播放器卡片；失败时显示错误卡片。

### 2) videoEmbed
- `type`: `videoEmbed`
- `attrs`: `{ id: string }`
- `parseHTML`: 识别 `<atoman-video data-id="...">`
- `renderHTML`: 输出 `<atoman-video data-id="..."></atoman-video>`
- `NodeView`: 拉取视频元数据并渲染视频卡片；失败时显示错误卡片。

### 3) postEmbed
- `type`: `postEmbed`
- `attrs`: `{ id: string }`
- `parseHTML`: 识别 `<atoman-post data-id="...">`
- `renderHTML`: 输出 `<atoman-post data-id="..."></atoman-post>`
- `NodeView`: 拉取文章摘要并渲染引用卡片；失败时显示错误卡片。

## Markdown 序列化与回读规则

### 源码模式（Markdown → Tiptap）
推荐使用 **remark-parse + remark-directive** 识别自定义块：
1. 先把 Markdown 解析成 mdast。
2. 遇到 `containerDirective` 且 `name` 属于 `music | video | post` 时，校验 `attributes.id`。
3. 合法时转成语义 HTML 占位标签：
   - `music` → `<atoman-music data-id="..."></atoman-music>`
   - `video` → `<atoman-video data-id="..."></atoman-video>`
   - `post` → `<atoman-post data-id="..."></atoman-post>`
4. 再把整篇内容转成 HTML，交给 Tiptap `setContent()`。
5. Tiptap 通过各节点的 `parseHTML` 把这些标签收进自定义节点。

### 编辑模式（Tiptap JSON → Markdown）
推荐做自定义序列化层，而不是仅靠默认 Markdown 导出：
1. 遍历 Tiptap JSON。
2. 普通节点走默认 Markdown serializer。
3. 自定义节点按类型输出固定语法：
   - `musicEmbed` → `:::music{id="UUID"}` + 换行 + `:::`
   - `videoEmbed` → `:::video{id="UUID"}` + 换行 + `:::`
   - `postEmbed` → `:::post{id="UUID"}` + 换行 + `:::`
4. 块级节点前后各保留一个空行，保证 round-trip 稳定。

### 容错策略
- 缺少 `id` 或 `id` 非法：**不要转成自定义节点**，保留原始 Markdown 文本，并在源码模式给 lint/错误提示。
- `id` 合法但内容不存在/无权限：节点仍保留，NodeView 显示降级卡片（如“音乐不存在或不可见”）。
- 未知 directive（如 `:::channel`）第一阶段保持原样，不吞掉内容。

### 设计取舍
- “这是引用” 的判断依据是 **directive 语法**，不是 UUID 本身。
- 链接和引用语义分开：
  - 普通跳转：`[title](/post/xxx)`
  - 卡片引用：`:::post{id="xxx"}`

## 最小实现切片：先只做 postEmbed
目标：先验证 `post` 引用的整条链路，再复制到 `music` / `video`。

### 最小链路
1. 阅读态支持 `:::post{id="UUID"}`` 渲染为文章卡片。
2. Tiptap 支持 `postEmbed` 自定义块节点。
3. Markdown 源码模式能把 directive 解析回 `postEmbed`。
4. Tiptap 导出 Markdown 时能稳定输出 `:::post{id="UUID"}`。

### postEmbed 最小节点草图
- `name`: `postEmbed`
- `group`: `block`
- `atom`: `true`
- `selectable`: `true`
- `draggable`: `true`
- `isolating`: `true`
- `attrs`: `{ id: { default: '' } }`
- `parseHTML`: `tag: 'atoman-post[data-id]'`
- `renderHTML`: `['atoman-post', { 'data-id': id }]`
- `NodeView`: Vue 组件，根据 `id` 拉文章摘要卡片

### 最小序列化规则
- Markdown → Tiptap：`:::post{id="UUID"}` → `<atoman-post data-id="UUID"></atoman-post>` → `postEmbed`
- Tiptap → Markdown：`postEmbed(attrs.id)` → `:::post{id="UUID"}\n:::`

## 具体改动文件与优先级

### P0：必须先改
1. `web/src/composables/useMarkdownRenderer.ts`
   - 增加 directive 解析，使阅读态支持 `post` 引用。
2. `web/src/components/blog/editor/markdown/parseMarkdownToHtml.ts`
   - remark + directive 解析入口。
3. `web/src/components/blog/editor/markdown/serializeTiptapToMarkdown.ts`
   - Tiptap JSON → Markdown 自定义导出。
4. `web/src/components/blog/editor/tiptap/nodes/PostEmbed.ts`
   - `postEmbed` 节点定义。
5. `web/src/components/blog/editor/tiptap/nodeviews/PostEmbedView.vue`
   - 文章卡片 NodeView。

### P1：接入编辑器
6. `web/src/components/blog/TiptapMarkdownEditor.vue`
   - 新编辑器主组件。
7. `web/src/views/blog/PostEditorView.vue`
   - 替换当前 `MarkdownEditor` 接入点（现位于编辑区 `MarkdownEditor` 挂载处）。
8. `web/src/components/blog/MarkdownEditor.vue`
   - 删除旧 Vditor 实现或改为薄封装转发到 Tiptap 组件。

### P2：源码模式
9. `web/src/components/blog/editor/source/MarkdownSourceEditor.vue`
   - CodeMirror 6 承载 SV 左侧源码面板。
10. `web/src/components/blog/editor/tiptap/extensions.ts`
   - 汇总 StarterKit、自定义节点、图片、表格等扩展。

### P3：后续复制扩展
11. `web/src/components/blog/editor/tiptap/nodes/MusicEmbed.ts`
12. `web/src/components/blog/editor/tiptap/nodes/VideoEmbed.ts`
13. `web/src/components/blog/editor/tiptap/nodeviews/MusicEmbedView.vue`
14. `web/src/components/blog/editor/tiptap/nodeviews/VideoEmbedView.vue`

## 资源
- `web/src/views/blog/BlogHomeView.vue`
- `web/src/views/blog/ChannelView.vue`
- `web/src/views/blog/PostDetailView.vue`
- `web/src/views/blog/PostEditorView.vue`
- `web/src/components/blog/MarkdownEditor.vue`
- `web/src/router.ts`