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
- **UI**: Custom `A*` components (`ABtn`, `AModal`, `AEmpty`, `APageHeader`, `AInput`, `ATextarea`, `ASelect`, `ADropdown`, `APopover`) — **NO Naive UI**

## Design System

Minimalist Archive aesthetic (极简主义 + 档案馆美学). Feedly-inspired layout.

- Colors: pure black `#000` + pure white `#fff`, gray scale, green (status), red (danger)
- Typography: `font-black tracking-tighter` for titles, `font-black uppercase tracking-widest text-xs` for labels
- Cards: `border-2 border-black`, hard drop shadows `shadow-[10px_10px_0px_0px_rgba(0,0,0,1)]`
- Covers: `filter:grayscale(100%)`. No soft shadows. No icons.
- Topbar hover: underline only. Other elements: black/white inversion.

### Unified Component Usage

为避免样式漂移，后续前端开发默认遵循以下规则：

#### 1. 优先使用统一基础组件

- 按钮：使用 `ABtn`
- 输入框：使用 `AInput`
- 多行输入：使用 `ATextarea`
- 下拉选择：使用 `ASelect`
- 弹窗：使用 `AModal`
- 菜单弹层：使用 `ADropdown`
- 信息弹层：使用 `APopover`
- 确认弹窗：使用 `AConfirm`

除非是组件内部实现（如特殊时间选择器内部结构），不要直接在页面里重新发明一套按钮、输入框、select、dropdown、modal。

#### 2. 按钮用法

- `ABtn` 统一使用：`variant="primary|secondary|danger|ghost"`
- 尺寸统一使用：`size="sm|md|lg"`
- 需要整行按钮时使用：`block`
- 历史代码中的 `outline` / `danger` 写法可兼容，但新代码优先写 `variant`

#### 3. 表单控件用法

- 普通输入框优先用 `AInput`
- 多行文本优先用 `ATextarea`
- 页面级原生 `<select>` 优先替换为 `ASelect`
- 表单 label 固定放在控件上方，不用 placeholder 代替 label
- 错误提示与 hint 通过组件 props 或统一样式呈现，不随意单独发明样式

#### 4. 弹窗用法

- 弹窗优先使用 `AModal`
- 不要在页面中重复手写 `.a-modal-header / .a-modal-body / .a-modal-footer` 结构，优先使用 `AModal` 的 `title` 和 `#footer` 插槽
- 删除确认优先使用 `AConfirm`

#### 5. 颜色与样式来源

- 优先使用 `web/src/style.css` 中的 `--a-*` design tokens
- 不要在页面里继续大量硬编码 `#000 / #fff / #6b7280 / #9ca3af / #f3f4f6` 等颜色
- 允许保留少量“内容语义色”，例如：
  - 代码高亮主题色
  - warning / alert / 审核状态 / 结论状态色
- 但“UI 结构色”必须优先走 tokens

#### 6. 布局样式规则

- 新代码优先抽成 scoped class，而不是在模板里大量堆 `style="display:flex;..."`
- 页面级高频布局（flex、gap、padding、margin、width、position）应逐步从 inline style 收口到 `<style scoped>`
- 允许在过渡期保留少量 inline layout style，但新增代码不应继续扩大这种写法

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
