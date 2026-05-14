# 通知系统设计规格

**日期：** 2026-05-15  
**状态：** 已审阅确认  
**模块：** Notification（全平台统一通知）

---

## 一、背景与范围

通知系统是全平台统一模块，所有业务模块（Forum、Debate、Timeline、Blog 等）统一通过 `NotificationService` 发通知，不在各业务模块内各自实现。

首期触发来源为 Forum 模块，后续扩展无需改通知模块主体。

**不在范围内：**
- 私信（DM）—— 独立模块，见 `2026-05-15-dm-design.md`
- 邮件通知 —— 不做
- 赞通知以外的聚合 —— 二期

---

## 二、用户界面

### 入口

导航栏使用**文字按钮**（不使用图标），显示「收件箱」，通知 + 私信合计未读数 > 0 时在旁边显示角标数值。

```
[收件箱 5]   ← 导航栏右侧，5 = 通知未读 + 私信未读
```

点击后跳转至统一收件箱页 `/inbox`，不使用下拉面板。

### 收件箱页 `/inbox`

左右分栏布局：左侧列表，右侧详情/对话区。

左侧 Tab（4 个）：
- **回复我的** —— `forum_reply` 类型通知
- **给我的赞** —— `forum_like` 类型通知（合并聚合）
- **@我的** &nbsp;&nbsp;&nbsp;&nbsp;—— `forum_mention` 类型通知
- **私信** &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;—— DM 会话列表（来自私信模块）

每条通知：actor 用户名 + 动作描述 + 内容摘要 + 时间  
点击单条：右侧显示详情，同时标记为已读（通知类型跳转到来源内容）  
一键全部已读按钮（当前 Tab 范围内）  
分页加载（首期不做无限滚动）

---

## 三、数据模型

```go
type Notification struct {
    Base
    RecipientID uuid.UUID  `gorm:"type:uuid;not null;index"`
    Recipient   *User      `gorm:"foreignKey:RecipientID;references:UUID"`
    ActorID     *uuid.UUID `gorm:"type:uuid;index"`  // nil = 系统通知
    Actor       *User      `gorm:"foreignKey:ActorID;references:UUID"`
    Type        string     `gorm:"not null"`          // 见下方类型表
    SourceType  string     `gorm:"not null"`          // "forum_reply" | "forum_topic" | ...
    SourceID    uuid.UUID  `gorm:"type:uuid;not null"`
    Meta        JSONB      `gorm:"type:jsonb"`        // 展示用冗余字段，见下
    ReadAt      *time.Time
}

func (Notification) TableName() string { return "notifications" }
```

**索引：**
- `(recipient_id, read_at)` —— 未读列表查询
- `(recipient_id, source_type, source_id)` UNIQUE —— 去重用

### Meta JSONB 结构（按 source_type）

`forum_reply` / `forum_mention`：
```json
{
  "topic_id": "uuid",
  "topic_title": "Vue 3 性能优化",
  "reply_excerpt": "你提到的 shallowRef..."
}
```

`forum_solved`：
```json
{
  "topic_id": "uuid",
  "topic_title": "Vue 3 性能优化"
}
```

Meta 在创建通知时写入，通知列表查询不需要 JOIN 业务表。

`forum_like`（合并型）额外字段：
```json
{
  "topic_id": "uuid",
  "topic_title": "Vue 3 性能优化",
  "actor_count": 5,
  "recent_actors": ["alice", "bob", "carol"]
}
```
合并规则：`(recipient_id, source_type, source_id)` 已存在 `forum_like` 通知时，更新 Meta（actor_count + 1，recent_actors 保留最近 3 人），重置 `read_at = NULL`。

---

## 四、通知类型表

| Type | 触发时机 | 接收者 | 优先级 | 合并策略 |
|------|---------|--------|--------|----------|
| `forum_reply` | 有人回复了我的帖子或我的回复 | 帖主 / 被回复者 | 2 | 不合并 |
| `forum_mention` | 回复正文中 `@username` 提及了我 | 被提及者 | 3（最高）| 不合并 |
| `forum_solved` | 我的回复被标为解决方案 | 回复作者 | 1 | 不合并 |
| `forum_like` | 有人点赞了我的帖子或回复 | 被赞者 | 0 | **合并**（同 source 累加）|

**去重规则（优先级去重）：**  
同一 `(recipient_id, source_type, source_id)` 组合已存在通知时，若新通知优先级 ≥ 现有通知，则更新 `type` 为更高优先级并重置 `read_at = NULL`；否则忽略。

**自通知屏蔽：** `actor_id == recipient_id` 时不创建通知。

---

## 五、@提及解析

后端在保存 `ForumReply` 时，由 `NotificationService` 用正则扫描 `content`：

```go
var mentionRegexp = regexp.MustCompile(`@([A-Za-z0-9_-]{1,32})`)
```

- 查询 `Users` 表确认用户名存在
- 过滤掉 actor 自身
- 每个有效 mention 触发一条 `forum_mention` 通知

前端不需要在提交时附带 mentions 列表，后端是唯一可信来源。

---

## 六、实时推送（WebSocket）

### 架构

复用并扩展现有 `collab/hub.go`（gorilla/websocket）：

| | 现有 collab hub | 新增 user hub |
|---|---|---|
| 路由键 | `roomID`（博客 UUID）| `userID`（UUID）|
| 广播对象 | 同 room 所有连接 | 单个用户的所有连接 |
| URL | `/ws/collab/:room` | `/ws/user` |
| 鉴权 | 无（博客协作） | JWT Bearer token |

前端登录后建立一个 `/ws/user` 连接，通知与后续私信模块共用同一连接，通过消息类型字段区分：

```json
{ "event": "notification", "data": { ...通知对象... } }
{ "event": "dm", "data": { ...私信对象... } }
```

### WS 断线回退

WS 断开时，前端每 60s 轮询 `GET /api/notifications/unread-count`，重连后恢复 WS 推送。

---

## 七、API 端点

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/notifications` | 获取通知列表，`?page=1&type=reply\|like\|mention` |
| GET | `/api/notifications/unread-count` | 获取通知未读数（WS 断线回退 / 初始加载）|
| PUT | `/api/notifications/:id/read` | 标记单条已读 |
| PUT | `/api/notifications/read-all` | 一键全部已读（当前 type 范围）|
| WS  | `/ws/user` | JWT 鉴权，接收实时通知推送 |

所有端点需 JWT 鉴权（`AuthRequired` middleware）。

---

## 八、前端架构

### notificationStore (Pinia)

```typescript
interface NotificationState {
  unreadCount: number
  notifications: Notification[]
  loading: boolean
}
```

职责：
- 应用启动时调用 `fetchUnreadCount()` 初始化角标
- WS 连接由 inboxStore 统一建立（共用 `/ws/user`）
- WS 收到 `notification` 事件：`unreadCount++`，若当前在 `/inbox` 通知 Tab 则 prepend 列表
- 提供 `markRead(id)`、`markAllRead(type)` action

### inboxStore（协调层）

```typescript
totalUnread = notificationStore.unreadCount + dmStore.unreadCount
```

导航栏「收件箱 N」绑定 `totalUnread`，由 inboxStore 统一管理 `/ws/user` WS 连接生命周期，分发 `notification` / `dm` 事件到对应 store。

### 路由

```
/inbox     InboxPage.vue（4 Tab：回复我的 / 给我的赞 / @我的 / 私信）
```

### 跳转逻辑（source_type → URL）

```typescript
function notificationTargetUrl(n: Notification): string {
  switch (n.source_type) {
    case 'forum_reply':
      return `/forum/topics/${n.meta.topic_id}#reply-${n.source_id}`
    case 'forum_topic':
      return `/forum/topics/${n.source_id}`
    default:
      return '/inbox'
  }
}
```

---

## 九、扩展设计

未来新模块接入通知，只需：
1. 在 `NotificationService` 增加对应的 `CreateXxxNotification(...)` 方法
2. 在 `Type` 中注册新类型（如 `debate_reply`、`blog_comment`）
3. 在前端 `notificationTargetUrl` 中新增 case

不需要修改数据库表结构，不需要修改 WS 基础设施。

---

## 十、首期范围

- [x] `notifications` 表 + 迁移
- [x] `NotificationService`（创建、去重、@解析、赞合并）
- [x] Forum 回复/解决方案/点赞事件触发通知
- [x] `/ws/user` WebSocket 端点（user hub，由 inboxStore 管理）
- [x] REST API（列表、未读数、已读）
- [x] 前端 notificationStore + inboxStore
- [x] 导航栏「收件箱 N」文字按钮（合并通知 + 私信未读）
- [x] `/inbox` 页面（4 Tab：回复我的 / 给我的赞 / @我的 / 私信）

**二期（不在本次范围）：**
- 邮件通知
- 通知偏好设置（哪类通知开/关）
- 回复通知聚合（多人回复同一帖合并显示，目前每条独立）
