# Atoman

自托管发布与讨论平台：支持 Studio（博客/视频/播客）、RSS、论坛、辩题、音乐档案库。

## Tech Stack

| Layer | Stack |
|---|---|
| Frontend | Vue 3.5, Vite, TypeScript 5.9 (strict) |
| State / Router | Pinia 3, Vue Router 4 |
| Styling | Tailwind CSS v4 |
| Backend | Go, Gin, GORM, JWT |
| Database | PostgreSQL (prod) / SQLite (dev) |
| Storage | S3-compatible, MinIO (dev) |
| Infra | Docker Compose, Nginx, supervisord |

## Commands

```bash
# frontend (web/)
bun install && bun run dev      # 启动开发服务器
bun run build                   # 构建产物
bun run type-check && bun run lint

# backend (server/)
go build ./...
go run cmd/start_server/main.go
go run cmd/create_admin/main.go
```

## Directory Architecture

### Top-level

| Path | 职责 |
|---|---|
| `web/` | Vue 3 前端（SPA） |
| `server/` | Go 后端（HTTP API + 业务逻辑） |
| `docs/superpowers/specs/` | 持久需求、架构决策、产品规约（唯一权威规划源） |
| `docs/superpowers/plans/` | 可执行实施步骤、验证命令、分步计划 |
| `.claude/rules/` | 按路径按需注入的 AI 规则（backend.md / frontend-ui.md 等） |
| `plan/` | 旧草稿，**不作为权威规划，勿更新** |
| `doc/` | 遗留补充文档，**勿在此新建规划文档** |
| `nginx/` | Nginx 配置与 SSL 证书 |

### Frontend: `web/src/`

| Path | 职责 |
|---|---|
| `views/` | 页面组件（按功能模块分子目录） |
| `views/blog/` | 博客频道、文章详情、编辑器页 |
| `views/forum/` | 论坛帖子列表、帖子详情 |
| `views/debate/` | 辩题页面 |
| `views/music/` | 专辑、歌曲、歌词注释页 |
| `views/timeline/` | 时间线页面 |
| `views/orbit/` | Orbit 社交页面 |
| `views/feed/` | RSS 订阅页面 |
| `views/auth/` | 登录、注册、邮箱验证 |
| `components/` | 可复用 UI 组件 |
| `components/blog/` | 博客专用组件 |
| `components/forum/` | 论坛专用组件 |
| `components/debate/` | 辩题专用组件 |
| `components/shared/` | 跨模块共享组件 |
| `components/ui/` | 基础 UI 原子组件 |
| `stores/` | Pinia 状态（auth / debate / feed / forum / player / timeline） |
| `composables/` | Vue 组合式函数 |
| `router.ts` | 路由定义（所有页面路由在此） |
| `types.ts` | 全局 TypeScript 类型定义 |

### Backend: `server/`

| Path | 职责 |
|---|---|
| `cmd/start_server/` | 服务器入口（main.go） |
| `cmd/create_admin/` | 创建管理员账号 CLI |
| `cmd/migrate_revision_system/` | 修订系统数据库迁移 CLI |
| `internal/handlers/` | HTTP 请求处理器（每个模块一个文件） |
| `internal/handlers/auth_handler.go` | 注册/登录/JWT |
| `internal/handlers/blog_*.go` | 博客频道、文章、互动、上传 |
| `internal/handlers/forum_handler.go` | 论坛帖子与回复 |
| `internal/handlers/debate_handler.go` | 辩题 CRUD |
| `internal/handlers/songs_handler.go` | 歌曲管理 |
| `internal/handlers/albums_handler.go` | 专辑管理 |
| `internal/handlers/timeline_handler.go` | 时间线事件 |
| `internal/handlers/feed_handler.go` | RSS 订阅 |
| `internal/handlers/admin_handler.go` | 管理后台操作 |
| `internal/handlers/revision_handler.go` | 内容修订/版本系统 |
| `internal/model/` | GORM 数据模型（按模块分文件） |
| `internal/migrations/` | 数据库迁移文件 |
| `internal/middleware/` | JWT 鉴权、CORS 等中间件 |
| `internal/collab/` | WebSocket 协作（Hub） |
| `internal/service/` | 业务服务层 |
| `internal/storage/` | 文件存储接口（S3/MinIO） |

## Planning Rules

- 规划文档只写在 `docs/superpowers/specs/`（需求/决策）或 `docs/superpowers/plans/`（执行步骤）。
- 架构或产品行为变动 → 更新对应 spec；执行顺序或验证步骤变动 → 更新对应 plan。

## Action Principles

- **强制规划**：3 个以上步骤的任务，先输出计划并等待确认，再动代码。
- **先读后写**：修改任何文件前，必须先阅读现有实现，匹配现有模式。
- **强制验证**：前端改动跑 type-check；后端改动跑 `go build ./...`；完成前必须证明代码无误。
- **及时止损**：需求模糊、多次尝试不收敛时，主动停止并请求确认。
- **最小改动**：不添加推测性抽象、不清理无关代码、不加与任务无关的错误处理。


