# DevOps与环境一致性改造任务计划

## 目标
建立一致的本地开发和生产部署环境，消除数据库驱动和存储后端的环境差异。使用 Docker 管理本地基础设施依赖（PostgreSQL、MinIO），同时保持业务代码原生运行以方便断点调试。

## 现状
- 后端支持通过环境变量切换数据库（SQLite/PostgreSQL）和存储（Local/S3）。
- 存在特定于数据库方言的代码（如 `forum_migrate.go` 中的 `ltree` 判断）。
- 本地历史上默认使用 SQLite 和本地文件系统，导致开发和生产环境不一致。

## 决策记录
1. **采用方案 A**：本地开发环境使用 Docker 运行 PostgreSQL 和 MinIO，尽量消除与生产环境的差异。
2. **混合开发模式**：基础设施（数据库、对象存储）容器化运行；前后端应用（Go、Vue）继续使用宿主机原生方式（`go run`、`bun run dev`）启动，以保持最佳调试体验。
3. **统一存储路径**：上传逻辑优先由 `STORAGE_TYPE` 控制，本地开发在 `s3` 模式下也直接写入 MinIO，而不是继续分流到本地 `uploads/`。
4. **数据延续性**：需要把既有的 SQLite 数据与本地 `uploads/` 文件迁移到新的 PostgreSQL + MinIO 环境，保证切换后数据可继续使用。

## 阶段规划

### 阶段 1：构建本地基础设施配置 (完成)
- [x] 重构 `docker-compose.dev.yml`，仅承载本地基础设施服务
- [x] 配置 PostgreSQL 服务、数据卷与数据库初始化流程
- [x] 配置 MinIO 服务与 bucket 初始化流程
- [x] 创建 `.env.dev.example`，统一本地 PG / MinIO 连接参数

### 阶段 2：调整项目配置与清理差异代码 (完成)
- [x] 检查并梳理上传逻辑中依赖角色决定本地/S3存储的旧逻辑
- [x] 修改 `songs_handler.go`，使音频与封面在 `STORAGE_TYPE=s3` 时统一写入 S3/MinIO
- [x] 修改 `albums_handler.go`，使封面上传逻辑与环境配置保持一致
- [x] 保留 `local` 作为回退模式，但不再把“是否上传 S3”绑定到管理员角色

### 阶段 3：编写本地数据迁移脚本 (完成)
- [x] 编写 `migrate_db.go` 迁移工具
- [x] 自动扫描本地 `uploads/` 并上传到 MinIO bucket `atoman-assets`
- [x] 将 `dev.sqlite` 数据迁移到 PostgreSQL
- [x] 迁移过程中把 `/uploads/...` URL 重写为对应的 S3 URL
- [x] 处理 SQLite / PostgreSQL 之间的表名、布尔值、关联表字段差异

### 阶段 4：迁移结果验证与联调 (完成)
- [x] 启动后端并确认新环境可正常连接 PostgreSQL 和 MinIO
- [x] 抽样验证迁移后的 MinIO 对象可访问（已修复 bucket 公共读权限）
- [x] 验证关键功能：登录、音乐播放、博客内容展示、RSS 时间线读取
- [x] 确认旧的本地资源路径已不再作为主路径依赖
- [x] 修正后端启动优先读取 `.env` 导致仍连接 SQLite / local storage 的问题

### 阶段 5：生产部署模板整理 (完成)
- [x] 根据 ARM Mac 本地与 ARM Ubuntu 远程环境，梳理镜像与部署建议
- [x] 整理生产 `docker-compose` / Dockerfile 的推荐结构
- [x] 明确开发环境与生产环境的环境变量管理方式

## 当前进度
阶段 1、阶段 2、阶段 3、阶段 4、阶段 5 均已完成。本地开发环境现已完成从 SQLite + local uploads 到 PostgreSQL + MinIO 的切换，并通过实际前端联调验证。