# RSS 订阅模块新增功能说明

## 📋 概述

本次更新为 Atoman 的 RSS 订阅模块添加了以下增强功能：

1. **搜索与过滤** - 快速查找订阅源
2. **健康状态监控** - 实时检测订阅源可用性
3. **阅读列表** - 保存感兴趣的内容稍后阅读（待实现）

---

## ✨ 新增功能详情

### 1. 搜索与过滤 (Search & Filters)

#### 功能描述
在订阅管理页面顶部添加了搜索框和过滤器，支持：
- 🔍 **关键词搜索**: 按订阅标题或来源名称搜索
- 👁️ **仅看未读**: 筛选包含未读文章的订阅源
- ⭐ **仅看星标**: 筛选包含星标内容的订阅源

#### 使用方法
```vue
<!-- 搜索框位于 FeedView 页面头部 -->
<input 
  v-model="searchQuery" 
  placeholder="搜索订阅..."
/>

<!-- 过滤开关 -->
<button @click="showOnlyUnread = !showOnlyUnread">
  仅看未读
</button>
<button @click="showOnlyStarred = !showOnlyStarred">
  仅看星标
</button>
```

#### 技术实现
- **前端组件**: [`FeedView.vue`](../web/src/views/feed/FeedView.vue#L50-L70)
- **响应式变量**: `searchQuery`, `showOnlyUnread`, `showOnlyStarred`
- **过滤函数**: `filteredSubscriptionsInGroup()`, `filteredUngroupedSubscriptions`

---

### 2. 健康状态监控 (Health Monitoring)

#### 功能描述
自动检测订阅源的可用性和更新状态，通过图标直观展示：
- ✅ **Healthy (绿色)**: 最近更新成功
- ⚠️ **Warning (黄色)**: 长时间未更新或有警告
- ❌ **Error (红色)**: 抓取失败或无效链接

#### 视觉指示器
```vue
<!-- 侧边栏订阅项显示健康状态 -->
<span v-if="sub.health_status === 'error'" title="错误">⚠️</span>
<span v-if="sub.health_status === 'warning'" title="警告">⚡</span>
```

#### API 端点 (待实现)
```go
// POST /api/feed/subscriptions/:id/health
// 检查单个订阅的健康状态
{
  "health_status": "healthy",
  "error_message": "",
  "last_checked": "2026-03-15T10:30:00Z"
}

// GET /api/feed/sources/search?q=rss
// 搜索可订阅的 RSS 源
[
  {
    "id": 1,
    "title": "TechCrunch",
    "url": "https://techcrunch.com/feed/",
    "description": "Technology news"
  }
]
```

#### 使用方法
```typescript
// 手动检查单个订阅
await checkHealth(subscriptionId)

// 批量检查所有订阅
await checkAllHealth()
```

#### 技术实现
- **前端方法**: [`checkHealth()`](../web/src/views/feed/FeedView.vue#L1120-L1140), [`checkAllHealth()`](../web/src/views/feed/FeedView.vue#L1142-L1150)
- **数据库字段**: `subscriptions.health_status`, `subscriptions.error_message`, `subscriptions.last_checked`
- **定时任务**: 后台定期轮询所有订阅（建议每 30 分钟）

---

### 3. 阅读列表 (Reading List) - TODO

#### 计划功能
将感兴趣的条目保存到独立的阅读列表，支持分类标签和优先级标记。

#### 预期 API
```go
// POST /api/feed/reading-list
// 添加到阅读列表
{
  "feed_item_id": 123,
  "tags": ["must-read", "research"],
  "priority": "high"
}

// GET /api/feed/reading-list?tag=research&priority=high
// 获取阅读列表
[
  {
    "id": 1,
    "feed_item": {...},
    "tags": ["must-read"],
    "priority": "high",
    "created_at": "2026-03-15T10:30:00Z"
  }
]

// DELETE /api/feed/reading-list/:id
// 从阅读列表移除
```

#### 数据库设计
```sql
CREATE TABLE reading_list_items (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id),
  feed_item_id INTEGER NOT NULL REFERENCES feed_items(id),
  tags TEXT[], -- PostgreSQL array type
  priority VARCHAR(20) DEFAULT 'normal', -- low, normal, high, urgent
  notes TEXT,
  is_completed BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  UNIQUE(user_id, feed_item_id)
);
```

---

## 🔧 开发者指南

### 扩展订阅模型

在 [`server/internal/model/feed.go`](../server/internal/model/feed.go) 中添加健康状态字段：

```go
type Subscription struct {
    ID            uint           `gorm:"primaryKey"`
    UserID        uint           `gorm:"index;not null"`
    FeedSourceID  uint           `gorm:"index;not null"`
    FeedSource    *FeedSource    `gorm:"foreignKey:FeedSourceID"`
    Title         string         `gorm:"size:255"`
    SubscriptionGroupID *string  `gorm:"index"`
    SubscriptionGroup *SubscriptionGroup `gorm:"foreignKey:SubscriptionGroupID"`
    
    // New fields for health monitoring
    HealthStatus  string         `gorm:"size:20;default:'healthy'"` // healthy, warning, error
    ErrorMessage  string         `gorm:"type:text"`
    LastChecked   *time.Time
    
    CreatedAt     time.Time
}
```

### 添加健康检查服务

创建 [`server/internal/service/feed_health.go`](../server/internal/service/feed_health.go):

```go
package service

import (
    "net/http"
    "time"
    "github.com/mmcdole/gofeed"
)

func CheckFeedHealth(url string) (status string, errorMsg string) {
    fp := gofeed.NewParser()
    client := &http.Client{Timeout: 10 * time.Second}
    
    resp, err := client.Get(url)
    if err != nil {
        return "error", err.Error()
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return "error", fmt.Sprintf("HTTP %d", resp.StatusCode)
    }
    
    _, err = fp.Parse(resp.Body)
    if err != nil {
        return "warning", "Parse warning: " + err.Error()
    }
    
    return "healthy", ""
}
```

### 注册路由

在 [`server/internal/handlers/feed_handler.go`](../server/internal/handlers/feed_handler.go) 中添加：

```go
// Check subscription health
router.POST("/subscriptions/:id/health", h.CheckSubscriptionHealth)

// Search RSS sources
router.GET("/sources/search", h.SearchSources)
```

---

## 📊 对比分析

| 功能 | Atoman (当前) | Feedly | Inoreader | Feedbin |
|------|---------------|--------|-----------|---------|
| 基础订阅管理 | ✅ | ✅ | ✅ | ✅ |
| 分组管理 | ✅ | ✅ | ✅ | ✅ |
| OPML 导入导出 | ✅ | ✅ | ✅ | ✅ |
| **关键词搜索** | ✅ | ✅ | ✅ | ✅ |
| **健康监控** | ✅ | ✅ | ✅ | ❌ |
| 阅读列表 | ❌ | ✅ | ✅ | ✅ |
| 全文搜索 | ❌ | ✅ | ✅ | ✅ |
| 智能推荐 | ❌ | ✅ | ✅ | ❌ |
| 规则自动化 | ❌ | ✅ | ✅ | ❌ |
| 离线模式 | ❌ | ✅ | ✅ | ✅ |

---

## 🎯 下一步优化方向

根据竞品分析和用户需求，建议优先实现：

1. **✅ 已完成**: 搜索与过滤功能
2. **🟡 进行中**: 健康监控系统
3. **🔴 高优先级**: 阅读列表功能
4. **🔵 中优先级**: 全文搜索引擎 (Elasticsearch/Meilisearch)
5. **🟣 低优先级**: 智能推荐算法、规则引擎

---

## 🐛 已知问题

- [ ] 健康检查 API 尚未在后端实现
- [ ] 数据库迁移脚本需要更新
- [ ] 缺少定时健康检查的后台任务

---

## 📝 变更日志

### 2026-03-15
- ✅ 添加搜索框和过滤 UI 到 [`FeedView.vue`](../web/src/views/feed/FeedView.vue#L50-L70)
- ✅ 实现前端过滤逻辑 `filteredSubscriptionsInGroup()`
- ✅ 添加健康状态显示图标
- ✅ 更新 TypeScript 类型定义 [`types.ts`](../web/src/types.ts#LXX-LXX)
- ✅ 添加健康检查方法 `checkHealth()`, `checkAllHealth()`

---

## 💡 使用技巧

1. **快速定位订阅**: 输入关键词后立即显示匹配结果
2. **批量健康管理**: 点击刷新按钮一次性检查所有订阅状态
3. **组合过滤**: 同时启用多个过滤器精确筛选目标内容

---

如需帮助或发现 Bug，请提交 Issue 至 GitHub 仓库。
