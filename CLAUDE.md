# CLAUDE.md — Atoman Project Context

## Project Overview

Atoman is an open platform for freedom of speech: Studio (blog/video/podcast), RSS feed subscriptions, forum, debate, and music archive. Supports private/self-hosted deployment. License: GPL.

## Tech Stack

| Layer | Stack |
|---|---|
| Frontend | Vue 3.5 + Vite + TypeScript 5.9 + Pinia 3 + Vue Router 4 + Tailwind CSS v4 |
| Backend | Go + Gin + GORM + JWT |
| Database | PostgreSQL (prod) / SQLite (dev) |
| Storage | AWS S3-compatible (MinIO for dev) |
| Infra | Docker Compose + Nginx + supervisord |

## Module Status

### ✅ 完整实现的模块（基础CRUD完整，165+ API）

| Module | Status | Core Features | Enhancement Features |
|---|---|---|---|
| **Music** | ✅ Done | 上传、专辑、艺术家、播放器 | 纠正系统、批量处理、S3存储 |
| **Studio** | ✅ Done (Blog) | Channel-Collection-Post三层架构，文章发布 | 草稿/发布、置顶、书签分组、RSS生成；待扩展：Video、Podcast |
| **Forum** | ✅ Done | 分类、话题、嵌套回复 | 置顶/关闭、点赞、排序 |
| **Feed** | ✅ Done | RSS订阅、分组、时间线、已读 | OPML导入导出、健康检查、内部订阅 |
| **Debate** | ✅ Done | 话题、论点树、投票 | 引用系统、投票历史、结论功能 |
| **User/Auth** | ✅ Done | 登录注册、邮件验证 | 关注系统、用户设置、通知 |

### 📋 计划中的模块（未开始）

| Module | Status | Description |
|---|---|---|
| review | 📋 Planned | 音乐/电影/书籍/游戏评分系统 |
| video | 📋 Planned | 视频发布（Studio 子模块），YouTube 布局，本地上传 + 外链 |
| podcast | 📋 Planned | 播客发布（Studio 子模块），复用频道体系，RSS 输出 |
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

## Development Planning

所有功能规划、任务进度、设计决策均以 `plan/` 目录下的 Markdown 文件为准，不在本文件维护。

### plan/ 目录结构

每个模块对应三个文件：

| 文件 | 用途 |
|---|---|
| `plan/<module>_task_plan.md` | 阶段划分、任务清单、当前进度、已做决策 |
| `plan/<module>_findings.md` | 研究发现、技术调研结论 |
| `plan/<module>_progress.md` | 会话日志、操作记录、遇到的错误 |

### 当前各模块 plan 文件

| 模块 | Task Plan |
|---|---|
| Studio (Blog→Studio) | `plan/blog_task_plan.md` |
| Video | `plan/video_task_plan.md` |
| Podcast | `plan/podcast_task_plan.md` |
| Forum | `plan/forum_task_plan.md` |
| Music | `plan/music_task_plan.md` |
| Debate | `plan/debate_task_plan.md` |
| Timeline | `plan/timeline_task_plan.md` |

### 工作规范

- **开始任何功能前**，先读对应模块的 `task_plan.md` 确认当前阶段
- **每个阶段完成后**，更新 `task_plan.md` 状态（`in_progress` → `complete`）并在 `progress.md` 记录操作日志
- **新模块**按照上述三文件结构在 `plan/` 下新建

## Key References

- Project guidelines: `AGENTS.md`
- Feature plans: `plan/` directory（见上）
- Legacy docs: `doc/rss-feature-analysis.md`, `doc/feed-completed-features.md`
