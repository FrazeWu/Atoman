# DevOps 改造进度日志

## 2026-05-12
- 完成了本地 PostgreSQL + MinIO 基础设施准备。
- 完成了上传逻辑从“按角色决定本地/S3”向“按 `STORAGE_TYPE` 决定存储后端”的切换。
- 完成了 SQLite 数据与本地 `uploads/` 文件迁移到 PostgreSQL + MinIO。
- 修复了迁移脚本中的关键问题：SQLite 路径错误、MinIO bucket 未创建、SQLite/PG 表名大小写差异、布尔值类型转换、关联表字段过滤。
- 验证迁移后的 PostgreSQL 数据记录数量与本地文件上传数量均符合预期。
- 修复了 MinIO bucket 匿名访问权限，确认迁移后的对象可通过 `http://127.0.0.1:9100/atoman-assets/...` 正常访问。
- 修复了后端启动配置优先级：`server/cmd/start_server/main.go` 现在优先读取 `.env.dev`，避免运行中的服务误连 `server/.env` 中的 SQLite / local storage。
- 使用测试账号 `fazong / Fraze@700` 完成实际联调验证：
  - 登录成功，前端正常跳转到 `/feed`
  - RSS 时间线可正常加载
  - 音乐时间线页面可直接显示 MinIO 封面
  - 专辑详情页可触发播放，播放器显示的封面来自 MinIO，对应音频链路已生效
  - 博客列表页与文章详情页均可正常读取 PostgreSQL 中迁移后的数据
- 阶段 4 与阶段 5 已全部完成。