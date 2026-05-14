# Studio Spec

## Purpose
Studio 是 Atoman 的创作系统，核心不是“用户发帖”，而是“频道作为创作身份发布多类型内容”。频道承载创作归属、合集策展、协作发布与独立展示，用户账号则主要承担消费、互动与关注行为。

## Core Model

### Identity model
- 用户账号是消费身份。
- 频道是创作身份。
- 一个用户可以拥有多个频道。
- 内容归属 `channel_id`，而不是 `user_id`。

### Content model
- 频道可以发布文章、视频、播客三种内容。
- 内容首发归入默认合集。
- 合集属于频道，但可跨频道、跨类型引用内容。
- 内容发现页 `/blog` 负责浏览，不承担创作者管理职责。

### Collaboration model
- 协作通过邀请链接触发。
- 协作者以自己的频道身份加入协作。
- 发布权仅属于创建者频道。
- 所有编辑保留版本历史。
- 实时协作使用 Yjs + WebSocket。

## Core UX Rules

### Channel surfaces
- `/channel/:slug` 是独立频道主页。
- 频道主页结构：顶部身份区 + 左侧合集区 + 右侧内容区。
- `/channel/:slug/manage` 是独立管理页。
- 用户主页展示频道列表，而不是单纯的用户文章列表。

### Discovery and consumption
- `/blog` 是纯内容发现页。
- 内容类型支持文章 / 音乐 / 视频混排展示。
- 排序至少支持最新与最热。
- 频道切换入口应常驻顶部，创作身份始终可见。

### Editing
- 长期编辑器方向为 Tiptap。
- 保留 Markdown 持久化格式。
- 需要支持 WYSIWYG 与 SV 双模式。
- 站内引用使用 directive 语法映射到块级原子节点：`postEmbed`、`musicEmbed`、`videoEmbed`。

## Embedded Content Rules
- 嵌入节点的持久化 attrs 只保留 `id`。
- 标题、封面、作者等展示信息运行时拉取，不回写 Markdown。
- 阅读态与编辑态都必须支持 round-trip 稳定。
- `postEmbed` 是首个最小闭环，之后再复制到 `music` 与 `video`。

## Subscription and ownership rules
- 订阅层级：用户 > 频道 > 合集。
- 同一内容多路径订阅只出现一次。
- 合集本身不承担动态推送。
- 频道支持所有权转移，但原始作者署名不变。

## Durable Decisions
- 频道高于用户，是创作署名的主单位。
- `/blog` 定位为发现页，而不是创作者首页。
- 协作采取宽权限 + 版本历史恢复，而不是细粒度多人审批。
- 编辑器优先服务长期协作、站内嵌入与可扩展性，因此选择 Tiptap 路线。

## Related Features
- Video 与 Podcast 是 Studio 子模块，但保有各自独立的实施计划。
- 频道主页应支持文章、视频、播客分发与筛选。
