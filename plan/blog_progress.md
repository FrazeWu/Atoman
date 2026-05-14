# 进度日志

## 会话：2026-05-08（博客模块重设计讨论）

### 阶段 1：模型确认
- **状态：** complete
- 执行的操作：
  - 读取 BlogHomeView、ChannelView、PostDetailView、PostEditorView 源码
  - 与用户完成 4 轮问答，确认完整数据模型
- 核心结论：
  - 频道 = 创作身份（非分类容器）
  - 合集 = 频道下的策展集合，可跨频道/跨类型引用内容
  - Topbar 常驻频道切换
  - 协作：邀请链接制 + 实时协同（Yjs）+ 宽权限 + 版本历史
  - 发布权仅限创建者频道

### 阶段 2：实现规划
- **状态：** in_progress
- 下一步：讨论 Topbar 频道切换 UI 细节

## 测试结果
- 2026-05-09：`cd web && bun run type-check` ✅
- 2026-05-09：`cd server && go build ./...` ✅

## 最新实现进展
- 已完成博客 Phase 1 后端主轴：`Post.channel_id`、`Channel.slug`、迁移 backfill、路由与 CRUD 校验。
- 已完成博客前端基础路由与只读页第一轮重构：`/blog` 发现页、`/channel/:slug`、`/channel/:slug/manage`、用户页频道列表化。
- 已打通阅读态 `postEmbed` 最小闭环：`:::post{id="UUID"}` 在 `PostDetailView` 中会解析为引用卡片，并动态拉取文章标题/摘要/频道名。 
- 下一步进入编辑器切片：先做 `postEmbed` 的解析/序列化基础，再替换 `PostEditorView` 挂载点。
- 已补最小编辑器基础：
  - `web/src/components/blog/editor/markdown/parseMarkdownToHtml.ts`
  - `web/src/components/blog/editor/markdown/serializeTiptapToMarkdown.ts`
  - `MarkdownEditor.vue` 新增工具栏 `POST` 插入按钮
  - `PostEditorView.vue` 新增 `SV / WYSIWYG` 模式切换与 localStorage 记忆
- 已安装最小 Tiptap 依赖：`@tiptap/core`、`@tiptap/vue-3`、`@tiptap/starter-kit`、`@tiptap/pm`。
- 已新增：
  - `web/src/components/blog/editor/tiptap/nodes/PostEmbed.ts`
  - `web/src/components/blog/TiptapMarkdownEditor.vue`
- 已将 `PostEditorView.vue` 的编辑器挂载点从旧 `MarkdownEditor` 切换到 `TiptapMarkdownEditor`。
- 已验证：
  - `cd web && bun run type-check` ✅
  - `cd web && bun run build` ✅
- 未完成项：真实浏览器里的编辑/保存/阅读闭环验证；当前 Chrome DevTools MCP 被本机已有实例占用，需在可用浏览器会话中补测。  

## 错误日志
| 时间戳 | 错误 | 尝试次数 | 解决方案 |
|--------|------|---------|---------|
| 2026-05-08 | 读取模板时 pages 参数格式错误 | 1 | 修正后继续 |

## 五问重启检查
| 问题 | 答案 |
|------|------|
| 我在哪里？ | 阶段 2，实现规划讨论中 |
| 我要去哪里？ | 确认 Topbar 频道切换 UI，然后开始实现 |
| 目标是什么？ | 博客模块交互逻辑重设计 |
| 我学到了什么？ | 见 findings.md |
| 我做了什么？ | 完成模型讨论，写入 task_plan.md |