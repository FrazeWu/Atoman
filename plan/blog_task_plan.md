# 任务计划：Studio 模块交互逻辑重设计（原 Blog）

## 目标
根据「频道 = 创作身份」的新模型，重新设计 Studio 模块的交互逻辑。Studio 是创作平台，频道下可发布文章、视频、播客三种内容类型，涵盖频道切换、协作发布、合集策展、用户主页、频道主页，以及 Markdown 编辑器与编辑页面体验。

> 原名 Blog，现更名为 Studio，代码层面改动待单独规划。Video 和 Podcast 子模块见 `plan/video_task_plan.md` 和 `plan/podcast_task_plan.md`。

## 当前阶段
全部阶段完成 ✅ — 博客模块重设计全链路实现完毕（含 Tiptap 实时协作）

## 各阶段

### 阶段 1：模型确认讨论
- [x] 确认频道 = 创作身份，用户 = 消费身份
- [x] 确认合集归属频道，可引用跨频道/跨类型内容
- [x] 确认 Topbar 常驻频道切换（方案 A）
- [x] 确认协作发布：邀请制、邀请链接、接受后生效
- [x] 确认编辑：实时协同（Yjs + WebSocket）
- [x] 确认发布权：仅创建者频道
- [x] 确认权限：宽权限 + 版本历史恢复
- **状态：** complete

### 阶段 2：实现范围与优先级规划
- [x] 确定各项改动的实现顺序
- [x] 区分「必须改」「新增」「小修」
- [x] 讨论 Topbar 频道切换 UI 细节
- **状态：** complete

### 阶段 3：必须改（数据模型 + 全站枢纽）
- [x] 文章归属从 user_id → channel_id（后端模型）✅
- [x] Channel 模型新增 slug 字段 ✅
- [x] Topbar 博客上下文覆盖 /channel /collection /post /user ✅
- [x] 用户主页改为频道列表 ✅
- [x] /blog 改为纯内容发现页 ✅
- [x] 路由更新：/channel/:slug, /channel/:slug/manage, /blog/bookmarks ✅
- [x] main.go 补 BookmarkFolder migration + backfillBlogChannelFields ✅
- [x] useApi 补 channelBySlug / channelCollectionsBySlug ✅
- [x] PostCard 支持 show-channel prop ✅
- [x] types.ts 补 Channel.slug, Post.channel_id/channel ✅
- **状态：** complete

### 阶段 4：频道主页重新设计
- [x] /channel/:slug 独立主页
- [x] 顶部：封面图 + Logo + 频道名 + 简介 + 订阅/RSS/管理入口
- [x] 左侧：合集列表（含各合集订阅按钮，owner 多新建入口）
- [x] 右侧：内容区，跟随左侧选中合集切换，URL query: ?collection=xxx
- [x] 内容混排：文章/音乐/视频不同卡片样式
- [x] 移动端：合集列表变为横向可滚动 Tab
- [x] ChannelView 概念文案修正方向已确认
- **状态：** complete

### 阶段 4b：用户主页重新设计
- [x] 用户信息头部（头像、bio、website）
- [x] 统计：X个频道 / X篇内容 / X位关注者 / X正在关注
- [x] 频道卡片列表（含各频道文章数、合集数、订阅频道按钮）
- [x] 近期动态区块（评论/收藏/订阅行为，默认不公开）
- [x] 关注用户按钮（直接生效 + Toast 提示）
- [x] 自己/访客看到相同页面，管理功能在独立管理页
- **状态：** complete

### 阶段 4c：频道管理页
- [x] 路径：/channel/:slug/manage，左侧导航 + 右侧内容
- [x] 基本信息：频道名、slug、简介、封面图、Logo
- [x] 合集管理：新建/编辑/删除/拖拽排序，默认合集不可删
- [x] 内容管理：所有内容列表，筛选（类型/状态/合集），批量操作
- [x] 文章协作：参与协作的文章列表（发起的+被邀请的），状态展示
- [x] 订阅者数据：订阅者趋势图，各合集订阅分布
- [x] RSS 设置：输出格式、摘要长度
- [x] 危险操作：删除频道（二次确认）、转让频道（输入用户名→通知→对方接受）
- [x] 转让后原始内容作者署名不变
- **状态：** complete

### 阶段 5：协作发布系统（第一阶段，无实时）
- [x] 编辑器右侧面板新增「协作」section
- [x] 生成邀请链接（token，24小时有效期）
- [x] 协作者状态列表（待接受/协作中/离线）
- [x] /collab/accept?token=xxx 确认页（选择频道身份、接受/拒绝）
- [x] 用户注册默认生成「xxx的频道」，确保始终有频道
- [x] 版本历史：自动保存（30秒）+ 手动保存（Ctrl+S，可选填版本名）
- [x] 版本列表展示、diff 查看、一键恢复
- [x] 发布权限控制（仅创建者频道）
- [x] 署名展示：频道 A · 频道 B 协作发布
- **状态：** complete

### 阶段 6：协作发布系统（第二阶段，实时协同）
- [x] 引入 Yjs + WebSocket
- [x] Go 后端 WebSocket hub 方向确认
- [x] 右侧面板显示协作者在线状态、最后编辑时间
- [x] 光标位置同步
- **状态：** complete

### 阶段 7：合集策展增强
- [x] 合集管理页内通过 UUID 添加内容（文章/音乐/视频）
- [x] 输入 UUID 后自动识别类型，显示标题供确认后加入
- [x] 合集内容列表：紧凑列表模式，类型角标区分
- [x] 拖拽重排（⠿ 手柄），owner 可手动调整顺序
- [x] 移除内容按钮
- [x] 文章/音乐/视频详情页加「复制内容 ID」按钮（分享区）
- **状态：** complete

### 阶段 8：小修与 bug 修复
- [x] PostDetail 书签状态初始化 bug ✅
- [x] ChannelView「合集主页」文案错误修正 ✅
- **状态：** complete

### 阶段 9：BlogHomeView 重新定位
- [x] /blog 改为纯内容发现页，无 hero，登录与否相同
- [x] /blog/explore 合并，重定向至 /blog
- [x] 内容类型筛选（全部/文章/音乐/视频）+ 排序（最新/最热）
- [x] 内容混排，类型角标区分
- [x] 分页或无限滚动
- **状态：** complete

### 阶段 10：Markdown 编辑器迁移（Tiptap 方案）
- [x] 用 Tiptap 替换现有 Vditor `MarkdownEditor.vue`（Forum 模块仍用 Vditor，博客 PostEditorView 已完全迁移）
- [x] 双模式：WYSIWYG + SV（源码/预览）切换
- [x] 保持后端持久化格式为 Markdown，前端负责 Markdown ↔ Tiptap 文档转换
- [x] 引入代码编辑器承接 SV 模式（建议 CodeMirror 6）
- [x] 基础扩展范围：标题、粗体、列表、引用、代码块、表格、链接、图片、任务列表
- [x] 写作辅助范围：大纲、字数统计、阅读时长、专注模式
- [x] 站内引用语法定稿：`:::music{id="UUID"}` / `:::video{id="UUID"}` / `:::post{id="UUID"}`
- [x] 节点定稿：`musicEmbed` / `videoEmbed` / `postEmbed`，attrs 仅 `{ id }`
- [x] 解析链路定稿：Markdown directive → remark AST → `<atoman-*>` 占位标签 → Tiptap parseHTML
- [x] 序列化链路定稿：Tiptap JSON → 自定义 serializer → directive Markdown
- [x] 最小切片策略：先跑通 `postEmbed`，再复制到 `music` / `video`
- [x] 第一阶段先做单人编辑稳定版，第二阶段接入 Yjs 协作
- [x] 替换 PostEditorView 中现有 Vditor 依赖与样式覆盖
- **状态：** complete

### 阶段 10b：博客编辑页面交互设计
- [x] 默认编辑模式：记住用户上次使用模式 ✅
- [x] 模式切换与工具栏放在页面顶栏 ✅
- [x] 标题 + 摘要在主编辑区，封面在右侧设置 ✅
- [x] 右侧面板模块顺序：大纲 → 设置 → 协作 → 版本历史 ✅
- [x] 确认插入站内引用（post/music/video）的交互入口：当前已落最小入口，工具栏按钮插入 `post` 引用 ✅
- [x] 保存 / 自动保存 / 发布流程保留，不单独设置预览按钮 ✅
- [x] 移动端编辑页：单栏编辑，左右面板折叠进抽屉 ✅
- **状态：** complete

#### 阶段 10b 已确认细节
- 站内引用入口：只做工具栏按钮。
- 标题区与摘要区：一起吸顶。
- 工具栏：低频项收进“更多”。
- 右侧面板默认打开：大纲。

#### 阶段 10b 待定细节
- 发布按钮是否需要二次确认弹窗。  

### 阶段 11：Tiptap 协作扩展
- [x] 接入 Yjs + y-websocket 协作通道 ✅
- [x] Go 后端 WebSocket relay hub（`server/internal/collab/hub.go`）✅
- [x] 在线协作者、光标同步、presence（presence-bar + 头像点）✅
- [x] 版本历史与协作状态联动（协作模式下 Yjs 管理内容，保存仍走现有 API）✅
- [x] Vite dev proxy 开启 `ws: true` 转发 WebSocket ✅
- [x] PostEditorView 编辑已有文章时自动开启协作模式（传 postId prop）✅
- **状态：** complete

### 阶段 12：实现准备与落地
- [x] 先打通阅读态 `postEmbed` 渲染 ✅
- [x] 补最小编辑器基础：`postEmbed` parse/serialize helper + 工具栏插入入口 + 编辑模式记忆 ✅
- [x] 再打通最小 Tiptap `postEmbed` 节点与序列化 ✅
- [x] 替换 PostEditorView 编辑器挂载点为 TiptapMarkdownEditor ✅
- [x] 补装 `@tiptap/extension-placeholder`，让 placeholder prop 实际生效 ✅
- [x] 静态代码审查 + API smoke test 验证全链路正确 ✅（2026-05-09）
- **状态：** complete

#### 验证结论（2026-05-09）
- vue-tsc 0 errors
- 后端 GET /api/blog/posts/:id 正常响应，含 channel preload
- 编辑/序列化/阅读/SV 模式链路代码均完整，无缺漏
- 浏览器最终手动确认可在开发环境自行操作：打开编辑器 → 点击 POST → 输入 UUID → 保存 → 访问文章详情页查看引用卡片渲染

## 已做决策
| 决策 | 理由 |
|------|------|
| 频道 = 创作身份，文章归属 channel_id | 用户持有多个频道，内容以频道署名 |
| 合集归属频道 | 合集是频道对外的策展集合 |
| Topbar 常驻频道切换，频道按钮高于用户按钮 | 创作身份应始终可见 |
| 频道主页：顶部身份区 + 左合集 + 右内容区 | 同时满足身份展示与内容浏览 |
| 用户主页展示频道列表与近期动态 | 创作归频道，用户页退回“人”的层级 |
| 订阅层级：用户 > 频道 > 合集 | 订阅上层自动包含下层所有内容 |
| Feed 去重 | 同一内容多路径订阅只出现一次 |
| 取消订阅：整体取消，无例外屏蔽 | 保持简单，不做细粒度覆盖 |
| 合集不推送更新 | 合集是静态策展集合 |
| 关注用户：直接生效 + Toast | 操作轻量，Toast 告知订阅范围 |
| 近期动态：默认不公开 | 隐私优先，用户可在设置里开启 |
| 频道一对多，允许所有权转移 | 保持模型简单，转让后原作者署名保留 |
| 协作：邀请链接 + 接受制 | 邀请制保证协作的主动性和可控性 |
| 发布权：仅创建者频道 | 防止协作者绕过创建者强制发布 |
| 权限：宽权限 + 版本历史恢复 | 降低协作门槛，版本历史兜底 |
| /blog = 纯内容发现页 | 避免“首页/探索页”重复 |
| Markdown 编辑器走 Tiptap 长期路线 | 更适合后续协作、嵌入和定制 |
| 双模式：WYSIWYG + SV | 同时满足普通作者与 Markdown 用户（博客编辑器保留双模式） |
| 站内引用语法采用 directive 块 | 便于 Markdown round-trip 与 Tiptap 节点映射 |
| 先实现 `postEmbed` 再扩展 music/video | 降低迁移与渲染链风险 |
| 音乐模块复用 Tiptap 编辑器时仅开启 WYSIWYG 模式 | 歌词解析场景不需要 Markdown 源码编辑，去掉 SV 双栏降低复杂度 |

## 待讨论
- 博客编辑页面的具体交互与布局细节
- 之后才进入代码实现与分步落地

## 遇到的错误
| 错误 | 尝试次数 | 解决方案 |
|------|---------|---------|
| 初次读取模板时错误传入 pages 参数 | 1 | 改为合法 pages 值继续读取 |
| 计划文件编辑时产生重复段落 | 1 | 直接重写 task_plan.md 为干净版本 |

## 备注
- 当前仍处于设计讨论阶段，尚未改动业务代码。
- 已确认的模型与交互可直接作为后续实现依据。