# DevOps 改造研究发现

## 2026-05-12
- 本地开发推荐采用“基础设施容器化 + 业务进程原生运行”的混合模式：PostgreSQL 与 MinIO 使用 Docker，Go 后端和 Vue 前端继续在宿主机运行，便于断点调试与热更新。
- ARM Mac 本地与 ARM Ubuntu 远端在镜像架构上可以统一采用官方多架构镜像，如 `postgres:16-alpine`、`minio/minio`，避免额外维护 amd64-only 镜像。
- 现有 `docker-compose.prod.yml` 依赖预先构建好的前端产物和后端二进制，并通过 Nginx + supervisord 组合运行；这能工作，但不够简洁，后续更推荐使用多阶段 Docker 构建，把前端构建、后端编译、运行时镜像整理为更稳定的一体化生产镜像流程。
- 生产环境的配置管理应避免依赖开发态 `.env` 回退逻辑。开发期允许 `.env` / `.env.dev` 并存，但生产部署应显式指定唯一配置源，避免服务误连 SQLite 或误切回本地存储。
- MinIO 仅创建 bucket 不足以支持前端直接访问对象；如果前端要直接展示封面、音频、图片，需要额外设置 bucket 的匿名读取策略或改为后端签名 URL / 代理下载。
- 本次阶段 4 验证中发现：后端启动逻辑优先加载 `server/.env`，导致即使 PostgreSQL 与 MinIO 已迁移完毕，运行中的服务仍可能继续连接 SQLite 和本地 `uploads/`。这是环境一致性改造里最关键的运行时配置风险。