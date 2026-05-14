# DevOps Spec

## Purpose
DevOps 模块的目标是建立一致的本地开发和生产部署环境，消除数据库驱动与存储后端的环境差异。本地开发使用 Docker 管理 PostgreSQL 与 MinIO 等基础设施，同时保持前后端业务进程原生运行以便调试。

## Core Environment Model

### Local development model
- 基础设施（PostgreSQL、MinIO）容器化运行。
- Go 后端与 Vue 前端在宿主机原生运行。
- 开发环境必须尽量贴近生产，避免继续依赖 SQLite + 本地文件系统的历史默认模式。

### Storage and database rules
- 数据库通过环境变量切换，但默认开发路径应指向 PostgreSQL。
- 存储上传逻辑由 `STORAGE_TYPE` 控制。
- 在 `STORAGE_TYPE=s3` 时，本地开发也直接写入 MinIO，而不是退回本地 `uploads/`。
- `local` 作为回退模式存在，但不再与管理员角色绑定。

## Migration Rules
- 既有 SQLite 数据应迁移到 PostgreSQL。
- 本地 `uploads/` 文件应迁移到 MinIO bucket。
- 迁移过程中要处理 SQLite / PostgreSQL 差异，如表结构、布尔值、关联字段与特殊方言代码。
- 旧资源路径应重写为对应 S3/MinIO URL。

## Product and Ops Goals
- 本地与生产环境尽量一致。
- 基础设施切换后，业务功能（登录、音乐播放、博客展示、RSS 时间线）应保持可用。
- 生产配置不得依赖开发态 `.env` 回退逻辑。
- 生产部署推荐整理为更稳定的一体化多阶段构建流程。

## Durable Decisions
- 采用“基础设施容器化 + 业务进程原生运行”的混合开发模式。
- PostgreSQL 与 MinIO 是推荐的本地开发标准基础设施。
- 上传逻辑由环境配置驱动，而不是由用户角色决定。
- MinIO 若需前端直连访问对象，必须提供匿名读策略或改为签名 URL / 代理下载。
- 后端启动期的环境变量优先级是关键风险，必须避免误连 SQLite 或回退本地存储。

## Key Operational Risks
- 特定数据库方言代码可能破坏环境一致性。
- 开发期 `.env` / `.env.dev` 并存时容易出现错误配置优先级。
- 只创建 bucket 不足以支持前端对象访问，还需要访问策略。
- 若继续保留本地上传路径作为主依赖，环境一致性目标就会失效。

## Boundaries
- 该模块关注环境一致性、数据迁移、构建与部署模板，不直接替代业务实现文档。
- 生产镜像结构可以后续继续优化，但环境一致性规则与配置约束属于长期保留知识。
