# 发现与决策

## 需求
社区模块负责承载更自由、日常、长期的讨论交流，与辩论模块形成分工：一个偏日常交流，一个偏结构化论证。

## 产品目标
- 让用户能围绕兴趣、问题、公告、资源分享发起讨论。
- 让讨论既能实时活跃，也能长期沉淀。
- 让治理能力足够支撑开放社区环境。
- 让高质量帖子能被收藏、推荐、回流。

## 推荐对象模型

### 1. ForumCategory
建议字段：
- `id`
- `name`
- `slug`
- `description`
- `sort_order`
- `topic_count`
- `is_locked`

### 2. ForumTopic
建议字段：
- `id`
- `category_id`
- `title`
- `body`
- `author_id`
- `status`（open / closed / hidden）
- `is_pinned`
- `is_featured`
- `view_count`
- `reply_count`
- `last_reply_at`

### 3. ForumReply
建议字段：
- `id`
- `topic_id`
- `parent_id`
- `author_id`
- `body`
- `status`
- `like_count`

### 4. Reaction / Bookmark / Subscription
- Reaction：点赞/踩/表情反馈
- Bookmark：收藏帖子
- Subscription：订阅帖子或分区更新

### 5. ModerationAction
记录：
- 置顶
- 关闭
- 隐藏
- 删除软删除
- 标记精华
- 用户禁言

## 推荐页面结构

### 1. 社区首页
- 分区导航
- 热门帖子
- 最新帖子
- 精华帖子
- 关注分区动态

### 2. 分区页
- 分区简介
- 排序切换（最新 / 最热 / 未回复 / 精华）
- 帖子列表
- 发帖入口

### 3. 帖子详情页
- 标题、作者、时间、标签
- 正文
- 收藏/点赞/订阅操作
- 回复区（支持嵌套）
- 管理操作入口

### 4. 发帖页
- 标题
- 正文编辑器
- 分区选择
- 标签
- 附件（可选）

## API 规划建议
- `POST /api/forum/categories`
- `GET /api/forum/categories`
- `POST /api/forum/topics`
- `GET /api/forum/topics/:id`
- `PUT /api/forum/topics/:id`
- `POST /api/forum/topics/:id/replies`
- `POST /api/forum/replies/:id/replies`
- `POST /api/forum/topics/:id/reactions`
- `POST /api/forum/topics/:id/bookmark`
- `POST /api/forum/topics/:id/subscribe`
- `POST /api/forum/moderation/:id/action`

## 分阶段开发建议

### Phase A
- 分区管理
- 发帖
- 回复
- 排序列表

### Phase B
- 收藏/点赞/订阅
- 精华/置顶/关闭
- 举报与审核

### Phase C
- 搜索
- 标签聚合
- 通知联动
- 用户内容管理

## 风险与取舍
- **低质量灌水风险**：需要频率限制和内容治理。
- **嵌套回复层级风险**：层级过深会影响可读性。
- **排序争议**：最新和最热之间需要明确切换。
- **运营成本**：推荐和精华机制需要人工或半自动支持。

## 最推荐的首期范围
1. 分区页
2. 主题帖 CRUD
3. 嵌套回复
4. 点赞收藏
5. 置顶关闭
6. 举报入口