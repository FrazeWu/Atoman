# CLAUDE.md — Atoman Project Context

## Project Overview

Atoman is an open platform for freedom of speech: blogs, RSS feed subscriptions, forum, debate, and music archive. Supports private/self-hosted deployment. License: GPL.

## Tech Stack

| Layer | Stack |
|---|---|
| Frontend | Vue 3.5 + Vite + TypeScript 5.9 + Pinia 3 + Vue Router 4 + Tailwind CSS v4 |
| Backend | Go + Gin + GORM + JWT |
| Database | PostgreSQL (prod) / SQLite (dev) |
| Storage | AWS S3-compatible (MinIO for dev) |
| Infra | Docker Compose + Nginx + supervisord |

## Module Status (Updated 2026-04-08)

### ✅ 完整实现的模块（基础CRUD完整，165+ API）

| Module | Status | Core Features | Enhancement Features |
|---|---|---|---|
| **Music** | ✅ Done | 上传、专辑、艺术家、播放器 | 纠正系统、批量处理、S3存储 |
| **Blog** | ✅ Done | Channel-Collection-Post三层架构 | 草稿/发布、置顶、书签分组、RSS生成 |
| **Forum** | ✅ Done | 分类、话题、嵌套回复 | 置顶/关闭、点赞、排序 |
| **Feed** | ✅ Done | RSS订阅、分组、时间线、已读 | OPML导入导出、健康检查、内部订阅 |
| **Debate** | ✅ Done | 话题、论点树、投票 | 引用系统、投票历史、结论功能 |
| **User/Auth** | ✅ Done | 登录注册、邮件验证 | 关注系统、用户设置、通知 |

### 📋 计划中的模块（未开始）

| Module | Status | Description |
|---|---|---|
| review | 📋 Planned | 音乐/电影/书籍/游戏评分系统 |
| timeline | 📋 Planned | 历史事件年表可视化 |
| reading | 📋 Planned | 电子书阅读器 + 笔记 |
| P2P video/music | 📋 Planned | IPFS去中心化分享 |

## Development Commands

```bash
# Frontend (from web/)
bun install
bun run dev        # hot-reload dev server
bun run build      # production build + type-check
bun run type-check
bun run lint
bun run format

# Backend (from server/)
go build ./...
go run cmd/start_server/main.go
go run cmd/create_admin/main.go
```

## Code Style

- **Prettier**: no semicolons, single quotes, 100-char line width
- **Vue SFC order**: `<template>` → `<script setup lang="ts">` → `<style scoped>`
- **Import order**: Vue core → third-party → `@/` aliases → relative
- **Components**: PascalCase. Stores: `use` + Name + `Store`. Routes: lowercase.
- **Pinia**: Setup Store syntax only (Composition API style)
- **Router**: Home = direct import; all others = lazy `() => import(...)`
- **UI**: Custom `A*` components (`ABtn`, `AModal`, `AEmpty`, `APageHeader`) — **NO Naive UI**

## Design System

Minimalist Archive aesthetic (极简主义 + 档案馆美学). Feedly-inspired layout.

- Colors: pure black `#000` + pure white `#fff`, gray scale, green (status), red (danger)
- Typography: `font-black tracking-tighter` for titles, `font-black uppercase tracking-widest text-xs` for labels
- Cards: `border-2 border-black`, hard drop shadows `shadow-[10px_10px_0px_0px_rgba(0,0,0,1)]`
- Covers: `filter:grayscale(100%)`. No soft shadows. No icons.
- Topbar hover: underline only. Other elements: black/white inversion.

## Pending Work (Priority Order)

### ✅ 已完成（2026-04-08）
1. ~~Fix TypeScript errors~~ ✅ Done (2026-04-07)
2. ~~Verify debate module completeness~~ ✅ Done (2026-04-08)
3. ~~Feed keyboard shortcuts (j/k/o/m/s/r)~~ ✅ Done (2026-04-08)

### 🔴 第一优先级：Feed 模块实用增强（选定）

#### 1. ~~健康检查UI补全（1天）~~ ✅ Done (2026-04-26)
**状态**: 已完成
- 在订阅列表显示健康状态图标（✓ ⚠ ✕）
- 添加"检查所有健康"按钮
- 悬停显示错误信息和检查时间
- 文件：`web/src/stores/feed.ts`, `web/src/views/feed/FeedView.vue`

#### 2. ~~重复文章检测（1天）~~ ✅ Done (2026-04-26)
**状态**: 已完成
- 标题相似度算法（Levenshtein Distance）
- URL规范化比较
- 标记重复并可选隐藏
- 文件：`server/internal/service/duplicate_detector.go`, `server/internal/model/feed.go`, `web/src/views/feed/FeedView.vue`

#### 3. ~~阅读统计图表（1天）~~ ✅ Done (2026-04-26)
**状态**: 已完成
- 后端聚合统计API
- Chart.js图表展示
- 按订阅源分类统计
- 文件：`server/internal/handlers/feed_handler.go`, `web/src/views/feed/FeedStatsView.vue`, `web/src/router.ts`

#### 4. 规则过滤系统（2天）
**功能**: 关键词黑白名单 + 自动操作
- FilterRule CRUD
- 正则表达式支持
- 自动标记已读/删除/收藏
- 文件：`server/internal/model/feed.go`, `server/internal/service/filter_engine.go`, `web/src/views/feed/FilterRulesView.vue`

#### 5. 性能优化（1天）
**功能**: 虚拟滚动 + 图片懒加载
- 集成 vue-virtual-scroller
- IntersectionObserver 懒加载
- 文件：`web/src/views/feed/FeedView.vue`

### 🟡 第二优先级：已有模块增强

#### Feed Q3 增强（中优先级）
- 全文抓取（summary-only feeds）- 2天
- RSS文章评论功能 - 1天
- 分享功能（链接/社交媒体）- 1天
- 离线阅读（PWA）- 2-3天
- 推荐系统 - 2-3天

#### Blog 增强
- 全文搜索（跨Post）- 1天
- 标签系统 - 1天
- 推荐系统 - 2-3天

#### Music 增强
- 播放列表功能 - 1-2天
- 歌词同步显示（LRC）- 1天

### 🟢 第三优先级：全新模块

#### Review 模块（2-3天）⭐ 推荐优先
**基础功能**:
- Review CRUD（评论创建、编辑、删除、列表）
- 评分系统（1-10分）
- 短评/长评支持
- 分类（music/movie/book/game）
- 探索页面

**技术实现**:
- 后端：`server/internal/handlers/review_handler.go`, `server/internal/model/review.go`
- 前端：`web/src/views/review/`
- 数据库模型：`Review(id, user_id, item_type, item_name, rating, short_review, long_review)`

#### Timeline 模块（3-4天）
**基础功能**:
- Event CRUD
- 时间轴可视化展示
- 事件分类、地理位置
- 人物关联

**技术实现**:
- 后端：`server/internal/handlers/timeline_handler.go`, `server/internal/model/timeline.go`
- 前端：`web/src/views/timeline/`（需要时间轴可视化组件）
- 数据库模型：`TimelineEvent`, `TimelinePerson`

#### Reading 模块（4-5天）
**基础功能**:
- Book CRUD
- 阅读进度追踪
- 书签、笔记系统
- 书架管理

**技术实现**:
- 后端：`server/internal/handlers/reading_handler.go`, `server/internal/model/reading.go`
- 前端：`web/src/views/reading/`（需要集成 epub.js）
- 数据库模型：`Book`, `ReadingProgress`, `BookNote`

### 🔵 第四优先级：跨模块功能

- 全站搜索（统一搜索入口）- 2-3天
- 高级权限系统（角色管理）- 3-4天
- 通知中心增强（推送、邮件）- 2天
- 主题系统（暗黑模式、自定义配色）- 2-3天

### 🟣 第五优先级：Feed Q4 功能

- Newsletter订阅（email → RSS）- 3天
- API开放平台（OAuth + 限流）- 5-7天
- 主题定制 - 2-3天

## Key References

- Project guidelines: `AGENTS.md`
- Plans & decisions: `.sisyphus/plans/`, `.sisyphus/decisions/`
- Feature docs: `doc/rss-feature-analysis.md`, `doc/feed-completed-features.md`
