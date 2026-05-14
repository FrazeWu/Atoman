# Atoman 平台模块决策规格

**日期：** 2026-05-14  
**状态：** 已审阅确认  
**来源：** 2026-05-14 头脑风暴会话（plan/ 目录全模块评审）

---

## 一、背景与范围

本文档记录 Atoman 平台各模块经由需求合理性与可行性逐一讨论后达成的设计决策。  
覆盖模块：编辑器基线、Studio、Forum、Video、Podcast、Debate、Timeline、Music。  
所有决策以此文件为单一权威源，旧 `plan/` 目录文件中与本文冲突的内容以本文为准。

---

## 二、编辑器基线（全局）

所有内容输入场景统一使用 **SV（Split View）模式**：
- 左侧：CodeMirror 6 编辑 Markdown 源码
- 右侧：实时预览
- 博客（Studio）保留 Yjs + y-codemirror.next 实现实时协作
- **plain 模式保留**：辩论论点等轻量场景使用纯 textarea

**已废弃：**
- WYSIWYG 模式（Tiptap）完全移除
- Tiptap 相关依赖、节点、NodeView 全部清理

**参考实现文档：** [plan/editor_task_plan.md](../../plan/editor_task_plan.md)

---

## 三、Studio（原 Blog）

### 核心模型
- 频道 = 创作身份，内容归属 `channel_id`，不归 `user_id`
- 用户 = 消费身份，持有多个频道
- 合集归属频道，可引用跨频道/跨类型内容

### 协作发布
- 邀请制，生成 token 链接（24h 有效），接受后生效
- 发布权：仅创建者频道
- 版本历史：每次保存快照，可查 diff，可恢复

### 站内 Embed 语法（SV 基线下）
在 SV 模式中插入站内引用使用以下语法（与现有实现一致）：

```
:::post{id="UUID"}
:::

:::music{id="UUID"}
:::

:::video{id="UUID"}
:::
```

- 编辑侧：CodeMirror 工具栏 POST/MUSIC/VIDEO 按钮通过 `window.prompt` 弹窗输入 UUID 后插入
- 预览侧：`useMarkdownRenderer.ts` 通过正则解析 `:::post{id="..."}:::` 并渲染引用卡片
- 后端持久化：原始 directive 语法存入 Markdown 字段
- 阅读态：同预览侧，由 `useMarkdownRenderer` 渲染 `.atoman-post-embed` 卡片

### 发布流程
- 一键发布，**不加二次确认弹窗**

### 频道主页
- 分 Tab：全部 / 文章 / 视频 / 播客
- 独立 RSS：`/channel/:slug/rss/article`

---

## 四、Forum（社区）

### 编辑器策略
| 场景 | 模式 |
|------|------|
| 发帖（ForumTopic） | SV（Markdown） |
| 回复（ForumReply） | plain（纯文本 + 图片） |

### 嵌套回复
- 最多 **2 层**（顶层回复 depth=0，子回复 depth=1），后端拒绝 depth≥2 的创建请求（400）
- 子回复默认只显示前 **2 条**，其余折叠，显示"展开 N 条回复"按钮
- 想回复子回复时，只能发新顶层回复并 `@用户名`（不允许多级嵌套）

### 标签（Tags）
- 自由输入，无预设标签库，发帖时输入、回车或逗号分隔，显示为 chip
- 存储：`ForumTopic.Tags` 使用现有 `StringSlice` 类型（JSON text array）
- 过滤：`GET /api/forum/topics?tag=xxx`，后端用 `tags LIKE '%"xxx"%'`
- 前端：帖子卡片和详情显示标签 chip，点击跳转过滤列表

### 解决方案标记（Solved）
- `ForumTopic` 添加 `IsSolved bool`、`SolvedReplyID *uuid.UUID`
- `ForumReply` 添加 `IsSolved bool`（冗余，方便渲染）
- **手动标记**：楼主或管理员点击"✓ 标为解决" → `POST /api/forum/replies/:id/solve`
- **自动标记**：回复点赞数 ≥ `SiteSettings["forum.solved_auto_threshold"]`（默认 10）时自动触发
- **取消**：楼主或管理员可取消 → `DELETE /api/forum/replies/:id/solve`
- 一帖只有一个 Solved 标记，新标记自动替换旧标记
- 前端：被标记回复显示绿色 ✓ 徽章；帖子标题旁显示"已解决"pill；帖子卡片同步显示

### SiteSettings（站点配置表）
- 用于存储管理员可调整的全局配置参数
- 表结构：`key`（unique）、`value`（string）、`description`、`updated_at`
- 命名空间前缀区分模块：`forum.xxx`、`timeline.xxx` 等
- 管理员后台接口：`GET /api/admin/settings`、`PUT /api/admin/settings/:key`
- 首期配置项：`forum.solved_auto_threshold`（默认 `"10"`）

### 核心约束
- 不允许匿名发帖
- 分区一层结构，用户可申请新分区，管理员审核
- 排序：最新 / 最热 / 精华
- 不做声望系统（后续单独规划）
- 举报达阈值自动折叠
- 通知（@提及 + 回复通知）：**独立通知模块**，不在 Forum 首期范围
- 私信（DM）：**独立私信模块**，不在 Forum 首期范围

### 首期范围
分区页 → 主题帖 CRUD → 嵌套回复（2层上限）→ 点赞/收藏 → 置顶/关闭/精华 → 举报 → 标签 → 解决方案标记 → 分区申请

---

## 五、Video

### 核心模型
- 归属 Studio 频道体系
- 支持本地上传 + 外部链接（YouTube/Bilibili 等）
- 合集：**复用 Studio 合集体系**，视频可加入频道合集（与博客/播客统一，不单独建设）

### 推荐算法（首期）
混合策略：**同频道 60% + 同标签 40%**  
实现方式：简单打分排序，不依赖行为埋点，可用 SQL 实现

### 播放统计
| 指标 | 期次 |
|------|------|
| 播放量计数 | 首期 |
| 完播率 | 二期 |

### 独立 RSS
`/channel/:slug/rss/video`

### 待讨论（二期）
- 完播率采集与展示
- 播放统计详细 Dashboard

---

## 六、Podcast

### 核心模型
- Show = Studio 频道（`Channel`）
- Episode = 频道内播客内容（复用 Channel-Post 架构）
- 音频文件必须本地上传，无外链选项

### 封面策略
- 节目封面（Channel 封面）：**必填**
- 单集封面：**可选覆盖**，RSS `<itunes:image>` 优先用单集封面，fallback 到节目封面

### RSS 输出
- 路径：`/channel/:slug/rss/podcast`
- 含 `<enclosure>` 标签，符合播客规范
- 可被 Overcast、Apple Podcasts 等客户端订阅

### 阶段规划
| 功能 | 期次 |
|------|------|
| 发布与 RSS 合规 | 首期 |
| Feed 模块接入播客 RSS | 二期 |
| 跨节目收听队列 | 二期 |
| 时间章节（Chapters） | 不做 |

---

## 七、Debate（辩论）

### 核心模型
- 结构化论点树（非扁平评论流）
- 一个议题可配置多个立场（不强制正反二元）
- 证据独立建模（`DebateEvidence`），保留 `source_type` + `source_ref` 字段，**不做可信度分级**

### 权限与治理
| 角色 | 能力 |
|------|------|
| 游客 | 浏览（匿名可看） |
| 登录用户 | 发起议题、发表论点、投票 |
| 管理员 | 折叠、锁定、置顶、裁定争议 |

- **无主持人/裁判角色**，管理员承担全部治理职能
- 证据不做可信度标注，但 `source_ref` + `title` + `excerpt` 必须展示

### 首期范围
议题 CRUD → 多立场 → 论点树 → 投票表态 → 简单总结区 → 管理员折叠/锁定

### 阶段规划
| 功能 | 期次 |
|------|------|
| 论点树拖拽重组 | 二期 |
| 证据来源展示 | 首期（无可信度） |
| 高级筛选/跨模块引用 | 二期 |

---

## 八、Timeline（时间线）

### 核心模型
- Event 为核心实体
- 支持不完整与不确定日期（`date_precision: day / month / year / range / uncertain`）
- **首期支持 BCE（公元前）**：`start_date` / `end_date` 允许负值年份
- 人物、地点、来源独立建模，独立页面

### 用户提交流程（维基形式）
参照音乐模块维基规则的**轻量版**：
- 登录用户可创建/编辑事件条目
- 每次编辑生成版本快照（revision）
- 状态只分两态：**草稿（draft）/ 已发布（published）**
- 管理员可回滚版本、隐藏条目

> 首期不完全照搬 Music 的 Open/Confirmed/Disputed 三态，避免过重。

### 事件关系类型
首期仅支持"相关事件"（无类型），二期再做 因果 / 并行 / 前置 / 后续 等类型标注。

### 首期范围
事件 CRUD → 基础时间轴视图 → 事件详情页 → 人物/地点关联 → 来源列表 → 搜索筛选

### 阶段规划
| 功能 | 期次 |
|------|------|
| BCE 日期支持 | 首期 |
| 地图联动 | 二期 |
| 事件关系类型 | 二期 |
| 高级可视化/知识图谱 | 三期 |

---

## 九、Music（音乐档案库）

### 维基模型
- 专辑为基础对象，单曲 = type:single 的专辑
- 登录用户可直接编辑 Open / Disputed 条目
- 每次编辑生成 revision 快照，回滚通过"基于旧版本创建新版本"实现

### 条目状态
| 状态 | 说明 |
|------|------|
| Open | 任何登录用户可直接编辑 |
| Disputed | 有争议，继续开放编辑，详情页显示争议提示 |
| Confirmed | 管理员锁定，普通用户只能提修改建议 |

- **Disputed 标记：仅详情页显示**，列表页不显示，不干扰浏览
- **Confirmed 条目改动：** 普通用户通过"修改建议"入口提交变更描述（类 PR），由管理员决定是否合并；无自动合并，管理员为唯一裁定人

### Artist 维基化
与 Album 同步推进，走完全相同的维基流程与状态机。

### 权限
- 普通登录用户：编辑 Open/Disputed，对 Confirmed 提交建议
- 管理员：确认/重新开放/解决争议、裁定条目合并

---

## 十、实施优先级建议

基于依赖关系与功能密度，建议按以下顺序推进：

| 优先级 | 模块 | 原因 |
|--------|------|------|
| P0 | 编辑器 SV 统一（editor_task_plan v2） | 所有模块编辑器的前置依赖 |
| P1 | Forum | 规划最完整，进入实现成本最低 |
| P1 | Music（Artist 维基化补全） | 核心已完成，仅补 Artist 侧 |
| P2 | Video | Studio 频道体系已就绪，可快速接入 |
| P2 | Podcast | 与 Video 并行，共用频道体系 |
| P3 | Debate | 规划完整，需论点树 UI 设计 |
| P3 | Timeline | 数据模型最复杂，需先稳定 Event 表结构 |

---

## 十一、跨模块共用规范

1. **评论统一复用 Forum 格式**：plain 文本 + 图片，Video、Podcast、Music 讨论均使用
2. **合集统一复用 Studio 合集体系**：文章/视频/播客均可加入频道合集
3. **RSS 分类型独立输出**：每种内容类型有独立 RSS 路径
4. **播放条（底部 Player）**：音乐/播客共用，不重复建设
5. **嵌套回复深度限制**：Forum 与 Debate 均需设最大嵌套层级（建议 ≤ 5 层），防可读性崩溃

---

## 十二、遗留风险

| 风险 | 影响模块 | 应对 |
|------|---------|------|
| Studio embed 语法需独立预览解析器 | Studio | 实现时单独排期，不依赖 Tiptap |
| Timeline 维基治理首期轻量化，后续状态扩展可能有迁移成本 | Timeline | 状态字段预留扩展空间，不要 hard-code 两态 |
| Debate 无主持人，管理员治理压力集中 | Debate | 首期需配套管理员快捷折叠/锁定 UI，降低操作成本 |
| Video 推荐依赖标签质量 | Video | 首期按同频道兜底，标签缺失时 fallback 到全局热门 |
| Forum 嵌套回复层级无上限 | Forum | 实现时在后端或前端设置最大嵌套深度（建议 5 层） |
